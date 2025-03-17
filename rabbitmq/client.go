package rabbitmq

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//go:generate mockgen -source=client.go -destination=mocks/rabbit.go
type RabbitConfig struct {
	Rabbit_User string `env:"RABBITMQ_USER" env-required:"true" env-description:"RabbitMQ user"`
	Rabbit_Pass string `env:"RABBITMQ_PASS" env-required:"true" env-description:"RabbitMQ password"`
	Rabbit_Host string `env:"RABBITMQ_HOST" env-required:"true" env-description:"RabbitMQ hosting address"`
	Rabbit_Port string `env:"RABBITMQ_PORT" env-required:"true" env-description:"RabbitMQ port"`
}

type logger interface {
	Info(...any)
	Infof(string, ...any)
	Errorf(string, ...any)
}
type storage interface {
	IncreateBookCount(context.Context, primitive.ObjectID) error
}
type RabbitService struct {
	Conn     *amqp.Connection
	Logger   logger
	Storage  storage
	Channels []*amqp.Channel
}

func New(config *RabbitConfig, logger logger, storage storage) (*RabbitService, error) {
	connection, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s/", config.Rabbit_User, config.Rabbit_Pass, config.Rabbit_Host, config.Rabbit_Port))
	if err != nil {
		return nil, err
	}
	return &RabbitService{
		Conn:    connection,
		Logger:  logger,
		Storage: storage,
	}, nil
}
func (s *RabbitService) CreateChannel() (*amqp.Channel, error) {
	channel, err := s.Conn.Channel()
	if err != nil {
		return nil, err
	}
	s.Channels = append(s.Channels, channel)
	return channel, nil
}
func (s *RabbitService) CloseChannel(channel *amqp.Channel) {
	ChannelsCopy := make([]*amqp.Channel, len(s.Channels))
	copy(ChannelsCopy, s.Channels)

	s.Channels = make([]*amqp.Channel, 0)
	for _, ch := range ChannelsCopy {
		if ch == channel {
			continue
		}
		s.Channels = append(s.Channels, ch)
	}

	channel.Close()
}
func (s *RabbitService) Close() error {
	for _, c := range s.Channels {
		if err := c.Close(); err != nil {
			return err
		}
	}
	return s.Conn.Close()
}
