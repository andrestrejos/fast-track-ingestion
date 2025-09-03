package amqp


import (
"context"


amqp "github.com/rabbitmq/amqp091-go"
)


type Client struct {
conn *amqp.Connection
channel *amqp.Channel
}


func New(url string) (*Client, error) {
conn, err := amqp.Dial(url)
if err != nil {
return nil, err
}
ch, err := conn.Channel()
if err != nil {
return nil, err
}
return &Client{conn: conn, channel: ch}, nil
}


func (c *Client) Close() {
c.channel.Close()
c.conn.Close()
}


func (c *Client) Declare(exchange, queue, rk string) (amqp.Queue, error) {
err := c.channel.ExchangeDeclare(exchange, "direct", true, false, false, false, nil)
if err != nil {
return amqp.Queue{}, err
}
q, err := c.channel.QueueDeclare(queue, true, false, false, false, nil)
if err != nil {
return amqp.Queue{}, err
}
err = c.channel.QueueBind(q.Name, rk, exchange, false, nil)
return q, err
}


func (c *Client) Publish(ctx context.Context, exchange, rk string, body []byte) error {
return c.channel.PublishWithContext(ctx, exchange, rk, false, false, amqp.Publishing{
ContentType: "application/json",
Body: body,
})
}


func (c *Client) Consume(queue string) (<-chan amqp.Delivery, error) {
return c.channel.Consume(queue, "", false, false, false, false, nil)
}