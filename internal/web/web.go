// Paquete web: utilidades HTTP compartidas (respuestas JSON, registro de accesos, apagado ordenado).
package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
)

// JSON responde con un objeto en formato JSON.
func JSON(w http.ResponseWriter, estado int, valor any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store") // las respuestas de la API nunca se guardan en cachés
	w.WriteHeader(estado)
	_ = json.NewEncoder(w).Encode(valor)
}

// Error responde con {"error": "..."}; los mensajes están pensados para mostrarse al usuario.
func Error(w http.ResponseWriter, estado int, mensaje string) {
	JSON(w, estado, map[string]string{"error": mensaje})
}

// LeerJSON decodifica el cuerpo de la petición (máximo 1 MB, sin campos desconocidos).
func LeerJSON(w http.ResponseWriter, r *http.Request, destino any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decodificador := json.NewDecoder(r.Body)
	decodificador.DisallowUnknownFields()
	if err := decodificador.Decode(destino); err != nil {
		return fmt.Errorf("el cuerpo de la petición no es válido: %w", err)
	}
	return nil
}

// Salud es el endpoint que usan Docker y el gateway para saber si el servicio está vivo.
func Salud(w http.ResponseWriter, _ *http.Request) {
	JSON(w, http.StatusOK, map[string]string{"estado": "ok"})
}

// EsUUID valida el formato 8-4-4-4-12 en hexadecimal.
func EsUUID(texto string) bool {
	if len(texto) != 36 {
		return false
	}
	for i, c := range texto {
		switch i {
		case 8, 13, 18, 23:
			if c != '-' {
				return false
			}
		default:
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				return false
			}
		}
	}
	return true
}

type registroEstado struct {
	http.ResponseWriter
	estado int
}

func (r *registroEstado) WriteHeader(estado int) {
	r.estado = estado
	r.ResponseWriter.WriteHeader(estado)
}

// rutaSegura reemplaza ids y tokens por {id} para que los secretos de los enlaces
// (un token por empleado) nunca queden escritos en los registros.
func rutaSegura(ruta string) string {
	partes := strings.Split(ruta, "/")
	for i, p := range partes {
		if EsUUID(p) || (i > 0 && len(p) == 32) {
			partes[i] = "{id}"
		}
	}
	return strings.Join(partes, "/")
}

// Middleware envuelve un manejador con recuperación de pánicos y registro de cada petición.
func Middleware(siguiente http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inicio := time.Now()
		rec := &registroEstado{ResponseWriter: w, estado: http.StatusOK}
		defer func() {
			if fallo := recover(); fallo != nil {
				slog.Error("pánico en el manejador", "error", fallo, "pila", string(debug.Stack()))
				Error(rec, http.StatusInternalServerError, "error interno del servidor")
			}
			if r.URL.Path != "/salud" { // no llenar los logs con el healthcheck
				slog.Info("petición", "metodo", r.Method, "ruta", rutaSegura(r.URL.Path), "estado", rec.estado,
					"duracion_ms", time.Since(inicio).Milliseconds(), "id_peticion", r.Header.Get("X-Request-Id"))
			}
		}()
		siguiente.ServeHTTP(rec, r)
	})
}

// Servir levanta el servidor HTTP y lo apaga de forma ordenada cuando el contexto termina.
func Servir(ctx context.Context, puerto string, manejador http.Handler) error {
	servidor := &http.Server{
		Addr:              ":" + puerto,
		Handler:           Middleware(manejador),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	errores := make(chan error, 1)
	go func() {
		slog.Info("servidor HTTP escuchando", "puerto", puerto)
		errores <- servidor.ListenAndServe()
	}()
	select {
	case err := <-errores:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		cierre, cancelar := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelar()
		return servidor.Shutdown(cierre)
	}
}
