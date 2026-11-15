// Paquete mensajeria: cliente de RabbitMQ (publicar con confirmación y consumir con reintentos).
package mensajeria

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"bulkphish/internal/eventos"
)

// Manejador procesa un evento. Si devuelve error, se reintenta y luego va a la cola de muertos.
type Manejador func(ctx context.Context, sobre eventos.Sobre) error

// Cliente agrupa la conexión y el canal de publicación.
type Cliente struct {
	conn     *amqp.Connection
	canalPub *amqp.Channel
	mu       sync.Mutex // un canal AMQP no admite publicaciones concurrentes
}

// Conectar abre la conexión (con reintentos) y declara la topología base.
func Conectar(ctx context.Context, url string) (*Cliente, error) {
	var ultimoError error
	for intento := 1; intento <= 30; intento++ {
		conn, err := amqp.Dial(url)
		if err == nil {
			c := &Cliente{conn: conn}
			if err := c.declararTopologia(); err != nil {
				_ = conn.Close()
				return nil, err
			}
			return c, nil
		}
		ultimoError = err
		slog.Warn("RabbitMQ aún no responde, reintentando", "intento", intento, "error", err)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
	return nil, fmt.Errorf("no se pudo conectar a RabbitMQ: %w", ultimoError)
}

// declararTopologia crea (si no existen) los intercambios y la cola de mensajes muertos.
// Es idempotente: todos los servicios la ejecutan al arrancar.
func (c *Cliente) declararTopologia() error {
	ch, err := c.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	if err := ch.ExchangeDeclare(eventos.Intercambio, "topic", true, false, false, false, nil); err != nil {
		return err
	}
	if err := ch.ExchangeDeclare(eventos.IntercambioMuertos, "fanout", true, false, false, false, nil); err != nil {
		return err
	}
	if _, err := ch.QueueDeclare(eventos.ColaMuertos, true, false, false, false, nil); err != nil {
		return err
	}
	if err := ch.QueueBind(eventos.ColaMuertos, "", eventos.IntercambioMuertos, false, nil); err != nil {
		return err
	}

	pub, err := c.conn.Channel()
	if err != nil {
		return err
	}
	// Modo "confirmación": el broker avisa cuando guardó el mensaje.
	if err := pub.Confirm(false); err != nil {
		return err
	}
	c.canalPub = pub
	return nil
}

// Cerrar libera la conexión.
func (c *Cliente) Cerrar() { _ = c.conn.Close() }

// Publicar envía un evento y espera la confirmación del broker.
func (c *Cliente) Publicar(ctx context.Context, clave string, sobre eventos.Sobre) error {
	cuerpo, err := json.Marshal(sobre)
	if err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	confirmacion, err := c.canalPub.PublishWithDeferredConfirmWithContext(ctx, eventos.Intercambio, clave, false, false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent, // sobrevive a un reinicio de RabbitMQ
			MessageId:    sobre.ID,
			Type:         sobre.Tipo,
			Timestamp:    sobre.OcurridoEn,
			Body:         cuerpo,
		})
	if err != nil {
		return err
	}
	aceptado, err := confirmacion.WaitContext(ctx)
	if err != nil {
		return err
	}
	if !aceptado {
		return errors.New("RabbitMQ no confirmó el mensaje")
	}
	return nil
}

// Consumir crea la cola (si no existe), la enlaza a las claves indicadas y procesa
// los mensajes uno por uno. Bloquea hasta que el contexto termine o la conexión se caiga.
func (c *Cliente) Consumir(ctx context.Context, cola string, claves []string, manejador Manejador) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	argumentos := amqp.Table{"x-dead-letter-exchange": eventos.IntercambioMuertos}
	if _, err := ch.QueueDeclare(cola, true, false, false, false, argumentos); err != nil {
		return err
	}
	for _, clave := range claves {
		if err := ch.QueueBind(cola, clave, eventos.Intercambio, false, nil); err != nil {
			return err
		}
	}
	if err := ch.Qos(10, 0, false); err != nil {
		return err
	}
	entregas, err := ch.ConsumeWithContext(ctx, cola, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	slog.Info("consumidor listo", "cola", cola, "claves", claves)

	for {
		select {
		case <-ctx.Done():
			return nil
		case entrega, abierto := <-entregas:
			if !abierto {
				if ctx.Err() != nil {
					return nil
				}
				return errors.New("la conexión con RabbitMQ se cerró")
			}
			procesar(ctx, cola, entrega, manejador)
		}
	}
}

// procesar aplica hasta 3 intentos con espera creciente; si todos fallan,
// rechaza el mensaje y RabbitMQ lo desvía a la cola de mensajes muertos.
func procesar(ctx context.Context, cola string, entrega amqp.Delivery, manejador Manejador) {
	var sobre eventos.Sobre
	if err := json.Unmarshal(entrega.Body, &sobre); err != nil {
		slog.Error("mensaje ilegible, se descarta", "cola", cola, "error", err)
		_ = entrega.Reject(false)
		return
	}

	var err error
	espera := 500 * time.Millisecond
	for intento := 1; intento <= 3; intento++ {
		if err = manejador(ctx, sobre); err == nil {
			_ = entrega.Ack(false)
			return
		}
		slog.Warn("fallo al procesar evento", "cola", cola, "tipo", sobre.Tipo, "intento", intento, "error", err)
		select {
		case <-ctx.Done():
			_ = entrega.Nack(false, true) // devolver a la cola al apagar
			return
		case <-time.After(espera):
		}
		espera *= 2
	}
	slog.Error("evento enviado a la cola de mensajes muertos", "cola", cola, "tipo", sobre.Tipo, "id", sobre.ID, "error", err)
	_ = entrega.Nack(false, false)
}
