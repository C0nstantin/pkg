package queuer

import (
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type rejector struct {
}

func (r rejector) Reject(delivery *amqp.Delivery) error {
	return delivery.Reject(false)
}

func StartSimpleWorker(c *Config, handler WorkerHandler, rej WorkerRejector) {
	if rej == nil {
		rej = rejector{}
	}
	worker, err := NewWorker(c, handler, rej)
	if err != nil {
		log.Fatal(err)
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
