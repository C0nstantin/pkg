package queuer

import (
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RepeatableRejector repeat send message with progressive ttl
// if TTLBase = 1 minute and TTLRang = 10 minutes,
//
// time series will be that:  1 minute, 11 minutes, 21 minutes and next while maxRepeat < repeat
// times(header repeat_number).
//
// After times for repeat will be > MaxRepeat message put to queue with postfix .fail

func StartRepeatableWorker(conf *Config, handler WorkerHandler, TTLBase int32, TTLRang int32, maxRepeat int32) {
	rejector := &RepeatableRejector{
		TTLBase:   TTLBase,
		TTLRang:   TTLRang,
		MaxRepeat: maxRepeat,
		Cnf:       conf,
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
		queName := conf.QueName + postfix
		_, err = worker.channel.QueueDeclare(queName,
			conf.QueueOptions.Durable,
			conf.QueueOptions.AutoDelete,
			conf.QueueOptions.Exclusive,
			conf.QueueOptions.NoWait,
			Args)
		if err != nil {
			log.Fatalf("can't declare queue %s: %s", queName, err)
		}
		err = worker.channel.QueueBind(queName, conf.RoutKey+postfix, conf.Exchange+".topic", false, nil)
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
