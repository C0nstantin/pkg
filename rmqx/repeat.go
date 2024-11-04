package rmqx

import (
	"fmt"
	"github.com/C0nstantin/pkg/errors"
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
		log.Println(" message send to  fail que ")
	} else {
		currentRepeat++
		if delivery.Headers == nil {
			delivery.Headers = amqp.Table{}
		}
		delivery.Headers["repeat_number"] = currentRepeat
		expiration = strconv.Itoa(int((r.TTLBase + r.TTLRang*(currentRepeat-1)) * 1000))
		que = ".wait"
		log.Println(" message send to wait que with ttl  = " + expiration)
	}

	err := PublishMessage(Config{
		ConnectionUrl:   r.Cnf.ConnectionUrl,
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

func NewRepeatWorkerPool(cnf *Config, workerCount int, handler Handler, errHandler ErrorHandler, TTLBase, TTLRang, MaxRepeat int32) (*WorkerPool, error) {
	rejector := &RepeatableRejector{
		TTLBase:   TTLBase,
		TTLRang:   TTLRang,
		MaxRepeat: MaxRepeat,
		Cnf:       cnf,
	}
	if workerCount <= 0 {
		return nil, errors.New("worker count must be greater than 0")
	}
	if cnf.QueName == "" {
		return nil, errors.New("queue name must be not empty")
	}
	if cnf.Exchange == "" {
		return nil, errors.New("exchange name must be not empty")
	}
	if cnf.RoutKey == "" {
		return nil, errors.New("routing key must be not empty")
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
		conn:    conn,
	}

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
		return nil, errors.Errorf("NewWorker amqp declare exchanger %s error: %s", cnf.Exchange+".topic", err)
	}

	for _, postfix := range []string{".wait", ".fail"} { // declaration retry exchange

		var Args amqp.Table
		if postfix == ".wait" {
			Args = amqp.Table{
				"x-dead-letter-exchange":    cnf.Exchange,
				"x-dead-letter-routing-key": cnf.RoutKey,
			}
		} else {
			Args = nil
		}
		queName := cnf.QueName + postfix
		_, err = ch.QueueDeclare(queName,
			cnf.QueueOptions.Durable,
			cnf.QueueOptions.AutoDelete,
			cnf.QueueOptions.Exclusive,
			cnf.QueueOptions.NoWait,
			Args)
		if err != nil {
			return nil, errors.Errorf("can't declare queue %s: %s", queName, err)
		}
		err = ch.QueueBind(queName, cnf.RoutKey+postfix, cnf.Exchange+".topic", false, nil)
		if err != nil {
			return nil, errors.Errorf("Repitable worker return error when bind retry que: %v ", err)
		}
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
