package rmqx

import (
	"context"
	"github.com/C0nstantin/pkg/errors"
	"github.com/C0nstantin/pkg/log"
	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	ErrConnectionClosed = errors.New("Connection closed. ")
	ErrChanelClosed     = errors.New("Chanel closed. ")
)

// Pool interface represents a pool of workers
type Pool interface {
	Start(ctx context.Context)
	Stop()
}
type Handler interface {
	Handle(delivery *amqp.Delivery, logger log.Logger) error
}

type Worker interface {
	Run(ctx context.Context) error
	Close() error
}

type ErrorHandler interface {
	ErrorHandle(e error, r *amqp.Delivery)
}
type Rejector interface {
	Reject(delivery *amqp.Delivery) error
}

type EmptyRejector struct{}

func (e *EmptyRejector) Reject(delivery *amqp.Delivery) error {
	return delivery.Reject(false)
}

type EmptyHandler struct{}

func (e *EmptyHandler) Handle(delivery *amqp.Delivery, logger log.Logger) error {
	return nil
}
