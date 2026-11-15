package main

import (
	"log/slog"
	"os"

	"bulkphish/internal/resultados"
)

func main() {
	if err := resultados.Ejecutar(); err != nil {
		slog.Error("el servicio terminó con error", "error", err)
		os.Exit(1)
	}
}
