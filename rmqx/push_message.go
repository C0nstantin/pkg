package rmqx

import (
	"context"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"gitlab.wm.local/wm/pkg/errors"
	"gitlab.wm.local/wm/pkg/log"
)

// PublishMessage publishes a message to a RabbitMQ exchange.
// It establishes a connection to the RabbitMQ server, creates a channel, declares an exchange,
// and publishes the message to the exchange with the specified routing key.
// It returns an error if any of the steps fail.
func PublishMessage(c Config, publishing *amqp.Publishing) error {

	conn, err := amqp.Dial(c.ConnectionUrl)
	if err != nil {
		return fmt.Errorf(" connect to %s return error:  %w", c.ConnectionUrl, err)
	}
	defer func(conn *amqp.Connection) {
		err := conn.Close()
		if err != nil {
			log.Printf("can not  close connection err = %s", err)
		}
	}(conn)

	channel, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("PushMessage create Channel error: %w", err)
	}

	defer func(channel *amqp.Channel) {
		err := channel.Close()
		if err != nil {
			log.Printf("cannot close connection err=%s", err)
		}
	}(channel)
	err = channel.ExchangeDeclare(c.Exchange,
		c.ExchangeOptions.Kind,
		c.ExchangeOptions.Durable,
		c.ExchangeOptions.AutoDelete,
		c.ExchangeOptions.Internal,
		c.ExchangeOptions.NoWait,
		c.ExchangeOptions.Args)
	if err != nil {
		return fmt.Errorf("PushMessage declare exchanger %s error: %w", c.Exchange, err)
	}

	err = channel.PublishWithContext(
		context.Background(),
		c.Exchange,
		c.RoutKey,
		false,
		false,
		*publishing)
	if err != nil {
		return fmt.Errorf("PushMessage to %s, with routekey %s return error %w ", c.Exchange, c.RoutKey, err)
	}
	log.Infof("Send message: %s to exchange %s :->  %s ", string(publishing.MessageId), c.Exchange, c.RoutKey)

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
		return err
	}
	err := p.ch.PublishWithContext(ctx, exchange, routingKey, false, false, *publishing)
	if err != nil {
		return errors.E(fmt.Errorf("PushMessage to %s, with routekey %s return error %w ", exchange, routingKey, err))
	}
	return nil
}

func (p *PusherImpl) connect() error {
	conn, err := amqp.Dial(p.connectUrl)
	if err != nil {
		return errors.E(fmt.Errorf(" connect to %s return error:  %w", p.connectUrl, err))
	}
	p.conn = conn
	ch, err := conn.Channel()
	if err != nil {
		return errors.E(fmt.Errorf("PushMessage create Channel error: %w", err))
	}
	p.ch = ch
	return nil
}

func (p *PusherImpl) Close() error {
	if p.ch != nil {
		err := p.ch.Close()
		if err != nil {
			log.Printf("cannot close connection err=%s", err)
		}
	}
	if p.conn != nil {
		err := p.conn.Close()
		if err != nil {
			log.Printf("cannot close connection err=%s", err)
		}
	}
	return nil
}
