package queuer

import (
	"log"
	"time"
)

func StartSimpleWorker(c *Config, handler WorkerHandler, rej WorkerRejector) error {

	worker, err := NewWorker(c, handler, rej)
	if err != nil {
		return err
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
