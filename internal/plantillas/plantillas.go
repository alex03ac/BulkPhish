// Paquete plantillas: catálogo de correos simulados, uno por tipo de ataque.
// Lo usan dos servicios: "correos" (para redactar el mensaje) y "campanas"
// (para mostrar la capacitación cuando alguien hace clic).
package plantillas

import (
	"bytes"
	"html/template"
)

// Plantilla describe un tipo de ataque simulado y lo que se enseña al respecto.
type Plantilla struct {
	ID              string
	Nombre          string
	Tipo            string // familia del ataque: credenciales, financiero, logística, ingeniería social
	Descripcion     string // texto para el administrador
	RemitenteNombre string
	RemitenteCorreo string
	Asunto          string
	cuerpo          string   // HTML del correo (usa html/template)
	Senales         []string // señales de alerta que se muestran en la capacitación
	Consejo         string
}

// DatosCorreo son los valores que se insertan en el correo de cada destinatario.
type DatosCorreo struct {
	Nombre        string
	EnlaceTrampa  string // enlace que registra el clic
	EnlaceReporte string // enlace del botón "Reportar phishing"
}

func armar(contenido string) string {
	return `<!doctype html><html lang="es"><body style="margin:0;background:#eef1f4;font-family:Arial,Helvetica,sans-serif;color:#1b2733">
<div style="max-width:620px;margin:0 auto;padding:20px">
<div style="background:#e8f1f8;border:1px solid #c5dbec;border-radius:6px;padding:8px 12px;font-size:12px;color:#34536b;margin-bottom:12px">
Barra de tu cliente de correo &middot; <a href="{{.EnlaceReporte}}" style="color:#0b5a8a;font-weight:bold">Reportar phishing</a>
</div>
<div style="background:#fff;border-radius:8px;padding:28px;border:1px solid #dde3e8;line-height:1.55;font-size:15px">` + contenido + `</div>
<p style="font-size:11px;color:#7b8794;text-align:center;margin-top:14px">Mensaje generado como parte de un simulacro interno de seguridad.</p>
</div></body></html>`
}

func boton(texto string) string {
	return `<p style="margin:24px 0"><a href="{{.EnlaceTrampa}}" style="background:#0b5a8a;color:#fff;text-decoration:none;padding:12px 22px;border-radius:6px;font-weight:bold;display:inline-block">` + texto + `</a></p>`
}

var catalogo = []Plantilla{
	{
		ID: "restablecer-clave", Nombre: "Restablecimiento de contraseña", Tipo: "Robo de credenciales",
		Descripcion:     "Suplanta al área de TI y exige renovar la contraseña antes de que expire.",
		RemitenteNombre: "Soporte de TI", RemitenteCorreo: "soporte-ti@mesa-de-ayuda-corp.test",
		Asunto: "Acción requerida: tu contraseña vence hoy",
		cuerpo: armar(`<p>Hola {{.Nombre}},</p>
<p>Nuestro sistema detectó que tu contraseña corporativa <strong>vence hoy a las 18:00</strong>. Si no la renuevas, tu cuenta quedará bloqueada y perderás acceso al correo y a los sistemas internos.</p>` +
			boton("Renovar contraseña ahora") + `<p style="color:#52606d;font-size:13px">Equipo de Soporte de TI</p>`),
		Senales: []string{
			"Crea urgencia con una hora límite y amenaza con bloquear tu cuenta.",
			"El dominio del remitente no pertenece a tu empresa (mesa-de-ayuda-corp.test).",
			"Te pide iniciar sesión desde un enlace del correo en lugar de usar el portal oficial.",
		},
		Consejo: "TI nunca te pedirá renovar la contraseña desde un enlace de correo. Entra por el portal oficial o llama a la mesa de ayuda.",
	},
	{
		ID: "factura-pendiente", Nombre: "Factura pendiente de pago", Tipo: "Fraude financiero",
		Descripcion:     "Simula a un proveedor que reclama el pago de una factura vencida.",
		RemitenteNombre: "Cobranzas - Proveedor", RemitenteCorreo: "cobranzas@facturacion-proveedores.test",
		Asunto: "Factura N.º 48213 vencida: evita el recargo por mora",
		cuerpo: armar(`<p>Estimado/a {{.Nombre}},</p>
<p>No hemos recibido el pago de la factura <strong>N.º 48213</strong> por <strong>S/ 12,450.00</strong>, vencida hace 5 días. Si no se regulariza hoy se aplicará un recargo por mora y se suspenderá el servicio.</p>` +
			boton("Ver factura y detalle de pago") + `<p style="color:#52606d;font-size:13px">Departamento de Cobranzas</p>`),
		Senales: []string{
			"Menciona un monto y un recargo para presionarte a actuar sin verificar.",
			"No indica quién eres como cliente ni a qué contrato corresponde la factura.",
			"El enlace descarga o abre un sitio externo en vez de dirigirte a un portal conocido.",
		},
		Consejo: "Verifica toda factura con el área de Finanzas antes de abrir enlaces o realizar pagos, aunque parezca de un proveedor habitual.",
	},
	{
		ID: "entrega-paquete", Nombre: "Paquete retenido en aduana", Tipo: "Ingeniería social",
		Descripcion:     "Avisa de un envío retenido que requiere un pago o datos adicionales.",
		RemitenteNombre: "Seguimiento de Envíos", RemitenteCorreo: "avisos@seguimiento-envios-express.test",
		Asunto: "Tu paquete no pudo ser entregado",
		cuerpo: armar(`<p>Hola {{.Nombre}},</p>
<p>Intentamos entregar tu paquete pero la dirección no pudo confirmarse. El envío se devolverá al remitente en <strong>24 horas</strong> si no actualizas tus datos de entrega.</p>` +
			boton("Reprogramar entrega") + `<p style="color:#52606d;font-size:13px">Servicio de Seguimiento de Envíos</p>`),
		Senales: []string{
			"No esperabas ningún paquete y el mensaje no incluye un número de guía verificable.",
			"Impone un plazo corto para que actúes sin pensar.",
			"El remitente usa un dominio genérico que no corresponde a una empresa de mensajería real.",
		},
		Consejo: "Si no esperas un envío, ignora el mensaje. Si lo esperas, consulta el seguimiento directamente en la web de la empresa de mensajería.",
	},
	{
		ID: "beneficio-rrhh", Nombre: "Beneficio exclusivo de Recursos Humanos", Tipo: "Ingeniería social",
		Descripcion:     "Ofrece un bono o beneficio con cupos limitados para atraer el clic por interés.",
		RemitenteNombre: "Recursos Humanos", RemitenteCorreo: "beneficios@rrhh-corporativo-bonos.test",
		Asunto: "Bono de reconocimiento: confirma tus datos antes del viernes",
		cuerpo: armar(`<p>Hola {{.Nombre}},</p>
<p>Por tu desempeño este año fuiste seleccionado/a para recibir un <strong>bono de reconocimiento</strong>. Los cupos son limitados y se asignan por orden de confirmación.</p>` +
			boton("Confirmar mi bono") + `<p style="color:#52606d;font-size:13px">Área de Recursos Humanos</p>`),
		Senales: []string{
			"Promete un beneficio inesperado y limita los cupos para apurarte.",
			"Pide confirmar datos personales o bancarios a través de un enlace.",
			"RR. HH. comunicaría un bono por los canales oficiales y no desde un dominio ajeno.",
		},
		Consejo: "Ante un beneficio inesperado, consulta directamente con Recursos Humanos por un canal que tú conozcas y no con los datos del correo.",
	},
}

// Todas devuelve el catálogo completo.
func Todas() []Plantilla {
	copia := make([]Plantilla, len(catalogo))
	copy(copia, catalogo)
	return copia
}

// Buscar devuelve una plantilla por su identificador.
func Buscar(id string) (Plantilla, bool) {
	for _, p := range catalogo {
		if p.ID == id {
			return p, true
		}
	}
	return Plantilla{}, false
}

// Renderizar genera el HTML del correo para un destinatario concreto.
// html/template escapa los datos automáticamente (evita inyección de HTML).
func (p Plantilla) Renderizar(datos DatosCorreo) (string, error) {
	t, err := template.New(p.ID).Parse(p.cuerpo)
	if err != nil {
		return "", err
	}
	var salida bytes.Buffer
	if err := t.Execute(&salida, datos); err != nil {
		return "", err
	}
	return salida.String(), nil
}
