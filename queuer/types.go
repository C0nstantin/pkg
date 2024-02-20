package queuer

import (
	"errors"
	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	ErrConnectionClosed = errors.New("Connection closed. ")
	ErrChanelClosed     = errors.New("Chanel closed. ")
)

type WorkerHandler interface {
	Handle(delivery *amqp.Delivery) error
}

type EmptyWorkerHandeler struct{}

func (emptyHandler *EmptyWorkerHandeler) Handle(*amqp.Delivery) error {
	return nil
}

type WorkerRejector interface {
	Reject(delivery *amqp.Delivery) error
}
