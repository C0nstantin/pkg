package rmqx

import (
	"context"
	"fmt"
	"github.com/C0nstantin/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
	"sync"
)

type WorkerPool struct {
	workers []Worker
	conn    *amqp.Connection
	wg      *sync.WaitGroup
}

func (p *WorkerPool) Start(ctx context.Context) error {

	errChan := make(chan error, len(p.workers))
	if p.wg == nil {
		p.wg = &sync.WaitGroup{}
	}
	for _, worker := range p.workers {
		p.wg.Add(1)
		go func(w Worker) {
			defer p.wg.Done()
			err := w.Run(ctx)
			if err != nil {
				errChan <- err
				return
			}
		}(worker)
	}
	go func() { initMetrics() }()
	select {
	case <-ctx.Done():
		err := p.Stop()
		if err != nil {
			return err
		}
		return nil
	case err := <-errChan:
		return errors.E(err)
	}
}

func (p *WorkerPool) Stop() (err error) {

	if !p.conn.IsClosed() {
		err = p.conn.Close()
		if err != nil {
			fmt.Printf("failed to close connection:  %s", err)
		}
	}
	for _, worker := range p.workers {
		err := worker.Close()
		if err != nil {
			fmt.Printf("failed to close worker:  %s", err)
		}
	}
	p.wg.Wait()
	return err
}

func initSimpleQue(conn *amqp.Connection, config *Config) error {
	channel, err := conn.Channel()
	if err != nil {
		return errors.E(err)
	}
	_, err = channel.QueueDeclare(
		config.QueName,
		config.QueueOptions.Durable,
		config.QueueOptions.AutoDelete,
		config.QueueOptions.Exclusive,
		config.QueueOptions.NoWait,
		config.QueueOptions.Args)
	if err != nil {
		return errors.E(err)
	}
	if config.Exchange != "" {
		err = channel.ExchangeDeclare(
			config.Exchange,
			config.ExchangeOptions.Kind,
			config.ExchangeOptions.Durable,
			config.ExchangeOptions.AutoDelete,
			config.ExchangeOptions.Internal,
			config.ExchangeOptions.NoWait,
			config.ExchangeOptions.Args)
		if err != nil {
			return errors.E(err)
		}
		err = channel.QueueBind(
			config.QueName,
			config.RoutKey,
			config.Exchange,
			false,
			nil)
	}
	if err != nil {
		return errors.E(err)
	}
	err = channel.Close()
	if err != nil {
		return errors.E(err)
	}
	return nil
}
