package queuer

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	logs "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

func PublishMessage(c Config, publishing *amqp.Publishing) error {
	err := c.MergeDefaults()
	if err != nil {
		return err
	}

	conn, err := amqp.Dial(c.DSN)
	if err != nil {
		return fmt.Errorf(" connect to %s return error:  %w", c.DSN, err)
	}
	defer func(conn *amqp.Connection) {
		err := conn.Close()
		if err != nil {
			logs.Errorf("can not  close connection err = %s", err)
		}
	}(conn)

	channel, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("PushMessage create Channel error: %w", err)
	}

	defer func(channel *amqp.Channel) {
		err := channel.Close()
		if err != nil {
			logs.Errorf("cannot close connection err=%s", err)
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
	logs.Printf("Send message: %s to exchange %s :->  %s ", string(publishing.MessageId), c.Exchange, c.RoutKey)

	return nil
}

func PublishTextMessage(c Config, message []byte) error {
	return PublishMessage(c, &amqp.Publishing{
		ContentType: "text/plain",
		Body:        message,
	})
}
