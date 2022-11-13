package handle

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"go-chat/internal/pkg/timeutil"
	"go-chat/internal/repository/cache"
	"go-chat/internal/repository/dao"
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
	conf               *config.Config
	rds                *redis.Client
	ws                 *cache.WsClientSession
	room               *cache.RoomStorage
	sidServer          *cache.SidServer
	groupMemberDao     *dao.GroupMemberDao
	recordsService     *service.TalkRecordsService
	contactService     *service.ContactService
	userService        *service.UserService
	talkSessionService *service.TalkSessionService
	getuiService       *push.GeTuiService
	jpushService       *push.JpushService
}

func NewSubscribeConsume(conf *config.Config, rds *redis.Client, ws *cache.WsClientSession, room *cache.RoomStorage, sidServer *cache.SidServer, groupMemberDao *dao.GroupMemberDao, recordsService *service.TalkRecordsService, contactService *service.ContactService, userService *service.UserService, talkSessionService *service.TalkSessionService, getuiService *push.GeTuiService, jpushService *push.JpushService) *SubscribeConsume {
	return &SubscribeConsume{conf: conf, rds: rds, ws: ws, room: room, sidServer: sidServer, groupMemberDao: groupMemberDao, recordsService: recordsService, contactService: contactService, userService: userService, talkSessionService: talkSessionService, getuiService: getuiService, jpushService: jpushService}
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

	if s.conf.GetEnv() == "alone" {
		//异步推送 文本消息、转发消息、文件消息、红包消息
		if data.MsgType == 1 || data.MsgType == 2 || data.MsgType == 3 || data.MsgType == 11 {
			//判断当前会话是否免打扰
			if !s.talkSessionService.Dao().IsDisturb(data.UserId, data.ReceiverId, data.TalkType) {
				go func() {
					clientIds := make([]string, 0)
					if msg.TalkType == 1 {
						clientId, _ := s.userService.Dao().GetClientId(data.ReceiverId)
						if clientId != "" {
							clientIds = append(clientIds, clientId)
						}
					}
					if msg.TalkType == 2 {
						//offlineIds := make([]int, 0)
						userIds := s.groupMemberDao.GetMemberIds(data.ReceiverId)
						// for _, userId := range userIds {
						// 	is_online := s.isOnline(ctx, userId)
						// 	if !is_online {
						// 		offlineIds = append(offlineIds, userId)
						// 	}
						// }
						//clientIds, _ = s.userService.Dao().GetClientIds(offlineIds)
						clientIds, _ = s.userService.Dao().GetClientIds(userIds)

					}
					logrus.Info("clientIds：", jsonutil.Encode(clientIds))
					//推送离线消息
					if len(clientIds) > 0 {
						//文本消息、转发消息
						if data.MsgType == 1 || data.MsgType == 3 {
							s.pushMessage(ctx, msg.TalkType, data.ReceiverId, clientIds, data.Nickname, data.GroupName, data.Content)
						}
						//文件消息
						if data.MsgType == 2 {
							s.pushMessage(ctx, msg.TalkType, data.ReceiverId, clientIds, data.Nickname, data.GroupName, "图文消息")
						}
						//红包消息
						if data.MsgType == 11 {
							s.pushMessage(ctx, msg.TalkType, data.ReceiverId, clientIds, data.Nickname, data.GroupName, "红包消息")
						}
					}
				}()
			}

		}
	}
	logrus.Info("开始推送cids：", jsonutil.Encode(cids))
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
	logrus.Info("结束推送消息：", time.Now().Unix())
}

//判断目标用户是否在线
// func (s *SubscribeConsume) isOnline(ctx context.Context, receiverId int) bool {
// 	sids := s.sidServer.All(ctx, 1)
// 	is_online := false
// 	for _, sid := range sids {
// 		if s.ws.IsCurrentServerOnline(ctx, sid, entity.ImChannelDefault, strconv.Itoa(receiverId)) {
// 			is_online = true
// 		}
// 	}
// 	logrus.Info(strconv.Itoa(receiverId), "：用户在线状态：", strconv.FormatBool(is_online))
// 	return is_online
// }

//推送离线消息
func (s *SubscribeConsume) pushMessage(ctx context.Context, talkType int, receiverId int, clientIds []string, userName string, groupName, body string) {
	if talkType == 1 {
		// is_online := s.isOnline(ctx, receiverId)
		// if !is_online {
		msgId, _ := s.jpushService.PushMessageByCid(clientIds, userName, body)
		logrus.Info("私聊推送结果：", msgId)
		//}
	}
	if talkType == 2 {
		if len(clientIds) > 0 {
			msgId, _ := s.jpushService.PushMessageByCid(clientIds, groupName, fmt.Sprintf("%s：%s", userName, body))
			logrus.Info("群聊推送结果：", msgId)
		}
	}
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
