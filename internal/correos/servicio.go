// Paquete correos: consumidor que envía los correos simulados.
// Escucha "campana.creada", redacta un correo por destinatario con la plantilla elegida,
// lo entrega por SMTP al buzón de pruebas (Mailpit) y publica "correo.enviado".
// No conoce a campanas-api ni a la base de datos: solo reacciona a eventos.
package correos

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"mime"
	"net/smtp"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"bulkphish/internal/config"
	"bulkphish/internal/eventos"
	"bulkphish/internal/mensajeria"
	"bulkphish/internal/plantillas"
)

type servicio struct {
	mq         *mensajeria.Cliente
	direccion  string // host:puerto del servidor SMTP
	urlPublica string // base de los enlaces del correo (ej. http://localhost:8080)
}

const colaCorreos = "correos.campana_creada"

// Ejecutar arranca el consumidor y bloquea hasta recibir una señal de apagado.
func Ejecutar() error {
	config.IniciarRegistro("correos")
	ctx, parar := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer parar()

	mq, err := mensajeria.Conectar(ctx, config.Obligatorio("AMQP_URL"))
	if err != nil {
		return err
	}
	defer mq.Cerrar()

	s := &servicio{
		mq:         mq,
		direccion:  config.Texto("SMTP_HOST", "mailpit") + ":" + config.Texto("SMTP_PUERTO", "1025"),
		urlPublica: strings.TrimRight(config.Texto("URL_PUBLICA", "http://localhost:8080"), "/"),
	}
	return mq.Consumir(ctx, colaCorreos, []string{eventos.CampanaCreada}, s.manejarCampana)
}

// manejarCampana envía los correos de una campaña con hasta 5 envíos en paralelo.
func (s *servicio) manejarCampana(ctx context.Context, sobre eventos.Sobre) error {
	var datos eventos.DatosCampanaCreada
	if err := sobre.Decodificar(&datos); err != nil {
		return fmt.Errorf("evento ilegible: %w", err)
	}
	plantilla, ok := plantillas.Buscar(datos.PlantillaID)
	if !ok {
		return fmt.Errorf("la plantilla %q no existe", datos.PlantillaID)
	}
	slog.Info("enviando campaña", "campana_id", datos.CampanaID, "destinatarios", len(datos.Destinatarios))

	var (
		grupo    sync.WaitGroup
		limite   = make(chan struct{}, 5)
		mu       sync.Mutex
		enviados int
	)
	for _, d := range datos.Destinatarios {
		grupo.Add(1)
		limite <- struct{}{}
		go func(d eventos.DestinatarioCampana) {
			defer grupo.Done()
			defer func() { <-limite }()

			if err := s.enviar(plantilla, d); err != nil {
				slog.Error("no se pudo enviar el correo", "destinatario_id", d.DestinatarioID, "error", err)
				return
			}
			evento, err := eventos.NuevoSobre(eventos.CorreoEnviado, eventos.DatosCorreoEnviado{
				CampanaID: datos.CampanaID, DestinatarioID: d.DestinatarioID, EnviadoEn: time.Now().UTC(),
			})
			if err == nil {
				err = s.mq.Publicar(ctx, eventos.CorreoEnviado, evento)
			}
			if err != nil {
				slog.Error("no se pudo publicar correo.enviado", "destinatario_id", d.DestinatarioID, "error", err)
				return
			}
			mu.Lock()
			enviados++
			mu.Unlock()
		}(d)
	}
	grupo.Wait()
	slog.Info("campaña enviada", "campana_id", datos.CampanaID, "enviados", enviados, "total", len(datos.Destinatarios))
	return nil
}

// enviar redacta y entrega un correo a un destinatario.
func (s *servicio) enviar(p plantillas.Plantilla, d eventos.DestinatarioCampana) error {
	cuerpo, err := p.Renderizar(plantillas.DatosCorreo{
		Nombre:        strings.Fields(d.Nombre)[0],
		EnlaceTrampa:  s.urlPublica + "/t/" + d.Token,
		EnlaceReporte: s.urlPublica + "/r/" + d.Token,
	})
	if err != nil {
		return err
	}
	mensaje := construirMensaje(p, d.Nombre, d.Correo, cuerpo)
	return smtp.SendMail(s.direccion, nil, p.RemitenteCorreo, []string{d.Correo}, mensaje)
}

// construirMensaje arma el correo MIME: encabezados codificados y cuerpo HTML en base64.
func construirMensaje(p plantillas.Plantilla, nombreDestino, correoDestino, html string) []byte {
	codificar := func(t string) string { return mime.QEncoding.Encode("utf-8", t) }
	var b strings.Builder
	fmt.Fprintf(&b, "From: %s <%s>\r\n", codificar(p.RemitenteNombre), p.RemitenteCorreo)
	fmt.Fprintf(&b, "To: %s <%s>\r\n", codificar(nombreDestino), correoDestino)
	fmt.Fprintf(&b, "Subject: %s\r\n", codificar(p.Asunto))
	fmt.Fprintf(&b, "Date: %s\r\n", time.Now().Format(time.RFC1123Z))
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("X-BulkPhish-Simulacro: 1\r\n") // marca para que TI pueda filtrar estos correos
	b.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	b.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")
	codificado := base64.StdEncoding.EncodeToString([]byte(html))
	for len(codificado) > 76 {
		b.WriteString(codificado[:76] + "\r\n")
		codificado = codificado[76:]
	}
	b.WriteString(codificado + "\r\n")
	return []byte(b.String())
}
