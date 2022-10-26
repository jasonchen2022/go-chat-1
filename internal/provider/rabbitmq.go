package provider

import (
	"context"
	"fmt"
	"strconv"

	"go-chat/config"

	"github.com/streadway/amqp"
)

func NewRabbitMQClient(ctx context.Context, conf *config.Config) *amqp.Connection {

	url := fmt.Sprintf("amqp://%s:%s@%s:%s", conf.RabbitMQ.UserName, conf.RabbitMQ.Password, conf.RabbitMQ.Host, strconv.Itoa(conf.RabbitMQ.Port))
	client, err := amqp.Dial(url)
	if err != nil {
		return nil
	}
	return client
}
