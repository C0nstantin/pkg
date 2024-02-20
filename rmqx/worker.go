package rmqx

import (
	"context"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"gitlab.wm.local/wm/pkg/errors"
	"log"
	"math/rand"
)

type baseWorker struct {
	name            string
	config          *Config
	conn            *amqp.Connection
	channel         *amqp.Channel
	msgs            <-chan amqp.Delivery
	notifyCloseConn chan *amqp.Error
	notifyCloseChan chan *amqp.Error
	fatalErrors     chan error
	handler         Handler
	rejector        Rejector
	done            chan *amqp.Delivery
	handlerErrors   chan handlerError
	logger          *log.Logger
	errorHandler    ErrorHandler
}

func NewWorker(name string, config Config, conn *amqp.Connection, handler Handler, rejector Rejector, errorHandler ErrorHandler) (Worker, error) {
	if name == "" {
		name = fmt.Sprintf("worker-%d", rand.Int())
	}

	logger := log.New(log.Writer(), name, log.LstdFlags)
	logger.SetPrefix(fmt.Sprintf("[%s] ", name))

	return &baseWorker{
		name:            name,
		config:          &config,
		conn:            conn,
		handler:         handler,
		rejector:        rejector,
		notifyCloseConn: make(chan *amqp.Error),
		notifyCloseChan: make(chan *amqp.Error),
		done:            make(chan *amqp.Delivery),
		handlerErrors:   make(chan handlerError),
		fatalErrors:     make(chan error),
		logger:          logger,
		errorHandler:    errorHandler,
	}, nil

}

func (b *baseWorker) Run(ctx context.Context) error {
	err := b.connect()
	if err != nil {
		return err
	}
	b.logger.Printf("✅ Start consume que %s", b.config.QueName)
	go func() {
		select {
		case err := <-b.notifyCloseChan:
			b.fatalErrors <- fmt.Errorf("%s, %w", ErrChanelClosed, err)
		case err := <-b.notifyCloseConn:
			b.fatalErrors <- fmt.Errorf("%s %w", ErrConnectionClosed, err)
		}
	}()
	go b.run(ctx)
	for {
		select {
		case <-ctx.Done():
			//b.Close()
			return nil
		case msg := <-b.done:
			b.logger.Printf("message %s done", msg.MessageId)
		case err := <-b.handlerErrors:
			b.logger.Printf("handler error: %s  try rejected", err.err)
			b.Reject(err)
		case err := <-b.fatalErrors:
			b.logger.Printf("fatal error in worker: %s", err)
			b.logger.Printf("worker closing")
			b.Close()
			return err
		}
	}
}

func (b *baseWorker) Handle(ctx context.Context, msg *amqp.Delivery) {
	err := b.handler.Handle(msg, b.logger)
	if err != nil {
		b.logger.Printf("Error handle message: %s", err)
		b.handlerErrors <- handlerError{err: err, msg: msg}
		return
	}
	err = msg.Ack(false)
	if err != nil {
		b.fatalErrors <- err
		return
	}
	b.done <- msg
}

func (b *baseWorker) Close() {
	b.logger.Printf("✅ Stop consume que %s", b.config.QueName)
	if !b.channel.IsClosed() {
		err := b.channel.Close()
		if err != nil {
			b.logger.Printf("failed to close channel:  %s", err)
		}
	}
}

func (b *baseWorker) connect() error {
	if b.conn == nil {
		return errors.New("connection is not initialized")
	}
	ch, err := b.conn.Channel()
	if err != nil {
		return errors.E(fmt.Errorf("failed to open a channel: %w", err))
	}
	b.channel = ch
	err = b.channel.Qos(1, 0, false)
	if err != nil {
		return errors.E(fmt.Errorf("failed to set qos: %w", err))
	}

	b.notifyCloseConn = b.conn.NotifyClose(make(chan *amqp.Error))
	b.notifyCloseChan = b.channel.NotifyClose(make(chan *amqp.Error))
	messages, err := b.channel.Consume(
		b.config.QueName,
		b.name,
		b.config.ConsumeOptions.AutoAck,
		b.config.ConsumeOptions.Exclusive,
		b.config.ConsumeOptions.NoLocal,
		b.config.ConsumeOptions.NoWait, nil)

	if err != nil {
		return errors.E(fmt.Errorf("failed to register a consumer: %w", err))
	}
	b.msgs = messages
	return nil
}

func (b *baseWorker) run(ctx context.Context) {
	for msg := range b.msgs {
		b.Handle(ctx, &msg)

	}
}

func (b *baseWorker) Reject(e handlerError) {
	//handle error
	if b.errorHandler != nil {
		b.errorHandler.ErrorHandle(e.err, e.msg)
	}
	err := b.rejector.Reject(e.msg)
	if err != nil {
		b.fatalErrors <- errors.E(fmt.Errorf("failed to reject message: %w", err))
	}
}
