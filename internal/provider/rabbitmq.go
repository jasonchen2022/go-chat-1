package provider

import (
	"context"
	"fmt"
	"strconv"

	"go-chat/config"
	"go-chat/internal/pkg/logger"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/producer"
)

func NewRocketMQClient(ctx context.Context, conf *config.Config) rocketmq.Producer {
	host := fmt.Sprintf("%s:%s", conf.RabbitMQ.Host, strconv.Itoa(conf.RabbitMQ.Port))
	p, err := rocketmq.NewProducer(
		producer.WithNameServer([]string{host}),
	)
	if err != nil {
		logger.Error("初始化MQ出错：", err.Error())
		return nil
	}
	err = p.Start()

	if err != nil {
		logger.Error("NewProducer：", err.Error())
		return nil
	}

	return p
}
