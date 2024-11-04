package rmqx

import (
	"context"
	"github.com/C0nstantin/pkg/errors"
	"github.com/C0nstantin/pkg/utils"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

// PublishMessage publishes a message to a RabbitMQ exchange.
// It establishes a connection to the RabbitMQ server, creates a channel, declares an exchange,
// and publishes the message to the exchange with the specified routing key.
// It returns an error if any of the steps fail.
func PublishMessage(c Config, publishing *amqp.Publishing) error {

	conn, err := amqp.Dial(c.ConnectionUrl)
	if err != nil {
		return NewFatalError(errors.Errorf(" connect to %s return error:  %v", c.ConnectionUrl, err), publishing.Body)
	}
	defer utils.DeferCloseLog(conn)

	channel, err := conn.Channel()
	if err != nil {
		return NewFatalError(errors.Errorf("PushMessage create Channel error: %v", err), publishing.Body)
	}
	defer utils.DeferCloseLog(channel)

	err = channel.ExchangeDeclare(c.Exchange,
		c.ExchangeOptions.Kind,
		c.ExchangeOptions.Durable,
		c.ExchangeOptions.AutoDelete,
		c.ExchangeOptions.Internal,
		c.ExchangeOptions.NoWait,
		c.ExchangeOptions.Args)
	if err != nil {
		return NewFatalError(errors.Errorf("PushMessage declare exchanger %s error: %v", c.Exchange, err), publishing.Body)
	}

	err = channel.PublishWithContext(
		context.Background(),
		c.Exchange,
		c.RoutKey,
		false,
		false,
		*publishing)
	if err != nil {
		return NewFatalError(errors.Errorf("PushMessage to %s, with routekey %s return error %v ", c.Exchange, c.RoutKey, err), publishing.Body)
	}
	log.Printf("Send message: %s to exchange %s :->  %s ", string(publishing.MessageId), c.Exchange, c.RoutKey)

	return nil
}

// PublishTextMessage publishes a text message using the specified configuration and message data.
// It calls the PublishMessage function internally, passing the appropriate content type and message body.
// It merges the default configuration with the provided one and establishes a connection to the RabbitMQ server.
// Then it creates a channel and declares an exchange based on the configuration.
// Finally, it publishes the message to the exchange with the specified routing key.
// It returns an error if any of the steps fail.
func PublishTextMessage(c Config, message []byte) error {
	return PublishMessage(c, &amqp.Publishing{
		ContentType: "text/plain",
		Body:        message,
	})
}

type Pusher interface {
	PushMessage(ctx context.Context, exchange, routingKey string, publishing *amqp.Publishing) error
	Close() error
}

type PusherImpl struct {
	conn       *amqp.Connection
	ch         *amqp.Channel
	connectUrl string
}

func NewPusherImpl(connectUrl string) *PusherImpl {
	return &PusherImpl{
		connectUrl: connectUrl,
	}
}

func (p *PusherImpl) PushMessage(ctx context.Context, exchange, routingKey string, publishing *amqp.Publishing) error {
	if err := p.connect(); err != nil {
		return NewFatalError(err, publishing.Body)
	}
	err := p.ch.PublishWithContext(ctx, exchange, routingKey, false, false, *publishing)
	if err != nil {
		return NewFatalError(errors.Errorf("PushMessage to %s, with routekey %s return error %v", exchange, routingKey, err), publishing.Body)
	}
	return nil
}

func (p *PusherImpl) connect() error {
	conn, err := amqp.Dial(p.connectUrl)
	if err != nil {
		return errors.Er(err, " connect to %s return,", p.connectUrl)
	}
	p.conn = conn
	ch, err := conn.Channel()
	if err != nil {
		return errors.Errorf("PushMessage create Channel %v", err)
	}
	p.ch = ch
	return nil
}

func (p *PusherImpl) Close() error {
	utils.DeferCloseLog(p.ch)
	utils.DeferCloseLog(p.conn)
	return nil
}
