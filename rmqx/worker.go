package rmqx

import (
	"context"
	"fmt"
	"github.com/C0nstantin/pkg/errors"
	"github.com/C0nstantin/pkg/log"
	"github.com/C0nstantin/pkg/utils"
	amqp "github.com/rabbitmq/amqp091-go"
	"math/rand"
)

type internalError struct {
	err error
	msg *amqp.Delivery
}
type FatalError struct {
	err error
	msg []byte
}

func (f FatalError) Error() string {
	return fmt.Sprintf("worker Fatal error: %s", f.err)
}

func (f FatalError) Unwrap() error {
	return f.err
}

func NewFatalError(err error, msg []byte) error {
	return &FatalError{err: err, msg: msg}
}

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
	errors          chan internalError
	logger          log.Logger
	errorHandler    ErrorHandler
}

func NewWorker(name string, config *Config, conn *amqp.Connection, handler Handler, rejector Rejector, errorHandler ErrorHandler) (Worker, error) {
	if name == "" {
		name = fmt.Sprintf("worker-%d", rand.Int())
	}

	logger := log.NewLogger()
	logger.AddField("worker", name)

	return &baseWorker{
		name:            name,
		config:          config,
		conn:            conn,
		handler:         handler,
		rejector:        rejector,
		notifyCloseConn: make(chan *amqp.Error),
		notifyCloseChan: make(chan *amqp.Error),
		done:            make(chan *amqp.Delivery),
		errors:          make(chan internalError),
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
	b.logger.Infof("✅ Start consume que %s, exchange %s, routing key %s", b.config.QueName, b.config.Exchange, b.config.RoutKey)
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
			b.logger.Info("worker closing")
			utils.DeferCloseLog(b)
			return nil
		case msg := <-b.done:
			b.logger.Printf("message %s done", msg.MessageId)
			workerMetrics.MsgsHandled.Inc()
		case err := <-b.errors:
			b.logger.Printf("handler error: %s  try rejected", err.err)
			workerMetrics.MsgsRejected.Inc()
			b.Reject(err)
		case err := <-b.fatalErrors:
			b.logger.Errorf("fatal error in worker: %s", err)
			b.logger.Errorf("trace error %+v", err)
			b.logger.Info("worker closing")
			utils.DeferCloseLog(b)
			return err
		}
	}
}

func (b *baseWorker) Handle(ctx context.Context, msg *amqp.Delivery) {
	errChan := make(chan error)
	done := make(chan struct{})
	go func() {
		err := b.handler.Handle(msg, b.logger)
		var e *FatalError
		if errors.As(err, &e) {
			_ = msg.Reject(false)
			b.fatalErrors <- e
			return
		}
		if err != nil {
			errChan <- err
			return
		}
		done <- struct{}{}
	}()

	select {
	case err := <-errChan:
		b.errors <- internalError{err: err, msg: msg}
		b.logger.Errorf("Error handle message: %s", err)
		return

	case <-ctx.Done():
		b.logger.Errorf("Context done: %s", ctx.Err())
		return

	case <-done:
		b.logger.Printf("message %s done", msg.Body)
		err := msg.Ack(false)
		if err != nil {
			b.logger.Errorf("Error ack message: %s", err)
			b.fatalErrors <- err
			return
		}
		b.done <- msg
	}
}

func (b *baseWorker) Close() error {
	b.logger.Infof("✅ Stop consume que %s", b.config.QueName)
	if !b.channel.IsClosed() {
		err := b.channel.Close()
		if err != nil {
			b.logger.Errorf("failed to close channel:  %s", err)
			return err
		}
	}
	return nil
}

func (b *baseWorker) connect() error {
	if b.conn == nil {
		return errors.New("can not initialize connection")
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
		workerMetrics.MsgsReceived.Inc()
		b.Handle(ctx, &msg)

	}
}

func (b *baseWorker) Reject(e internalError) {
	//handle error
	if b.errorHandler != nil {
		b.errorHandler.ErrorHandle(e.err, e.msg)
	}
	err := b.rejector.Reject(e.msg)
	if err != nil {
		b.fatalErrors <- errors.Errorf("failed to reject message: %v", err)
	}
}
