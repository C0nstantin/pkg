package rmqx

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"gitlab.wm.local/wm/pkg/errors"
	"sync"
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
			ConnectionUrl:   r.Cnf.ConnectionUrl,
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

func NewRetryWorkerPool(cnf *Config, workerCount int, handler Handler, errHandler ErrorHandler, TTL, MaxRetry int32) (*WorkerPool, error) {
	rejector := &RetryRejector{
		MaxRetry: MaxRetry,
		Cnf:      cnf,
	}

	if workerCount <= 0 {
		return nil, errors.New("worker count must be greater than 0")
	}
	if handler == nil {
		return nil, errors.New("handler must be not nil")
	}
	conn, err := amqp.Dial(cnf.ConnectionUrl)
	if err != nil {
		return nil, errors.E(err)
	}
	pool := &WorkerPool{
		workers: make([]Worker, workerCount),
		count:   workerCount,
		conn:    conn,
		wg:      sync.WaitGroup{},
	}

	if len(cnf.Exchange) == 0 || len(cnf.QueName) == 0 || len(cnf.RoutKey) == 0 {
		return nil, errors.New("Invalid config for repeating")
	}

	Args := amqp.Table{
		"x-dead-letter-exchange":    cnf.Exchange + ".topic",
		"x-dead-letter-routing-key": cnf.RoutKey + ".retry",
	}

	cnf.QueueOptions.Args = Args

	err = initSimpleQue(conn, cnf)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, errors.E(err)
	}
	err = ch.ExchangeDeclare(
		cnf.Exchange+".topic",
		cnf.ExchangeOptions.Kind,
		cnf.ExchangeOptions.Durable,
		cnf.ExchangeOptions.AutoDelete,
		cnf.ExchangeOptions.Internal,
		cnf.ExchangeOptions.NoWait,
		cnf.ExchangeOptions.Args)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("NewWorker amqp declare exchanger %s error: %s", cnf.Exchange+".topic", err))
	}

	for _, postfix := range []string{".retry", ".fail"} {
		var Args amqp.Table
		if postfix == ".retry" {
			Args = amqp.Table{
				"x-dead-letter-exchange":    cnf.Exchange,
				"x-dead-letter-routing-key": cnf.RoutKey,
				"x-message-ttl":             TTL * 1000, // second
			}
		} else {
			Args = nil
		}

		_, err = ch.QueueDeclare(cnf.QueName+postfix,
			cnf.QueueOptions.Durable,
			cnf.QueueOptions.AutoDelete,
			cnf.QueueOptions.Exclusive,
			cnf.QueueOptions.NoWait,
			Args)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Error after declarate que retry: %v", err))
		}
		err = ch.QueueBind(cnf.QueName+postfix, cnf.RoutKey+postfix, cnf.Exchange+".topic", false, nil)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Repitable worker return error when bind retry que: %v ", err))
		}
	}
	err = ch.Close()
	if err != nil {
		return nil, errors.E(err)
	}
	for i := 0; i < workerCount; i++ {
		worker, err := NewWorker(fmt.Sprintf("worker-%d", i), cnf, conn, handler, rejector, errHandler) // Replace with your worker implementation
		if err != nil {
			return nil, err
		}
		pool.workers[i] = worker
	}
	return pool, nil
}
