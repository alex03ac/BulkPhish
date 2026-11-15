package campanas

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"bulkphish/internal/eventos"
	"bulkphish/internal/mensajeria"
)

// relevo lee la bandeja de salida y publica en RabbitMQ lo que todavía no se envió.
//
// ¿Por qué existe? Guardar en la base de datos y publicar en RabbitMQ son dos
// operaciones distintas: si el servicio se cae entre ambas, se perdería el evento.
// Con este patrón (transactional outbox) el evento queda guardado junto con los datos
// y se publica tarde o temprano, aunque haya caídas. Garantía: "al menos una vez".
type relevo struct {
	repo *repositorio
	mq   *mensajeria.Cliente
}

func (v *relevo) ejecutar(ctx context.Context) {
	reloj := time.NewTicker(time.Second)
	defer reloj.Stop()
	for {
		for {
			n, err := v.publicarLote(ctx)
			if err != nil {
				slog.Error("relevo: fallo al publicar", "error", err)
				break
			}
			if n == 0 {
				break
			}
		}
		select {
		case <-ctx.Done():
			return
		case <-reloj.C:
		}
	}
}

// publicarLote toma hasta 50 eventos pendientes (FOR UPDATE SKIP LOCKED permite
// varias réplicas del servicio sin que publiquen lo mismo) y los marca como publicados.
func (v *relevo) publicarLote(ctx context.Context) (int, error) {
	tx, err := v.repo.bd.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	filas, err := tx.Query(ctx, `
		SELECT id::text, clave_ruta, sobre::text FROM bandeja_salida
		WHERE publicado_en IS NULL ORDER BY creado_en LIMIT 50 FOR UPDATE SKIP LOCKED`)
	if err != nil {
		return 0, err
	}
	type pendiente struct{ id, clave, sobre string }
	lote := []pendiente{}
	for filas.Next() {
		var p pendiente
		if err := filas.Scan(&p.id, &p.clave, &p.sobre); err != nil {
			filas.Close()
			return 0, err
		}
		lote = append(lote, p)
	}
	filas.Close()

	publicados := 0
	for _, p := range lote {
		var sobre eventos.Sobre
		if err := json.Unmarshal([]byte(p.sobre), &sobre); err != nil {
			return publicados, err
		}
		if err := v.mq.Publicar(ctx, p.clave, sobre); err != nil {
			break // se confirmará lo ya publicado y se reintentará el resto
		}
		if _, err := tx.Exec(ctx, `UPDATE bandeja_salida SET publicado_en = now() WHERE id = $1`, p.id); err != nil {
			return publicados, err
		}
		publicados++
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	if publicados > 0 {
		slog.Info("relevo: eventos publicados", "cantidad", publicados)
	}
	if publicados < len(lote) {
		return publicados, errors.New("RabbitMQ no confirmó todos los eventos; se reintentará") // fuerza una pausa
	}
	return publicados, nil
}
