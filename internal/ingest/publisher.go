package ingest


import (
"context"
"encoding/json"
"sync"


"github.com/andrestrejos/fast-track-ingestion/internal/amqp"
"github.com/andrestrejos/fast-track-ingestion/internal/domain"
)


type Publisher struct {
AMQP *amqp.Client
Exchange string
RoutingKey string
}


func (p *Publisher) Run(ctx context.Context, wg *sync.WaitGroup, done chan<- struct{}) {
wg.Add(1)
go func() {
defer wg.Done()
events := []domain.PaymentEvent{
{UserID: 1, PaymentID: 1, DepositAmount: 10},
{UserID: 1, PaymentID: 2, DepositAmount: 20},
{UserID: 2, PaymentID: 3, DepositAmount: 20},
}
for _, e := range events {
b, _ := json.Marshal(e)
_ = p.AMQP.Publish(ctx, p.Exchange, p.RoutingKey, b)
}
done <- struct{}{}
}()
}