package rmqx

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"gitlab.wm.local/wm/pkg/errors"
	"sync"
)

func NewSimpleWorkerPool(config *Config, workerCount int, handler Handler, errorHandler ErrorHandler) (*WorkerPool, error) {
	if workerCount <= 0 {
		return nil, errors.New("worker count must be greater than 0")
	}
	if config.QueName == "" {
		return nil, errors.New("queue name must be not empty")
	}
	if handler == nil {
		return nil, errors.New("handler must be not nil")
	}
	conn, err := amqp.Dial(config.ConnectionUrl)
	if err != nil {
		return nil, errors.E(err)
	}
	pool := &WorkerPool{
		workers: make([]Worker, workerCount),
		count:   workerCount,
		conn:    conn,
		wg:      sync.WaitGroup{},
	}

	err = initSimpleQue(conn, config)
	if err != nil {
		return nil, err
	}

	for i := 0; i < workerCount; i++ {
		worker, err := NewWorker(fmt.Sprintf("worker-%d", i), config, conn, handler, &EmptyRejector{}, errorHandler) // Replace with your worker implementation
		if err != nil {
			return nil, err
		}
		pool.workers[i] = worker
	}
	return pool, nil
}
