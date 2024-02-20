package queuer

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"strconv"
	"time"
)

type RepeatableRejector struct {
	TTLBase   int32 // time base (second)
	TTLRang   int32 // for second repeating TTL = TTLBase + TTLRange*2 (second)
	MaxRepeat int32 // max repeat before send to error que
	Cnf       *Config
}

func (r RepeatableRejector) Reject(delivery *amqp.Delivery) error {
	var currentRepeat int32
	var expiration, que string
	if res, ok := delivery.Headers["repeat_number"]; ok {
		currentRepeat = res.(int32)
	}

	if currentRepeat >= r.MaxRepeat {
		expiration = ""
		delivery.Headers["repeat_number"] = ""
		que = ".fail"
		log.Printf(" message send to  fail que ")
	} else {
		currentRepeat++
		if delivery.Headers == nil {
			delivery.Headers = amqp.Table{}
		}
		delivery.Headers["repeat_number"] = currentRepeat
		expiration = strconv.Itoa(int((r.TTLBase + r.TTLRang*(currentRepeat-1)) * 1000))
		que = ".wait"
		log.Printf(" message send to wait que with ttl  = " + expiration)
	}

	err := PublishMessage(Config{
		DSN:             r.Cnf.DSN,
		Exchange:        delivery.Exchange + ".topic",
		RoutKey:         delivery.RoutingKey + que,
		ExchangeOptions: r.Cnf.ExchangeOptions,
		PublishOptions:  r.Cnf.PublishOptions,
		QueueOptions:    r.Cnf.QueueOptions,
		ConsumeOptions:  r.Cnf.ConsumeOptions,
	}, &amqp.Publishing{
		Headers:         delivery.Headers,
		ContentType:     delivery.ContentType,
		ContentEncoding: delivery.ContentEncoding,
		Expiration:      expiration,
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
	return nil
}
