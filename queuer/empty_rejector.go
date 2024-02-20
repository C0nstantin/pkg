package queuer

import amqp "github.com/rabbitmq/amqp091-go"

type EmptyRejector struct{}

func (e *EmptyRejector) Reject(delivery *amqp.Delivery) error {
	return delivery.Reject(false)
}
