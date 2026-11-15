package main

import (
	"log/slog"
	"os"

	"bulkphish/internal/correos"
)

func main() {
	if err := correos.Ejecutar(); err != nil {
		slog.Error("el servicio terminó con error", "error", err)
		os.Exit(1)
	}
}
