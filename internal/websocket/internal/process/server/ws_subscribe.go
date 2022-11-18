package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"unsafe"

	"go-chat/config"
	"go-chat/internal/entity"
	"go-chat/internal/pkg/logger"
	"go-chat/internal/pkg/worker"
	"go-chat/internal/provider"
	"go-chat/internal/websocket/internal/process/handle"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

type SubscribeContent struct {
	Event string `json:"event"`
	Data  string `json:"data"`
}

type WsSubscribe struct {
	rds     *redis.Client
	mq      *amqp.Connection
	conf    *config.Config
	consume *handle.SubscribeConsume
}

func NewWsSubscribe(rds *redis.Client, mq *amqp.Connection, conf *config.Config, consume *handle.SubscribeConsume) *WsSubscribe {
	return &WsSubscribe{rds: rds, mq: mq, conf: conf, consume: consume}
}

func (w *WsSubscribe) Setup(ctx context.Context) error {

	log.Println("WsSubscribe Setup")

	gateway := fmt.Sprintf(entity.IMGatewayPrivate, w.conf.ServerId())

	if w.mq == nil {
		conf := config.ReadConfig(config.ParseConfigArg())
		w.mq = provider.NewRabbitMQClient(ctx, conf)
		log.Println("Failed to open a channel:", "并重新初始化")
	}
	ch, err := w.mq.Channel()
	if err != nil {
		log.Println("Failed to open a channel:", err.Error())
		return err
	}

	// 声明一个群聊队列
	qGroup, err := ch.QueueDeclare(
		entity.IMGatewayAll, // name
		true,                // durable
		false,               // delete when usused
		false,               // exclusive
		false,               // no-wait
		nil,                 // arguments
	)
	if err != nil {
		log.Println("Failed to open a channel:", err.Error())
		return err
	}
	ch2, err := w.mq.Channel()
	if err != nil {
		log.Println("Failed to open a channel2:", err.Error())
		return err
	}
	// 声明一个私聊队列
	qPrivate, err := ch2.QueueDeclare(
		gateway, // name
		true,    // durable
		false,   // delete when usused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		log.Println("Failed to open a channel:", err.Error())
		return err
	}
	// 注册群聊消费者
	msgsGroup, err := ch.Consume(
		qGroup.Name, // queue
		"",          // 标签
		true,        // auto-ack
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
	if err != nil {
		log.Println("Failed to open a channel:", err.Error())
		return err
	}
	//注册私聊消费者
	msgsPrivate, err := ch2.Consume(
		qPrivate.Name, // queue
		"",            // 标签
		true,          // auto-ack
		false,         // exclusive
		false,         // no-local
		false,         // no-wait
		nil,           // args
	)
	if err != nil {
		log.Println("Failed to open a channel:", err.Error())
		return err
	}

	go func() {
		work := worker.NewWorker(20, 20)

		for d := range msgsGroup {
			work.Do(func() {
				result := *(*string)(unsafe.Pointer(&d.Body))
				logrus.Printf("Received a message: %s", result)
				var message *SubscribeContent
				if err := json.Unmarshal(d.Body, &message); err == nil {
					w.consume.Handle(message.Event, message.Data)
				} else {
					logger.Warnf("订阅消息格式错误 Err: %s \n", err.Error())
				}
			})
		}
		for d := range msgsPrivate {
			work.Do(func() {
				result := *(*string)(unsafe.Pointer(&d.Body))
				logrus.Printf("Received a message2: %s", result)
				var message *SubscribeContent
				if err := json.Unmarshal(d.Body, &message); err == nil {
					w.consume.Handle(message.Event, message.Data)
				} else {
					logger.Warnf("订阅消息格式错误 Err: %s \n", err.Error())
				}
			})
		}

		work.Wait()
	}()

	<-ctx.Done()

	return nil
}
