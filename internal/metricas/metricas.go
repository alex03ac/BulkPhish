// Paquete metricas: colectores de Prometheus compartidos por todos los servicios.
// Cada servicio expone /metrics (vía web.Servir) y Prometheus los recolecta.
package metricas

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// PeticionesHTTP cuenta las peticiones HTTP atendidas, por método, ruta y código.
	PeticionesHTTP = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "bulkphish_http_peticiones_total",
		Help: "Peticiones HTTP atendidas, por método, ruta y código de estado.",
	}, []string{"metodo", "ruta", "estado"})

	// DuracionHTTP mide la duración de las peticiones HTTP en segundos.
	DuracionHTTP = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "bulkphish_http_duracion_segundos",
		Help:    "Duración de las peticiones HTTP en segundos.",
		Buckets: prometheus.DefBuckets,
	}, []string{"metodo", "ruta"})

	// EventosPublicados cuenta los eventos publicados a RabbitMQ, por tipo.
	EventosPublicados = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "bulkphish_eventos_publicados_total",
		Help: "Eventos publicados a RabbitMQ, por tipo.",
	}, []string{"tipo"})

	// EventosProcesados cuenta los eventos consumidos, por cola, tipo y resultado (ok|muerto).
	EventosProcesados = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "bulkphish_eventos_procesados_total",
		Help: "Eventos consumidos de RabbitMQ, por cola, tipo y resultado.",
	}, []string{"cola", "tipo", "resultado"})

	// CorreosEnviados cuenta los correos simulados entregados al buzón de pruebas.
	CorreosEnviados = promauto.NewCounter(prometheus.CounterOpts{
		Name: "bulkphish_correos_enviados_total",
		Help: "Correos simulados entregados por SMTP a Mailpit.",
	})
)

// Manejador devuelve el handler que expone las métricas en formato Prometheus.
func Manejador() http.Handler {
	return promhttp.Handler()
}
