package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"unsafe"

	"go-chat/config"
	"go-chat/internal/pkg/logger"
	"go-chat/internal/websocket/internal/process/handle"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

type SubscribeContent struct {
	Event string `json:"event"`
	Data  string `json:"data"`
}

type WsSubscribe struct {
	rds     *redis.Client
	mq      rocketmq.Producer
	conf    *config.Config
	consume *handle.SubscribeConsume
}

func NewWsSubscribe(rds *redis.Client, mq rocketmq.Producer, conf *config.Config, consume *handle.SubscribeConsume) *WsSubscribe {
	return &WsSubscribe{rds: rds, mq: mq, conf: conf, consume: consume}
}

func (w *WsSubscribe) Setup(ctx context.Context) error {

	log.Println("WsSubscribe Setup")
	host := fmt.Sprintf("%s:%s", w.conf.RabbitMQ.Host, strconv.Itoa(w.conf.RabbitMQ.Port))
	c, err := rocketmq.NewPushConsumer(
		// 指定 Group 可以实现消费者负载均衡进行消费，并且保证他们的Topic+Tag要一样。
		// 如果同一个 GroupID 下的不同消费者实例，订阅了不同的 Topic+Tag 将导致在对Topic 的消费队列进行负载均衡的时候产生不正确的结果，最终导致消息丢失。(官方源码设计)
		consumer.WithGroupName(w.conf.RabbitMQ.ExchangeName),
		consumer.WithNameServer([]string{host}),
	)
	if err != nil {
		panic(err)
	}
	err = c.Subscribe(w.conf.RabbitMQ.ExchangeName, consumer.MessageSelector{}, func(ctx context.Context,
		msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		for _, msg := range msgs {
			result := *(*string)(unsafe.Pointer(&msg.Body))
			logrus.Printf("Received a message: %s", result)
			var message *SubscribeContent
			if err := json.Unmarshal(msg.Body, &message); err == nil {
				go func() {
					w.consume.Handle(message.Event, message.Data)
				}()
			} else {
				logger.Warnf("订阅消息格式错误 Err: %s \n", err.Error())
			}
		}
		// 消费成功，进行ack确认
		return consumer.ConsumeSuccess, nil
	})
	if err != nil {
		panic(err)
	}
	err = c.Start()
	if err != nil {
		panic(err)
	}
	defer func() {
		if err != nil {

			fmt.Printf("shutdown mqAdmin error: %s", err.Error())
		}
		err = c.Shutdown()
		if err != nil {

			fmt.Printf("shutdown Consumer error: %s", err.Error())
		}
	}()
	<-(chan interface{})(nil)
	return nil
}
