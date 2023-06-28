package queuer

import (
	"fmt"
	"os"

	"github.com/imdario/mergo"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Config struct {
	DSN             string
	Exchange        string
	RoutKey         string
	QueName         string
	ExchangeOptions ExchangeOptions
	PublishOptions  PublishOptions
	QueueOptions    QueueOptions
	ConsumeOptions  ConsumeOptions
}

type ExchangeOptions struct {
	Kind       string
	Durable    bool
	AutoDelete bool
	Internal   bool
	NoWait     bool
	Args       amqp.Table
}

type PublishOptions struct {
	Mandatory bool
	Immediate bool
}

type QueueOptions struct {
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Args       amqp.Table
}

type ConsumeOptions struct {
	AutoAck   bool
	Exclusive bool
	NoLocal   bool
	NoWait    bool
	Args      amqp.Table
}

var ConfigDefaults = &Config{
	DSN:             "",
	Exchange:        "",
	RoutKey:         "",
	QueName:         "",
	ExchangeOptions: ExchangeOptionsDefaults,
	PublishOptions:  PublishOptionsDefaults,
	QueueOptions:    QueueOptionsDefaults,
	ConsumeOptions:  ConsumeOptionsDefaults,
}

var ExchangeOptionsDefaults = ExchangeOptions{
	Kind:       "direct",
	Durable:    true,
	AutoDelete: false,
	Internal:   false,
	NoWait:     false,
	Args:       nil,
}

var PublishOptionsDefaults = PublishOptions{
	Mandatory: false,
	Immediate: false,
}

var QueueOptionsDefaults = QueueOptions{
	Durable:    true,
	AutoDelete: false,
	Exclusive:  false,
	NoWait:     false,
	Args:       nil,
}

var ConsumeOptionsDefaults = ConsumeOptions{
	AutoAck:   false,
	Exclusive: false,
	NoLocal:   false,
	NoWait:    false,
	Args:      nil,
}

func (c *Config) MergeDefaults() error {
	if err := mergo.Merge(c, ConfigDefaults); err != nil {
		return fmt.Errorf("merge with default config error %w ", err)
	}
	if dsn, exists := os.LookupEnv("RABBITMQ_DSN"); exists && c.DSN == "" {
		c.DSN = dsn
	}

	if exName, exists := os.LookupEnv("EX_NAME"); exists && c.Exchange == "" {
		c.Exchange = exName
	}

	return nil
}
