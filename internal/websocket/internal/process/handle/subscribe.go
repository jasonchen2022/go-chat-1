package handle

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"go-chat/internal/pkg/timeutil"
	"go-chat/internal/repository/cache"
	"go-chat/internal/repository/model"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"

	"go-chat/config"
	"go-chat/internal/entity"
	"go-chat/internal/pkg/im"
	"go-chat/internal/pkg/jsonutil"
	"go-chat/internal/service"
	"go-chat/internal/service/push"
)

type onConsumeFunc func(data string)

type SubscribeConsume struct {
	conf           *config.Config
	rds            *redis.Client
	ws             *cache.WsClientSession
	room           *cache.RoomStorage
	recordsService *service.TalkRecordsService
	contactService *service.ContactService
	userService    *service.UserService
	getuiService   *push.GeTuiService
}

func NewSubscribeConsume(conf *config.Config, rds *redis.Client, ws *cache.WsClientSession, room *cache.RoomStorage, recordsService *service.TalkRecordsService, contactService *service.ContactService, userService *service.UserService, getuiService *push.GeTuiService) *SubscribeConsume {
	return &SubscribeConsume{conf: conf, rds: rds, ws: ws, room: room, recordsService: recordsService, contactService: contactService, userService: userService, getuiService: getuiService}
}

func (s *SubscribeConsume) Handle(event string, data string) {

	handler := make(map[string]onConsumeFunc)

	// 注册消息回调事件
	handler[entity.EventTalk] = s.onConsumeTalk
	handler[entity.EventTalkMuteGroup] = s.onConsumeMuteGroup
	handler[entity.EventTalkUpdateGroup] = s.onConsumeUpdateGroup
	handler[entity.EventTalkKeyboard] = s.onConsumeTalkKeyboard
	handler[entity.EventOffOnline] = s.onConsumeOffOnline
	handler[entity.EventTalkRevoke] = s.onConsumeTalkRevoke
	handler[entity.EventTalkJoinGroup] = s.onConsumeTalkJoinGroup
	handler[entity.EventContactApply] = s.onConsumeContactApply
	handler[entity.EventTalkRead] = s.onConsumeTalkRead

	if f, ok := handler[event]; ok {
		f(data)
	} else {
		logrus.Warnf("Event: [%s]未注册回调方法\n", event)
	}
}

//强制离线消息
func (s *SubscribeConsume) onConsumeOffOnline(body string) {
	var msg struct {
		UserId   int    `json:"user_id"`
		ClientId string `json:"client_id"`
	}
	if err := json.Unmarshal([]byte(body), &msg); err != nil {
		logrus.Error("[SubscribeConsume] onConsumeOffOnline Unmarshal err: ", err.Error())
		return
	}
	s.onSendPrivate(msg.UserId, entity.EventOffOnline, msg)
}

// onConsumeMuteGroup 聊天消息事件
func (s *SubscribeConsume) onConsumeMuteGroup(body string) {
	var msg struct {
		GroupId int   `json:"group_id"`
		IsMute  int64 `json:"is_mute"`
	}
	if err := json.Unmarshal([]byte(body), &msg); err != nil {
		logrus.Error("[SubscribeConsume] onConsumeTalk Unmarshal err: ", err.Error())
		return
	}
	s.onSendGroup(msg.GroupId, entity.EventTalkMuteGroup, msg)
}

// onConsumeUpdateGroup 聊天消息事件
func (s *SubscribeConsume) onConsumeUpdateGroup(body string) {
	var msg struct {
		GroupId int `json:"group_id"`
	}
	if err := json.Unmarshal([]byte(body), &msg); err != nil {
		logrus.Error("[SubscribeConsume] onConsumeTalk Unmarshal err: ", err.Error())
		return
	}
	s.onSendGroup(msg.GroupId, entity.EventTalkUpdateGroup, msg)
}

func (s *SubscribeConsume) onSendGroup(receiverId int, messageType string, msg interface{}) {
	ctx := context.Background()
	cids := make([]int64, 0)
	ids := s.room.All(ctx, &cache.RoomOption{
		Channel:  im.Session.Default.Name(),
		RoomType: entity.RoomImGroup,
		Number:   strconv.Itoa(int(receiverId)),
		Sid:      s.conf.ServerId(),
	})

	cids = append(cids, ids...)
	if len(cids) == 0 {
		return
	}

	c := im.NewSenderContent()
	c.SetReceive(cids...)
	c.SetMessage(&im.Message{
		Event: messageType,
		Content: entity.MapStrAny{
			"data": msg,
		},
	})

	im.Session.Default.Write(c)
}

func (s *SubscribeConsume) onSendPrivate(receiverId int, messageType string, msg interface{}) {
	ctx := context.Background()
	cids := make([]int64, 0)
	ids := s.ws.GetUidFromClientIds(ctx, s.conf.ServerId(), im.Session.Default.Name(), strconv.Itoa(receiverId))
	cids = append(cids, ids...)
	if len(cids) == 0 {
		return
	}
	c := im.NewSenderContent()
	c.SetReceive(cids...)
	c.SetMessage(&im.Message{
		Event: messageType,
		Content: entity.MapStrAny{
			"data": msg,
		},
	})

	im.Session.Default.Write(c)
}

// onConsumeTalk 聊天消息事件
func (s *SubscribeConsume) onConsumeTalk(body string) {

	logrus.Info("收到订阅消息：", time.Now().Unix(), body)
	var msg struct {
		TalkType   int   `json:"talk_type"`
		SenderID   int64 `json:"sender_id"`
		ReceiverID int64 `json:"receiver_id"`
		RecordID   int64 `json:"record_id"`
	}

	if err := json.Unmarshal([]byte(body), &msg); err != nil {
		logrus.Error("[SubscribeConsume] onConsumeTalk Unmarshal err: ", err.Error())
		return
	}

	ctx := context.Background()

	cids := make([]int64, 0)
	if msg.TalkType == entity.ChatPrivateMode {
		for _, val := range [2]int64{msg.SenderID, msg.ReceiverID} {
			ids := s.ws.GetUidFromClientIds(ctx, s.conf.ServerId(), im.Session.Default.Name(), strconv.Itoa(int(val)))

			cids = append(cids, ids...)
		}
	} else if msg.TalkType == entity.ChatGroupMode {
		ids := s.room.All(ctx, &cache.RoomOption{
			Channel:  im.Session.Default.Name(),
			RoomType: entity.RoomImGroup,
			Number:   strconv.Itoa(int(msg.ReceiverID)),
			Sid:      s.conf.ServerId(),
		})
		logrus.Info("获取群ID：", jsonutil.Encode(ids))
		cids = append(cids, ids...)
	}

	data, err := s.recordsService.GetTalkRecord(ctx, msg.RecordID)
	if err != nil {
		logrus.Error("[SubscribeConsume] 读取对话记录失败 err: ", err.Error())
		return
	}

	if len(cids) == 0 {
		return
	}

	// if data.MsgType == 1 {
	// 	// MsgTypeSystemText  = 0  // 系统文本消息
	// 	// MsgTypeText        = 1  // 文本消息
	// 	// MsgTypeFile        = 2  // 文件消息
	// 	// MsgTypeForward     = 3  // 会话消息
	// 	// MsgTypeCode        = 4  // 代码消息
	// 	// MsgTypeVote        = 5  // 投票消息
	// 	// MsgTypeGroupNotice = 6  // 群组公告
	// 	// MsgTypeFriendApply = 7  // 好友申请
	// 	// MsgTypeLogin       = 8  // 登录通知
	// 	// MsgTypeGroupInvite = 9  // 入群退群消息
	// 	// MsgTypeLocation    = 10 // 位置消息
	// 	// MsgTypeRedPackets  = 11 // 红包

	// 	clientId := "4efe67a49be747d7770f997031329e2c"
	// 	//私聊走单推通道
	// 	if msg.TalkType == 1 {
	// 		//clientId, err := s.userService.Dao().GetClientId(int(msg.ReceiverID))
	// 		// if err != nil {
	// 		// 	logrus.Error("[获取ClientId] 失败 err: ", err.Error())
	// 		// }

	// 		s.getuiService.PushSingleByCid(ctx, clientId, "私聊消息标题", "私聊消息body")
	// 	}
	// 	if msg.TalkType == 2 {

	// 		s.getuiService.PushSingleByCids(ctx, []string{clientId}, "群聊消息标题", "群聊消息body")

	// 	}

	// }

	c := im.NewSenderContent()
	c.SetReceive(cids...)
	c.SetMessage(&im.Message{
		Event: entity.EventTalk,
		Content: entity.MapStrAny{
			"sender_id":   msg.SenderID,
			"receiver_id": msg.ReceiverID,
			"talk_type":   msg.TalkType,
			"data":        data,
		},
	})

	im.Session.Default.Write(c)
	logrus.Info("结束订阅消息：", time.Now().Unix())
}

// onConsumeTalkKeyboard 键盘输入事件消息
func (s *SubscribeConsume) onConsumeTalkKeyboard(body string) {
	var msg struct {
		SenderID   int `json:"sender_id"`
		ReceiverID int `json:"receiver_id"`
	}

	if err := json.Unmarshal([]byte(body), &msg); err != nil {
		logrus.Error("[SubscribeConsume] onConsumeTalkKeyboard Unmarshal err: ", err.Error())
		return
	}

	cids := s.ws.GetUidFromClientIds(context.Background(), s.conf.ServerId(), im.Session.Default.Name(), strconv.Itoa(msg.ReceiverID))

	if len(cids) == 0 {
		return
	}

	c := im.NewSenderContent()
	c.SetReceive(cids...)
	c.SetMessage(&im.Message{
		Event: entity.EventTalkKeyboard,
		Content: entity.MapStrAny{
			"sender_id":   msg.SenderID,
			"receiver_id": msg.ReceiverID,
		},
	})

	im.Session.Default.Write(c)
}

// // onConsumeLogin 用户上线或下线消息
// func (s *SubscribeConsume) onConsumeLogin(body string) {
// 	var msg struct {
// 		Status int `json:"status"`
// 		UserID int `json:"user_id"`
// 	}

// 	if err := json.Unmarshal([]byte(body), &msg); err != nil {
// 		logrus.Error("[SubscribeConsume] onConsumeLogin Unmarshal err: ", err.Error())
// 		return
// 	}

// 	ctx := context.Background()
// 	cids := make([]int64, 0)

// 	uids := s.contactService.GetContactIds(ctx, msg.UserID)
// 	sid := s.conf.ServerId()
// 	for _, uid := range uids {
// 		ids := s.ws.GetUidFromClientIds(ctx, sid, im.Session.Default.Name(), fmt.Sprintf("%d", uid))

// 		cids = append(cids, ids...)
// 	}

// 	if len(cids) == 0 {
// 		return
// 	}

// 	c := im.NewSenderContent()
// 	c.SetReceive(cids...)
// 	c.SetMessage(&im.Message{
// 		Event:   entity.EventOnlineStatus,
// 		Content: msg,
// 	})

// 	im.Session.Default.Write(c)
// }

// onConsumeTalkRevoke 撤销聊天消息
func (s *SubscribeConsume) onConsumeTalkRevoke(body string) {
	var (
		msg struct {
			RecordId int `json:"record_id"`
		}
		record *model.TalkRecords
		ctx    = context.Background()
	)

	if err := jsonutil.Decode(body, &msg); err != nil {
		logrus.Error("[SubscribeConsume] onConsumeTalkRevoke Unmarshal err: ", err.Error())
		return
	}

	if err := s.recordsService.Db().First(&record, msg.RecordId).Error; err != nil {
		return
	}

	cids := make([]int64, 0)
	if record.TalkType == entity.ChatPrivateMode {
		for _, uid := range [2]int{record.UserId, record.ReceiverId} {
			ids := s.ws.GetUidFromClientIds(ctx, s.conf.ServerId(), im.Session.Default.Name(), strconv.Itoa(uid))
			cids = append(cids, ids...)
		}
	} else if record.TalkType == entity.ChatGroupMode {
		cids = s.room.All(ctx, &cache.RoomOption{
			Channel:  im.Session.Default.Name(),
			RoomType: entity.RoomImGroup,
			Number:   strconv.Itoa(record.ReceiverId),
			Sid:      s.conf.ServerId(),
		})
	}

	if len(cids) == 0 {
		return
	}

	c := im.NewSenderContent()
	c.SetReceive(cids...)
	c.SetMessage(&im.Message{
		Event: entity.EventTalkRevoke,
		Content: entity.MapStrAny{
			"talk_type":   record.TalkType,
			"sender_id":   record.UserId,
			"receiver_id": record.ReceiverId,
			"record_id":   record.Id,
		},
	})

	im.Session.Default.Write(c)
}

// nolint onConsumeContactApply 好友申请消息
func (s *SubscribeConsume) onConsumeContactApply(body string) {
	var (
		msg struct {
			ApplId int `json:"apply_id"`
			Type   int `json:"type"`
		}
		ctx = context.Background()
	)

	if err := jsonutil.Decode(body, &msg); err != nil {
		logrus.Error("[SubscribeConsume] onConsumeContactApply Unmarshal err: ", err.Error())
		return
	}

	apply := &model.ContactApply{}
	if err := s.contactService.Db().First(&apply, msg.ApplId).Error; err != nil {
		return
	}

	cids := s.ws.GetUidFromClientIds(ctx, s.conf.ServerId(), im.Session.Default.Name(), strconv.Itoa(apply.FriendId))
	if len(cids) == 0 {
		return
	}

	user := &model.Users{}
	if err := s.contactService.Db().First(&user, apply.FriendId).Error; err != nil {
		return
	}

	data := entity.MapStrAny{}
	data["sender_id"] = apply.UserId
	data["receiver_id"] = apply.FriendId
	data["remark"] = apply.Remark
	data["friend"] = entity.MapStrAny{
		"nickname":   user.Nickname,
		"remark":     apply.Remark,
		"created_at": timeutil.FormatDatetime(apply.CreatedAt),
	}

	c := im.NewSenderContent()
	c.SetReceive(cids...)
	c.SetMessage(&im.Message{
		Event:   entity.EventContactApply,
		Content: data,
	})

	im.Session.Default.Write(c)
}

// onConsumeTalkJoinGroup 加入群房间
func (s *SubscribeConsume) onConsumeTalkJoinGroup(body string) {
	var (
		ctx  = context.Background()
		sid  = s.conf.ServerId()
		data struct {
			Gid  int   `json:"group_id"`
			Type int   `json:"type"`
			Uids []int `json:"uids"`
		}
	)

	if err := json.Unmarshal([]byte(body), &data); err != nil {
		logrus.Error("[SubscribeConsume] onConsumeTalkJoinGroup Unmarshal err: ", err.Error())
		return
	}

	for _, uid := range data.Uids {
		cids := s.ws.GetUidFromClientIds(ctx, sid, im.Session.Default.Name(), strconv.Itoa(uid))

		for _, cid := range cids {
			opts := &cache.RoomOption{
				Channel:  im.Session.Default.Name(),
				RoomType: entity.RoomImGroup,
				Number:   strconv.Itoa(data.Gid),
				Sid:      s.conf.ServerId(),
				Cid:      cid,
			}

			if data.Type == 2 {
				_ = s.room.Del(ctx, opts)
			} else {
				_ = s.room.Add(ctx, opts)
			}
		}
	}
}

// onConsumeTalkRead 消息已读事件
func (s *SubscribeConsume) onConsumeTalkRead(body string) {
	var (
		ctx  = context.Background()
		sid  = s.conf.ServerId()
		data struct {
			SenderId   int   `json:"sender_id"`
			ReceiverId int   `json:"receiver_id"`
			Ids        []int `json:"ids"`
		}
	)

	if err := jsonutil.Decode(body, &data); err != nil {
		logrus.Error("[SubscribeConsume] onConsumeContactApply Unmarshal err: ", err.Error())
		return
	}

	cids := s.ws.GetUidFromClientIds(ctx, sid, im.Session.Default.Name(), fmt.Sprintf("%d", data.ReceiverId))

	c := im.NewSenderContent()
	c.SetReceive(cids...)
	c.SetMessage(&im.Message{
		Event: entity.EventTalkRead,
		Content: entity.MapStrAny{
			"sender_id":   data.SenderId,
			"receiver_id": data.ReceiverId,
			"ids":         data.Ids,
		},
	})

	im.Session.Default.Write(c)
}
