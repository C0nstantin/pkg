package rmqx

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type Config struct {
	ConnectionUrl string `yaml:"connection_url" env:"RABBITMQ_URL" env-required:"true"`
	Exchange      string `yaml:"exchange" env:"RABBITMQ_EXCHANGE" env-default:""`
	RoutKey       string `yaml:"routing_key" env:"RABBITMQ_ROUTING_KEY" env-default:""`
	QueName       string `yaml:"que_name" env:"RABBITMQ_QUEUE_NAME" env-default:""`

	ExchangeOptions ExchangeOptions
	PublishOptions  PublishOptions
	QueueOptions    QueueOptions
	ConsumeOptions  ConsumeOptions
}

type ExchangeOptions struct {
	Kind       string     `yaml:"kind" env:"RABBITMQ_EXCHANGE_KIND" env-default:"direct"`
	Durable    bool       `yaml:"durable" env:"RABBITMQ_EXCHANGE_DURABLE" env-default:"true"`
	AutoDelete bool       `yaml:"auto_delete" env:"RABBITMQ_EXCHANGE_AUTO_DELETE" env-default:"false"`
	Internal   bool       `yaml:"internal" env:"RABBITMQ_EXCHANGE_INTERNAL" env-default:"false"`
	NoWait     bool       `yaml:"nowait" env:"RABBITMQ_EXCHANGE_NOWAIT" env-default:"false"`
	Args       amqp.Table `yaml:"args" env:"RABBITMQ_EXCHANGE_ARGS"`
}

type PublishOptions struct {
	Mandatory bool `yaml:"mandatory" env:"RABBITMQ_EXCHANGE_MANDATORY" env-default:"false"`
	Immediate bool `yaml:"immediate" env:"RABBITMQ_EXCHANGE_IMMEDIATE" env-default:"false"`
}

type QueueOptions struct {
	Durable    bool       `yaml:"durable" env:"RABBITMQ_QUEUE_DURABLE" env-default:"true"`
	AutoDelete bool       `yaml:"auto_delete" env:"RABBITMQ_QUEUE_DELETE" env-default:"false"`
	Exclusive  bool       `yaml:"exclusive" env:"RABBITMQ_QUEUE_EXCLUSIVE" env-default:"false"`
	NoWait     bool       `yaml:"nowait" env:"RABBITMQ_QUEUE_NOWAIT" env-default:"false"`
	Args       amqp.Table `yaml:"args" env:"RABBITMQ_QUEUE_ARGS"`
}

type ConsumeOptions struct {
	AutoAck   bool       `yaml:"auto_ack" env:"RABBITMQ_CONSUME_AUTO_ACK" env-default:"false"`
	Exclusive bool       `yaml:"exclusive" env:"RABBITMQ_CONSUME_EXCLUSIVE" env-default:"false"`
	NoLocal   bool       `yaml:"noLocal" env:"RABBITMQ_CONSUME_NO_LOCAL" env-default:"false"`
	NoWait    bool       `yaml:"nowait" env:"RABBITMQ_CONSUME_NOWAIT" env-default:"false"`
	Args      amqp.Table `yaml:"args" env:"RABBITMQ_CONSUME_ARGS"`
}
