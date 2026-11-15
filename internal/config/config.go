// Paquete config: lectura de variables de entorno y configuración del registro (logs).
package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
)

// Texto devuelve la variable de entorno o el valor por defecto si no existe.
func Texto(clave, defecto string) string {
	if valor := os.Getenv(clave); valor != "" {
		return valor
	}
	return defecto
}

// Entero devuelve la variable de entorno como número o el valor por defecto.
func Entero(clave string, defecto int) int {
	if valor := os.Getenv(clave); valor != "" {
		if n, err := strconv.Atoi(valor); err == nil {
			return n
		}
	}
	return defecto
}

// Obligatorio termina el proceso si la variable no está definida.
func Obligatorio(clave string) string {
	valor := os.Getenv(clave)
	if valor == "" {
		fmt.Fprintf(os.Stderr, "falta la variable de entorno obligatoria: %s\n", clave)
		os.Exit(1)
	}
	return valor
}

// IniciarRegistro configura logs estructurados (JSON) hacia la salida estándar.
// Docker captura la salida estándar, así que no se escribe en archivos.
func IniciarRegistro(servicio string) *slog.Logger {
	manejador := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	registro := slog.New(manejador).With("servicio", servicio)
	slog.SetDefault(registro)
	return registro
}
