// Paquete basedatos: conexión a PostgreSQL con reintentos al arrancar.
package basedatos

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Conectar abre un pool de conexiones y espera hasta 60 s a que PostgreSQL responda.
func Conectar(ctx context.Context, url string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("url de base de datos inválida: %w", err)
	}
	cfg.MaxConns = 10

	var ultimoError error
	for intento := 1; intento <= 30; intento++ {
		pool, err := pgxpool.NewWithConfig(ctx, cfg)
		if err == nil {
			if err = pool.Ping(ctx); err == nil {
				return pool, nil
			}
			pool.Close()
		}
		ultimoError = err
		slog.Warn("PostgreSQL aún no responde, reintentando", "intento", intento, "error", err)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
	return nil, fmt.Errorf("no se pudo conectar a PostgreSQL: %w", ultimoError)
}
