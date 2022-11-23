package handler

import (
	"context"
	"log"
	"strconv"

	"go-chat/config"
	"go-chat/internal/entity"
	"go-chat/internal/pkg/ichat"
	"go-chat/internal/pkg/im"
	"go-chat/internal/repository/cache"
	"go-chat/internal/service"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

type DefaultWebSocket struct {
	rds                *redis.Client
	mq                 rocketmq.Producer
	config             *config.Config
	cache              *service.ClientService
	room               *cache.RoomStorage
	groupMemberService *service.GroupMemberService
	talkMessage        *service.TalkMessageService
}

func NewDefaultWebSocket(rds *redis.Client, mq rocketmq.Producer, config *config.Config, cache *service.ClientService, room *cache.RoomStorage, groupMemberService *service.GroupMemberService, talkMessage *service.TalkMessageService) *DefaultWebSocket {
	return &DefaultWebSocket{rds: rds, config: config, cache: cache, room: room, groupMemberService: groupMemberService, talkMessage: talkMessage}
}

// Connect 初始化连接
func (c *DefaultWebSocket) Connect(ctx *ichat.Context) error {
	conn, err := im.NewConnect(ctx.Context)
	if err != nil {
		logrus.Errorf("websocket connect error: %s", err.Error())
		return nil
	}

	// 创建客户端
	im.NewClient(ctx.RequestCtx(), conn, &im.ClientOptions{
		Uid:     ctx.UserId(),
		Channel: im.Session.Default,
		Storage: c.cache,
		Buffer:  10,
	}, im.NewClientCallback(
		// 连接成功回调
		im.WithOpenCallback(func(client im.IClient) {
			c.open(client)
		}),
		// 接收消息回调
		im.WithMessageCallback(func(client im.IClient, message []byte) {
			c.message(ctx, client, message)
		}),
		// 关闭连接回调
		im.WithCloseCallback(func(client im.IClient, code int, text string) {
			c.close(client, code, text)
			// fmt.Printf("客户端[%d] 已关闭连接，关闭提示【%d】%s \n", client.ClientId(), code, text)
		}),
	))

	return nil
}

// 连接成功回调事件
func (c *DefaultWebSocket) open(client im.IClient) {

	// 1.查询用户群列表
	ids := c.groupMemberService.Dao().GetUserGroupIds(client.ClientUid())
	// 2.客户端加入群房间
	for _, id := range ids {
		_ = c.room.Add(context.Background(), &cache.RoomOption{
			Channel:  im.Session.Default.Name(),
			RoomType: entity.RoomImGroup,
			Number:   strconv.Itoa(id),
			Sid:      c.config.ServerId(),
			Cid:      client.ClientId(),
		})
	}

	// // 推送上线消息
	// c.rds.Publish(context.Background(), entity.IMGatewayAll, jsonutil.Encode(entity.MapStrAny{
	// 	"event": entity.EventOnlineStatus,
	// 	"data": jsonutil.Encode(entity.MapStrAny{
	// 		"user_id": client.ClientUid(),
	// 		"status":  1,
	// 	}),
	// }))
}

// 消息接收回调事件
func (c *DefaultWebSocket) message(ctx *ichat.Context, client im.IClient, message []byte) {

	// content := string(message)

	// event := gjson.Get(content, "event").String()

	// // 创建一个Channel
	if c.mq == nil {
		// conf := config.ReadConfig(config.ParseConfigArg())
		// c.mq = provider.NewRocketMQClient(ctx.Context, conf)
		log.Println("Failed to open a channel:", "并重新初始化")
	}
	// channel, err := c.mq.Channel()
	// if err != nil {
	// 	log.Println("Failed to open a channel:", err.Error())

	// }
	// defer channel.Close()

	// // 声明exchange
	// if err := channel.ExchangeDeclare(
	// 	c.config.RabbitMQ.ExchangeName, //name
	// 	"fanout",                       //exchangeType
	// 	true,                           //durable
	// 	false,                          //auto-deleted
	// 	false,                          //internal
	// 	false,                          //noWait
	// 	nil,                            //arguments
	// ); err != nil {
	// 	log.Println("Failed to declare a exchange:", err.Error())
	// }

	// switch event {

	// // 对话键盘事件
	// case entity.EventTalkKeyboard:
	// 	var m *dto.KeyboardMessage
	// 	if err := json.Unmarshal(message, &m); err == nil {
	// 		body := entity.MapStrAny{
	// 			"event": entity.EventTalkKeyboard,
	// 			"data": jsonutil.Encode(entity.MapStrAny{
	// 				"sender_id":   m.Data.SenderID,
	// 				"receiver_id": m.Data.ReceiverID,
	// 			}),
	// 		}
	// 		c.talkMessage.SendAll(channel, jsonutil.Encode(body))
	// 	}

	// // 对话消息读事件
	// case entity.EventTalkRead:
	// 	var m *dto.TalkReadMessage
	// 	if err := json.Unmarshal(message, &m); err == nil {
	// 		c.groupMemberService.Db().Model(&model.TalkRecords{}).Where("id in ? and receiver_id = ? and is_read = 0", m.Data.MsgIds, client.ClientUid()).Update("is_read", 1)

	// 		body := entity.MapStrAny{
	// 			"event": entity.EventTalkRead,
	// 			"data": jsonutil.Encode(entity.MapStrAny{
	// 				"sender_id":   client.ClientUid(),
	// 				"receiver_id": m.Data.ReceiverId,
	// 				"ids":         m.Data.MsgIds,
	// 			}),
	// 		}

	// 		c.talkMessage.SendAll(channel, jsonutil.Encode(body))
	// 	}
	// default:
	// 	fmt.Printf("消息事件未定义%s", event)
	// }
}

// 客户端关闭回调事件
func (c *DefaultWebSocket) close(client im.IClient, code int, text string) {

	// 1.判断用户是否是多点登录

	// 2.查询用户群列表
	ids := c.groupMemberService.Dao().GetUserGroupIds(client.ClientUid())

	// 3.客户端退出群房间
	for _, id := range ids {
		_ = c.room.Del(context.Background(), &cache.RoomOption{
			Channel:  im.Session.Default.Name(),
			RoomType: entity.RoomImGroup,
			Number:   strconv.Itoa(id),
			Sid:      c.config.ServerId(),
			Cid:      client.ClientId(),
		})
	}

	// // 推送下线消息
	// c.rds.Publish(context.Background(), entity.IMGatewayAll, jsonutil.Encode(entity.MapStrAny{
	// 	"event": entity.EventOnlineStatus,
	// 	"data": jsonutil.Encode(entity.MapStrAny{
	// 		"user_id": client.ClientUid(),
	// 		"status":  0,
	// 	}),
	// }))
}
