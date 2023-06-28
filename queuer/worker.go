package queuer

import (
	"errors"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	log "github.com/sirupsen/logrus"
	"time"
)

var (
	ErrConnectionClosed = errors.New("Connection closed. ")
	ErrChanelClosed     = errors.New("Chanel closed. ")
)

type WorkerHandler interface {
	Handle(delivery *amqp.Delivery) error
}

type WorkerRejector interface {
	Reject(delivery *amqp.Delivery) error
}

type EmptyWorkerHandeler struct{}

func (emptyHandler *EmptyWorkerHandeler) Handle(*amqp.Delivery) error {
	return nil
}

type EmptyWorkerRejector struct{}

func (e *EmptyWorkerRejector) Reject(delivery *amqp.Delivery) error {
	return delivery.Reject(false)
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
	err := config.MergeDefaults()
	if err != nil {
		return nil, err
	}
	conn, err := amqp.Dial(config.DSN)
	if err != nil {
		return nil, fmt.Errorf("NewWorker amqp connect to %s error:  %w", config.DSN, err)
	}
	channel, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("NewWorker amqp create Channel error: %w", err)
	}
	err = channel.Qos(1, 0, false)
	if err != nil {
		return nil, err
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
			return nil, fmt.Errorf("NewWorker amqp declare exchanger %s error: %w", config.Exchange, err)
		}
	}

	que, err := channel.QueueDeclare(config.QueName,
		config.QueueOptions.Durable,
		config.QueueOptions.AutoDelete,
		config.QueueOptions.Exclusive,
		config.QueueOptions.NoWait,
		config.QueueOptions.Args)
	if err != nil {
		return nil, fmt.Errorf("NewWorker amqp que %s declare error: %w", config.QueName, err)
	}

	if len(config.Exchange) > 0 {
		err = channel.QueueBind(
			que.Name,
			config.RoutKey,
			config.Exchange,
			false,
			nil)
		if err != nil {
			return nil, fmt.Errorf("NewWorker amqp que bind %s to exchange %s with routeKey %s error %w",
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
				fErr := w.Rejector.Reject(&msg)
				if fErr != nil {
					w.Fatal <- fErr
				}
				w.Errors <- fmt.Errorf("message with id= %s, error: %w", msg.MessageId, err)
			} else {
				w.Done <- &msg

				err := msg.Ack(false)
				if err != nil {
					log.Errorf("can not send ack error=%s", err)
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
		log.Errorf("can nat close connection = %s", err)
		return
	}
}
