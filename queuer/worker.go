package queuer

import (
	"fmt"
	"github.com/pkg/errors"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type WorkerInterface interface {
	Handle(delivery *amqp.Delivery) error
	Reject(delivery *amqp.Delivery) error
	Close()
}

type Worker struct {
	config          *Config
	conn            *amqp.Connection
	channel         *amqp.Channel
	que             *amqp.Queue
	Errors          chan error
	notifyCloseConn chan *amqp.Error
	notifyCloseChan chan *amqp.Error
	Done            chan *amqp.Delivery
	Handler         WorkerHandler
	Rejector        WorkerRejector
	Fatal           chan error
}

func NewWorker(config *Config, handler WorkerHandler, rejector WorkerRejector) (*Worker, error) {
	if handler == nil {
		return nil, errors.New("handler can not be nil")
	}

	if rejector == nil {
		rejector = &EmptyRejector{}
	}
	if config == nil {
		return nil, errors.New("config can not be nil")
	}

	err := config.MergeDefaults()
	if err != nil {
		return nil, err
	}

	//connection
	conn, err := amqp.Dial(config.DSN)
	if err != nil {
		return nil, errors.Errorf("not possible to connect to %s:  %s", config.DSN, err)
	}

	//create channel
	channel, err := conn.Channel()
	if err != nil {
		return nil, errors.Errorf("can't create channel:  %s", err)
	}

	err = channel.Qos(1, 0, false)
	if err != nil {
		return nil, errors.Errorf("can't set Qos error:  %s", err)
	}
	if len(config.Exchange) > 0 {
		err = channel.ExchangeDeclare(
			config.Exchange,
			config.ExchangeOptions.Kind,
			config.ExchangeOptions.Durable,
			config.ExchangeOptions.AutoDelete,
			config.ExchangeOptions.Internal,
			config.ExchangeOptions.NoWait,
			config.ExchangeOptions.Args)
		if err != nil {
			return nil, errors.Errorf("can't declare exchange %s error: %s", config.Exchange, err)
		}
	}

	que, err := channel.QueueDeclare(config.QueName,
		config.QueueOptions.Durable,
		config.QueueOptions.AutoDelete,
		config.QueueOptions.Exclusive,
		config.QueueOptions.NoWait,
		config.QueueOptions.Args)
	if err != nil {
		return nil, errors.Errorf("can't declare queue %s: %s", config.QueName, err)
	}

	if len(config.Exchange) > 0 {
		err = channel.QueueBind(
			que.Name,
			config.RoutKey,
			config.Exchange,
			false,
			nil)
		if err != nil {
			return nil, errors.Errorf("can't bind  que %s to exchange %s with routeKey %s error %s",
				que.Name,
				config.Exchange,
				config.RoutKey,
				err)
		}
	}
	conNotify := conn.NotifyClose(make(chan *amqp.Error))
	chanNotify := channel.NotifyClose(make(chan *amqp.Error))

	return &Worker{
		conn:            conn,
		channel:         channel,
		que:             &que,
		config:          config,
		notifyCloseConn: conNotify,
		notifyCloseChan: chanNotify,
		Errors:          make(chan error),
		Fatal:           make(chan error),
		Done:            make(chan *amqp.Delivery),
		Handler:         handler,
		Rejector:        rejector,
	}, nil
}

func (w *Worker) Run() {
	messages, err := w.channel.Consume(
		w.que.Name,
		"",
		w.config.ConsumeOptions.AutoAck,
		w.config.ConsumeOptions.Exclusive,
		w.config.ConsumeOptions.NoLocal,
		w.config.ConsumeOptions.NoWait,
		w.config.ConsumeOptions.Args)
	if err != nil {
		w.Fatal <- err
	}
	log.Printf("✅ Start consume que %s", w.config.QueName)
	go func() {
		select {
		case err := <-w.notifyCloseChan:
			w.Fatal <- fmt.Errorf("%s, %w", ErrChanelClosed, err)
		case err := <-w.notifyCloseConn:
			w.Fatal <- fmt.Errorf("%s %w", ErrConnectionClosed, err)
		}
	}()

	forever := make(chan bool)
	go func(<-chan amqp.Delivery) {
		for msg := range messages {
			err := w.Handler.Handle(&msg)
			if err != nil {
				w.Errors <- fmt.Errorf("handle message with  %s return error: %w, and try send to rejector", msg, err)
				if fErr := w.Rejector.Reject(&msg); fErr != nil {
					w.Fatal <- fErr
				}
			} else {
				w.Done <- &msg
				err := msg.Ack(false)
				if err != nil {
					log.Printf("Error ack message: %s", err)
					return
				}
			}
		}
	}(messages)
	<-forever
}

func (w *Worker) Stop() {
	err := w.channel.Close()
	if err != nil {
		return
	}
	time.Sleep(2 * time.Second)
	err = w.conn.Close()
	if err != nil {
		log.Printf("failed close amqp connection , error: %s", err)
		return
	}
}
