package queuer

import (
	"log"
	"strconv"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RepeatableRejector repeat send message with progressive ttl
// if TTLBase = 1 munute and TTLRang = 10 minutes,
//
// time series will be that:  1 minute, 11 minutes, 21 minutes and next while maxRepeat < repeat
// times(header repeat_number).
//
// After times for repeat will be > MaxRepeat message put to queue with postfix .fail
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

func StartRepeatableWorker(conf *Config, handler WorkerHandler, rejector WorkerRejector, TTLBase int32, TTLRang int32, maxRepeat int32) {

	if rejector == nil {
		rejector = RepeatableRejector{
			TTLBase:   TTLBase,
			TTLRang:   TTLRang,
			MaxRepeat: maxRepeat,
			Cnf:       conf,
		}
	}
	if len(conf.Exchange) == 0 || len(conf.QueName) == 0 || len(conf.RoutKey) == 0 {
		log.Fatalf("Invalid config for repeating")
	}

	worker, err := NewWorker(conf, handler, rejector)
	if err != nil {
		log.Fatalf("Repitable worker return error when create new worker: %v ", err)
	}

	err = worker.channel.ExchangeDeclare(
		conf.Exchange+".topic",
		conf.ExchangeOptions.Kind,
		conf.ExchangeOptions.Durable,
		conf.ExchangeOptions.AutoDelete,
		conf.ExchangeOptions.Internal,
		conf.ExchangeOptions.NoWait,
		conf.ExchangeOptions.Args)
	if err != nil {
		log.Fatalf("NewWorker amqp declare exchanger %s error: %s", conf.Exchange+".topic", err)
	}

	for _, postfix := range []string{".wait", ".fail"} { // declaration retry exchange

		var Args amqp.Table
		if postfix == ".wait" {
			Args = amqp.Table{
				"x-dead-letter-exchange":    conf.Exchange,
				"x-dead-letter-routing-key": conf.RoutKey,
			}
		} else {
			Args = nil
		}

		_, err = worker.channel.QueueDeclare(conf.QueName+postfix,
			conf.QueueOptions.Durable,
			conf.QueueOptions.AutoDelete,
			conf.QueueOptions.Exclusive,
			conf.QueueOptions.NoWait,
			Args)
		if err != nil {
			log.Fatalf("Error after declarate que retry: %v", err)
		}
		err = worker.channel.QueueBind(conf.QueName+postfix, conf.RoutKey+postfix, conf.Exchange+".topic", false, nil)
		if err != nil {
			log.Fatalf("Repitable worker return error when bind retry que: %v ", err)
		}
	}

	go worker.Run()

	for {
		select {
		case err := <-worker.Fatal:
			worker.Stop()
			log.Println("Sleep to 5 second before fail")
			time.Sleep(5 * time.Second)
			log.Fatalf("Fatal error:  %s", err)
		case err := <-worker.Errors:
			log.Printf("Worker return Error %s", err)
		case done := <-worker.Done:
			var requestId string
			if r, ok := done.Headers["X-Request-Id"]; ok {
				requestId = r.(string)
			}
			log.Println("Done : " + requestId)
		}
	}
}
