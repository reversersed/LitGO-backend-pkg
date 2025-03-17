package rabbitmq

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type rabbitClient struct {
	*amqp.Connection
}
type RabbitConfig struct {
	Rabbit_User string `env:"RABBITMQ_USER" env-required:"true" env-description:"RabbitMQ user"`
	Rabbit_Pass string `env:"RABBITMQ_PASS" env-required:"true" env-description:"RabbitMQ password"`
	Rabbit_Host string `env:"RABBITMQ_HOST" env-required:"true" env-description:"RabbitMQ hosting address"`
	Rabbit_Port string `env:"RABBITMQ_PORT" env-required:"true" env-description:"RabbitMQ port"`
}

func New(config *RabbitConfig) (*rabbitClient, error) {
	connection, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s/", config.Rabbit_User, config.Rabbit_Pass, config.Rabbit_Host, config.Rabbit_Port))
	if err != nil {
		return nil, err
	}
	return &rabbitClient{connection}, nil
}
