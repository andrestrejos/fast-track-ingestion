package ingest

import (
	"context"
	"encoding/json"
	"log"

	"github.com/andrestrejos/fast-track-ingestion/internal/amqp"
	"github.com/andrestrejos/fast-track-ingestion/internal/db"
	"github.com/andrestrejos/fast-track-ingestion/internal/domain"
)

type Consumer struct {
	AMQP      *amqp.Client
	Repo      *db.Repo
	Queue     string
	Processed chan<- struct{} // canal para notificar mensajes procesados
}

func (c *Consumer) Run(ctx context.Context, stop <-chan struct{}) error {
	msgs, err := c.AMQP.Consume(c.Queue)
	if err != nil {
		return err
	}
	log.Println("Consumer attached to queue:", c.Queue)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-stop:
			return nil
		case d, ok := <-msgs:
			if !ok {
				return nil
			}

			var e domain.PaymentEvent
			if err := json.Unmarshal(d.Body, &e); err != nil {
				log.Printf("json error: %v", err)
				d.Nack(false, false) // no requeue
				continue
			}

			// intenta insertar
			if err := c.Repo.InsertPayment(ctx, e); err != nil {
				if err == db.ErrDuplicate {
					log.Println("duplicate detected -> inserting into skipped_messages:", e)
					if err2 := c.Repo.InsertSkipped(ctx, e); err2 != nil {
						log.Printf("skipped insert error: %v", err2)
						d.Nack(false, false)
						continue
					}
					d.Ack(false) // duplicado manejado correctamente
				} else {
					log.Printf("insert error: %v", err)
					d.Nack(false, false)
					continue
				}
			} else {
				log.Println("insert ok:", e)
				d.Ack(false)
			}

			// notifica que un mensaje fue procesado
			if c.Processed != nil {
				select {
				case c.Processed <- struct{}{}:
				default:
				}
			}
		}
	}
}
