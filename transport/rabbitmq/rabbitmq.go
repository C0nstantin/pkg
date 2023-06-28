package rabbitmq

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Sender interface {
	Send(key string, env []byte) error
	Close()
}

type Consumer interface {
	ConsumeToKey(ctx context.Context, key, queName string, r chan []byte) error
}

type EventsTransportImpl struct {
	init     bool
	con      *amqp.Connection
	dsn      string
	exchange string
}

func (t *EventsTransportImpl) Send(key string, event []byte) error {
	if !t.init {
		return fmt.Errorf("amqp transport not initial !, please use NewTransportAmqpWS function for init")
	}
	if t.con.IsClosed() {
		return fmt.Errorf("connection is closed! ")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ch, err := t.con.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()
	// logs.Infof("try send events to exchange %s, key %s, events %s", t.exchange, key, event)

	return ch.PublishWithContext(ctx, t.exchange, key, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: 0,
		Priority:     0,
		Body:         event,
	})
}
func (t *EventsTransportImpl) Close() {
	if !t.con.IsClosed() {
		err := t.con.Close()
		if err != nil {
			panic("error close connection" + err.Error())
		}
	}
}
func (t *EventsTransportImpl) ConsumeToKey(ctx context.Context, key, queName string, r chan []byte) error {
	if !t.init {
		return fmt.Errorf("amqp transport not initial!, please use NewTransportAmqpWS function for init")
	}
	if t.con.IsClosed() {
		return fmt.Errorf("connection is closed! ")
	}
	ch, err := t.con.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	que, err := ch.QueueDeclare(queName, false, false, false, false, amqp.Table{"x-expires": 60000})
	if err != nil {
		return err
	}

	err = ch.QueueBind(que.Name, key, t.exchange, false, nil)
	if err != nil {
		return err
	}
	msgs, err := ch.Consume(
		que.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}
	for {
		select {
		case d := <-msgs:
			if d.Body != nil {
				r <- d.Body
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func NewRQTransport(dsn, exchange, kind string) (*EventsTransportImpl, error) {
	conn, err := amqp.Dial(dsn)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(exchange, kind, true, false, false, false, nil)

	if err != nil {
		return nil, err
	}
	return &EventsTransportImpl{
		init:     true,
		con:      conn,
		exchange: exchange,
		dsn:      dsn,
	}, nil
}
