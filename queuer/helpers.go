package queuer

import (
	"errors"
	amqp "github.com/rabbitmq/amqp091-go"
)

type WorkerTestHandler struct {
	Res     []string
	ResChan chan *amqp.Delivery
	ErrChan chan *amqp.Delivery
	Err     []string
}

func (w *WorkerTestHandler) Handle(d *amqp.Delivery) error {
	if string(d.Body) == "ERROR" {
		w.ErrChan <- d
		w.Err = append(w.Err, string(d.Body))
		return errors.New("test Error ")
	}

	if string(d.Body) == "PANIC" {
		panic("Test Panic")
	}
	w.Res = append(w.Res, string(d.Body))
	w.ResChan <- d
	return nil
}

type WorkerTestRejector struct {
	Res     []string
	ResChan chan *amqp.Delivery
}

func (w *WorkerTestRejector) Reject(d *amqp.Delivery) error {
	/*
		if string(d.Body) == "ERROR" {
			return errors.New("Test Error")
		}

		if string(d.Body) == "PANIC" {
			panic("Test Panic")
		}
	*/
	w.ResChan <- d
	w.Res = append(w.Res, string(d.Body))
	return nil
}
