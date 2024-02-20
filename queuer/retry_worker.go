package queuer

import (
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// StartRetryWorker start worker with RetryRejector with and MaxRetry
// if handler function return error worker call RetryRejector.Reject(message)
// rejector put this message to queue ("{{base_queue}}.retry") with ttl = params TTL second, and when ttl is over resend
// message to base exchange with base route key
// after MaxRetry fail times put message to queue with name "{{base_queue}}.fail"
func StartRetryWorker(conf *Config, handler WorkerHandler, TTL, MaxRetry int32) {
	rejector := &RetryRejector{
		MaxRetry: MaxRetry,
		Cnf:      conf,
	}

	if len(conf.Exchange) == 0 || len(conf.QueName) == 0 || len(conf.RoutKey) == 0 {
		log.Fatalf("Invalid config for repeating")
	}

	Args := amqp.Table{
		"x-dead-letter-exchange":    conf.Exchange + ".topic",
		"x-dead-letter-routing-key": conf.RoutKey + ".retry",
	}

	conf.QueueOptions = QueueOptionsDefaults
	conf.QueueOptions.Args = Args
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

	for _, postfix := range []string{".retry", ".fail"} {
		var Args amqp.Table
		if postfix == ".retry" {
			Args = amqp.Table{
				"x-dead-letter-exchange":    conf.Exchange,
				"x-dead-letter-routing-key": conf.RoutKey,
				"x-message-ttl":             TTL * 1000, // second
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
