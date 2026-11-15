package main

import (
	"log/slog"
	"os"

	"bulkphish/internal/autenticacion"
)

func main() {
	if err := autenticacion.Ejecutar(); err != nil {
		slog.Error("el servicio terminó con error", "error", err)
		os.Exit(1)
	}
}
