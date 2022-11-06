package talk

import (
	"fmt"
	"strings"

	"go-chat/internal/entity"
	"go-chat/internal/http/internal/dto/web"
	"go-chat/internal/pkg/encrypt"
	"go-chat/internal/pkg/ichat"
	"go-chat/internal/pkg/timeutil"
	"go-chat/internal/repository/cache"
	"go-chat/internal/repository/model"
	"go-chat/internal/service"
)

type Talk struct {
	service            *service.TalkService
	talkListService    *service.TalkSessionService
	redisLock          *cache.RedisLock
	userService        *service.UserService
	wsClient           *cache.WsClientSession
	lastMessage        *cache.MessageStorage
	contactService     *service.ContactService
	unreadTalkCache    *cache.UnreadStorage
	contactRemarkCache *cache.ContactRemark
	groupService       *service.GroupService
	authPermission     *service.AuthPermissionService
}

func NewTalk(service *service.TalkService, talkListService *service.TalkSessionService, redisLock *cache.RedisLock, userService *service.UserService, wsClient *cache.WsClientSession, lastMessage *cache.MessageStorage, contactService *service.ContactService, unreadTalkCache *cache.UnreadStorage, contactRemarkCache *cache.ContactRemark, groupService *service.GroupService, authPermission *service.AuthPermissionService) *Talk {
	return &Talk{service: service, talkListService: talkListService, redisLock: redisLock, userService: userService, wsClient: wsClient, lastMessage: lastMessage, contactService: contactService, unreadTalkCache: unreadTalkCache, contactRemarkCache: contactRemarkCache, groupService: groupService, authPermission: authPermission}
}

// List 会话列表
func (c *Talk) List(ctx *ichat.Context) error {

	uid := ctx.UserId()

	// 获取未读消息数
	unReads := c.unreadTalkCache.GetAll(ctx.RequestCtx(), uid)
	if len(unReads) > 0 {
		c.talkListService.BatchAddList(ctx.RequestCtx(), uid, unReads)
	}

	data, err := c.talkListService.List(ctx.RequestCtx(), uid)
	if err != nil {
		return ctx.BusinessError(err.Error())
	}

	friends := make([]int, 0)
	for _, item := range data {
		if item.TalkType == 1 {
			friends = append(friends, item.ReceiverId)
		}
	}

	// 获取好友备注
	remarks, _ := c.contactService.Dao().Remarks(ctx.RequestCtx(), uid, friends)

	items := make([]*web.TalkListItem, 0)
	for _, item := range data {
		if item.Nickname != "" || item.GroupName != "" {
			value := &web.TalkListItem{
				Id:          int32(item.Id),
				TalkType:    int32(item.TalkType),
				ReceiverId:  int32(item.ReceiverId),
				IsTop:       int32(item.IsTop),
				IsDisturb:   int32(item.IsDisturb),
				IsRobot:     int32(item.IsRobot),
				Avatar:      item.UserAvatar,
				MsgText:     "",
				UpdatedTime: item.UpdatedAt.Unix(),
				UpdatedAt:   timeutil.FormatDatetime(item.UpdatedAt),
			}

			if num, ok := unReads[fmt.Sprintf("%d_%d", item.TalkType, item.ReceiverId)]; ok {
				value.UnreadNum = int32(num)
			}

			if item.TalkType == 1 {
				value.Name = item.Nickname
				value.Avatar = item.UserAvatar
				value.Avatar = item.UserAvatar
				if len(remarks) > 0 {
					value.RemarkName = remarks[item.ReceiverId]
				}
				//value.IsOnline = int32(strutil.BoolToInt(c.wsClient.IsOnline(ctx.Context, entity.ImChannelDefault, strconv.Itoa(int(value.ReceiverId)))))
			} else {
				value.Name = item.GroupName
				value.Avatar = item.GroupAvatar
			}

			// 查询缓存消息
			if msg, err := c.lastMessage.Get(ctx.RequestCtx(), item.TalkType, uid, item.ReceiverId); err == nil {
				value.MsgText = msg.Content
				value.UpdatedAt = msg.Datetime
				value.UpdatedTime = timeutil.ParseDateTime(msg.Datetime).Unix()
			}

			items = append(items, value)
		}
	}
	return ctx.Success(items)
}

// Create 创建会话列表
func (c *Talk) Create(ctx *ichat.Context) error {

	var (
		params = &web.CreateTalkRequest{}
		uid    = ctx.UserId()
		agent  = strings.TrimSpace(ctx.Context.GetHeader("user-agent"))
	)

	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)
	}

	if agent != "" {
		agent = encrypt.Md5(agent)
	}

	// 判断对方是否是自己
	if params.TalkType == entity.ChatPrivateMode && params.ReceiverId == ctx.UserId() {
		return ctx.BusinessError("创建失败")
	}

	key := fmt.Sprintf("talk:list:%d-%d-%d-%s", uid, params.ReceiverId, params.TalkType, agent)
	if !c.redisLock.Lock(ctx.RequestCtx(), key, 10) {
		return ctx.BusinessError("创建失败")
	}

	// 暂无权限
	if !c.authPermission.IsAuth(ctx.RequestCtx(), &service.AuthPermission{
		TalkType:   params.TalkType,
		UserId:     uid,
		ReceiverId: params.ReceiverId,
	}) {
		return ctx.BusinessError("暂无权限")
	}

	result, err := c.talkListService.Create(ctx.RequestCtx(), &model.TalkSessionCreateOpt{
		UserId:     uid,
		TalkType:   params.TalkType,
		ReceiverId: params.ReceiverId,
	})
	if err != nil {
		return ctx.BusinessError(err.Error())
	}

	item := &web.TalkListItem{
		Id:         int32(result.Id),
		TalkType:   int32(result.TalkType),
		ReceiverId: int32(result.ReceiverId),
		IsRobot:    int32(result.IsRobot),
		UpdatedAt:  timeutil.DateTime(),
	}

	if item.TalkType == entity.ChatPrivateMode {
		item.UnreadNum = int32(c.unreadTalkCache.Get(ctx.RequestCtx(), 1, params.ReceiverId, uid))
		item.RemarkName = c.contactService.Dao().GetFriendRemark(ctx.RequestCtx(), uid, params.ReceiverId)

		if user, err := c.userService.Dao().FindById(result.ReceiverId); err == nil {
			item.Name = user.Nickname
			item.Avatar = user.Avatar
		}
	} else if result.TalkType == entity.ChatGroupMode {
		if group, err := c.groupService.Dao().FindById(params.ReceiverId); err == nil {
			item.Name = group.Name
		}
	}

	// 查询缓存消息
	if msg, err := c.lastMessage.Get(ctx.RequestCtx(), result.TalkType, uid, result.ReceiverId); err == nil {
		item.MsgText = msg.Content
		item.UpdatedAt = msg.Datetime
	}

	return ctx.Success(&web.CreateTalkResponse{
		Id:         item.Id,
		TalkType:   item.TalkType,
		ReceiverId: item.ReceiverId,
		IsTop:      item.IsTop,
		IsDisturb:  item.IsDisturb,
		IsOnline:   item.IsOnline,
		IsRobot:    item.IsRobot,
		Name:       item.Name,
		Avatar:     item.Avatar,
		RemarkName: item.RemarkName,
		UnreadNum:  item.UnreadNum,
		MsgText:    item.MsgText,
		UpdatedAt:  item.UpdatedAt,
	})
}

// Delete 删除列表
func (c *Talk) Delete(ctx *ichat.Context) error {

	params := &web.DeleteTalkListRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)
	}

	if err := c.talkListService.Delete(ctx.Context, ctx.UserId(), params.Id); err != nil {
		return ctx.BusinessError(err.Error())
	}

	return ctx.Success(&web.DeleteTalkListResponse{})
}

// Top 置顶列表
func (c *Talk) Top(ctx *ichat.Context) error {

	params := &web.TopTalkListRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)
	}

	if err := c.talkListService.Top(ctx.Context, &service.TalkSessionTopOpt{
		UserId: ctx.UserId(),
		Id:     params.Id,
		Type:   params.Type,
	}); err != nil {
		return ctx.BusinessError(err.Error())
	}

	return ctx.Success(&web.TopTalkListResponse{})
}

// Disturb 会话免打扰
func (c *Talk) Disturb(ctx *ichat.Context) error {

	params := &web.DisturbTalkListRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)
	}

	if err := c.talkListService.Disturb(ctx.Context, &service.TalkSessionDisturbOpt{
		UserId:     ctx.UserId(),
		TalkType:   params.TalkType,
		ReceiverId: params.ReceiverId,
		IsDisturb:  params.IsDisturb,
	}); err != nil {
		return ctx.BusinessError(err.Error())
	}

	return ctx.Success(&web.DisturbTalkListResponse{})
}

func (c *Talk) ClearUnreadMessage(ctx *ichat.Context) error {

	params := &web.ClearTalkUnreadNumRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)
	}

	c.unreadTalkCache.Reset(ctx.RequestCtx(), params.TalkType, params.ReceiverId, ctx.UserId())

	return ctx.Success(&web.ClearTalkUnreadNumResponse{})
}
