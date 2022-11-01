package provider

import (
	"context"
	"fmt"
	"strconv"

	"go-chat/config"

	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

func NewRabbitMQClient(ctx context.Context, conf *config.Config) *amqp.Connection {

	url := fmt.Sprintf("amqp://%s:%s@%s:%s", conf.RabbitMQ.UserName, conf.RabbitMQ.Password, conf.RabbitMQ.Host, strconv.Itoa(conf.RabbitMQ.Port))
	client, err := amqp.Dial(url)
	if err != nil {
		logrus.Error("初始化MQ出错：", err.Error())
		return nil
	}
	return client
}
