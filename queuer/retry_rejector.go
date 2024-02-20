package queuer

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"time"
)

// RetryRejector Rejector for retry worker
type RetryRejector struct {
	MaxRetry int32 // max retry
	Cnf      *Config
}

func (r *RetryRejector) Reject(delivery *amqp.Delivery) error {
	var currentRepeat int32
	if res, ok := delivery.Headers["x-death"]; ok {
		for _, t := range res.([]interface{}) {
			currentRepeat = int32(t.(amqp.Table)["count"].(int64))
		}
	}

	if currentRepeat+1 >= r.MaxRetry {
		err := PublishMessage(Config{
			DSN:             r.Cnf.DSN,
			Exchange:        delivery.Exchange + ".topic",
			RoutKey:         delivery.RoutingKey + ".fail",
			ExchangeOptions: r.Cnf.ExchangeOptions,
			PublishOptions:  r.Cnf.PublishOptions,
			QueueOptions:    r.Cnf.QueueOptions,
			ConsumeOptions:  r.Cnf.ConsumeOptions,
		}, &amqp.Publishing{
			Headers:         delivery.Headers,
			ContentType:     delivery.ContentType,
			ContentEncoding: delivery.ContentEncoding,
			MessageId:       delivery.MessageId,
			Timestamp:       time.Time{},
			Type:            delivery.Type,
			UserId:          delivery.UserId,
			AppId:           delivery.AppId,
			Body:            delivery.Body,
		})
		if err != nil {
			return err
		}
		err = delivery.Ack(false)
		if err != nil {
			return err
		}
	} else {
		err := delivery.Reject(false)
		if err != nil {
			return err
		}
	}
	return nil
}
