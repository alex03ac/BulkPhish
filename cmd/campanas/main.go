package main

import (
	"log/slog"
	"os"

	"bulkphish/internal/campanas"
)

func main() {
	if err := campanas.Ejecutar(); err != nil {
		slog.Error("el servicio terminó con error", "error", err)
		os.Exit(1)
	}
}
