// Paquete resultados: consume los eventos y mantiene un modelo de lectura para el panel.
// Tiene dos caras en un mismo proceso:
//   - consumidor: escucha eventos de RabbitMQ y actualiza sus propias tablas;
//   - API HTTP de solo lectura: responde las estadísticas que muestra el panel.
//
// Separar "quien escribe" de "quien lee" se conoce como CQRS.
package resultados

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5"

	"bulkphish/internal/basedatos"
	"bulkphish/internal/config"
	"bulkphish/internal/eventos"
	"bulkphish/internal/mensajeria"
	"bulkphish/internal/web"
)

const colaResultados = "resultados.eventos"

type servidor struct {
	repo *repositorio
}

// Ejecutar arranca consumidor y API; si uno falla, se apaga el otro.
func Ejecutar() error {
	config.IniciarRegistro("resultados")
	ctx, parar := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer parar()

	bd, err := basedatos.Conectar(ctx, config.Obligatorio("BD_URL"))
	if err != nil {
		return err
	}
	defer bd.Close()
	mq, err := mensajeria.Conectar(ctx, config.Obligatorio("AMQP_URL"))
	if err != nil {
		return err
	}
	defer mq.Cerrar()

	repo := &repositorio{bd: bd}
	s := &servidor{repo: repo}

	ctxInterno, cancelar := context.WithCancel(ctx)
	defer cancelar()
	errores := make(chan error, 2)

	go func() {
		errores <- mq.Consumir(ctxInterno, colaResultados,
			[]string{eventos.CampanaCreada, eventos.CorreoEnviado, eventos.EnlaceClicado, eventos.CorreoReportado},
			repo.manejarEvento)
	}()
	go func() {
		errores <- web.Servir(ctxInterno, config.Texto("PUERTO", "8083"), s.rutas())
	}()

	err = <-errores // el primero que termine determina el resultado
	cancelar()
	<-errores
	return err
}

// manejarEvento decide qué hacer según el tipo de evento recibido.
func (r *repositorio) manejarEvento(ctx context.Context, sobre eventos.Sobre) error {
	switch sobre.Tipo {
	case eventos.CampanaCreada:
		var d eventos.DatosCampanaCreada
		if err := sobre.Decodificar(&d); err != nil {
			return err
		}
		return r.registrarCampana(ctx, d)
	case eventos.CorreoEnviado:
		var d eventos.DatosCorreoEnviado
		if err := sobre.Decodificar(&d); err != nil {
			return err
		}
		return r.marcar(ctx, "enviado_en", d.DestinatarioID, d.EnviadoEn)
	case eventos.EnlaceClicado:
		var d eventos.DatosInteraccion
		if err := sobre.Decodificar(&d); err != nil {
			return err
		}
		return r.marcar(ctx, "clic_en", d.DestinatarioID, d.OcurridoEn)
	case eventos.CorreoReportado:
		var d eventos.DatosInteraccion
		if err := sobre.Decodificar(&d); err != nil {
			return err
		}
		return r.marcar(ctx, "reportado_en", d.DestinatarioID, d.OcurridoEn)
	default:
		return fmt.Errorf("tipo de evento desconocido: %s", sobre.Tipo)
	}
}

func (s *servidor) rutas() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /salud", web.Salud)
	mux.HandleFunc("GET /api/resultados/resumen", s.resumen)
	mux.HandleFunc("GET /api/resultados/campanas", s.listarCampanas)
	mux.HandleFunc("GET /api/resultados/campanas/{id}", s.detalleCampana)
	return mux
}

func errorInterno(w http.ResponseWriter, contexto string, err error) {
	slog.Error(contexto, "error", err)
	web.Error(w, http.StatusInternalServerError, "Ocurrió un error inesperado. Inténtalo de nuevo.")
}

// resumen: indicadores globales del panel principal.
func (s *servidor) resumen(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	campanas, err := s.repo.campanas(ctx, 1000)
	if err != nil {
		errorInterno(w, "resumen: campañas", err)
		return
	}
	var total, enviados, clics, reportados int
	for _, c := range campanas {
		total += c.Total
		enviados += c.Enviados
		clics += c.Clics
		reportados += c.Reportados
	}
	porArea, err := s.repo.porArea(ctx, "")
	if err != nil {
		errorInterno(w, "resumen: por área", err)
		return
	}
	porPlantilla, err := s.repo.porPlantilla(ctx)
	if err != nil {
		errorInterno(w, "resumen: por plantilla", err)
		return
	}
	recientes := campanas
	if len(recientes) > 5 {
		recientes = recientes[:5]
	}
	web.JSON(w, http.StatusOK, map[string]any{
		"total_campanas": len(campanas),
		"global":         calcular(total, enviados, clics, reportados),
		"por_area":       porArea,
		"por_plantilla":  porPlantilla,
		"recientes":      recientes,
	})
}

func (s *servidor) listarCampanas(w http.ResponseWriter, r *http.Request) {
	lista, err := s.repo.campanas(r.Context(), 200)
	if err != nil {
		errorInterno(w, "listar campañas", err)
		return
	}
	web.JSON(w, http.StatusOK, lista)
}

// detalleCampana responde 404 mientras el evento "campana.creada" aún no fue procesado
// (consistencia eventual); el panel lo interpreta como "preparando" y vuelve a consultar.
func (s *servidor) detalleCampana(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !web.EsUUID(id) {
		web.Error(w, http.StatusNotFound, "No encontramos esa campaña.")
		return
	}
	ctx := r.Context()
	c, err := s.repo.campana(ctx, id)
	if err != nil {
		errorInterno(w, "detalle: campaña", err)
		return
	}
	if c == nil {
		web.Error(w, http.StatusNotFound, "Los resultados de esta campaña todavía se están preparando.")
		return
	}
	porArea, err := s.repo.porArea(ctx, id)
	if err != nil {
		errorInterno(w, "detalle: por área", err)
		return
	}
	dest, err := s.repo.destinatarios(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			dest = nil
		} else {
			errorInterno(w, "detalle: destinatarios", err)
			return
		}
	}
	web.JSON(w, http.StatusOK, map[string]any{
		"campana": c, "por_area": porArea, "destinatarios": dest,
	})
}
