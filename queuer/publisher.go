package queuer

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"golang.org/x/net/context"
)

// PublishMessage publishes a message to a RabbitMQ exchange.
// It establishes a connection to the RabbitMQ server, creates a channel, declares an exchange,
// and publishes the message to the exchange with the specified routing key.
// It returns an error if any of the steps fail.
func PublishMessage(c Config, publishing *amqp.Publishing) error {
	err := c.MergeDefaults()
	if err != nil {
		return err
	}
	conn, err := amqp.Dial(c.DSN)
	if err != nil {
		return fmt.Errorf("PushMessage connect to %s return error:  %w", c.DSN, err)
	}
	defer func(conn *amqp.Connection) {
		err := conn.Close()
		if err != nil {
			log.Printf("can not close connection err = %s", err)
		}
	}(conn)

	channel, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("PushMessage create Channel error: %w", err)
	}

	defer func(channel *amqp.Channel) {
		err := channel.Close()
		if err != nil {
			log.Printf("cannot close connection error: %s", err)
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
