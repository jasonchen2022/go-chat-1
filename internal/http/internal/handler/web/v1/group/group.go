package group

import (
	"fmt"
	"log"
	"time"

	"go-chat/config"
	"go-chat/internal/entity"
	"go-chat/internal/http/internal/dto/web"
	"go-chat/internal/pkg/encrypt"
	"go-chat/internal/pkg/ichat"
	"go-chat/internal/pkg/jsonutil"
	"go-chat/internal/pkg/logger"
	"go-chat/internal/pkg/sliceutil"
	"go-chat/internal/pkg/timeutil"
	"go-chat/internal/repository/cache"
	"go-chat/internal/repository/model"
	"go-chat/internal/service"

	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

type Group struct {
	service            *service.GroupService
	groupMemberService *service.GroupMemberService
	talkSessionService *service.TalkSessionService
	userService        *service.UserService
	redisLock          *cache.RedisLock
	contactService     *service.ContactService
	groupNoticeService *service.GroupNoticeService
	messageService     *service.TalkMessageService
	memberService      *service.MemberService
	mq                 *amqp.Connection
	config             *config.Config
}

func NewGroup(
	service *service.GroupService,
	groupMemberService *service.GroupMemberService,
	talkSessionService *service.TalkSessionService,
	redisLock *cache.RedisLock,
	contactService *service.ContactService,
	userService *service.UserService,
	groupNoticeService *service.GroupNoticeService,
	messageService *service.TalkMessageService,
	memberService *service.MemberService,
	mq *amqp.Connection,
	config *config.Config,
) *Group {
	return &Group{
		service:            service,
		groupMemberService: groupMemberService,
		talkSessionService: talkSessionService,
		redisLock:          redisLock,
		contactService:     contactService,
		userService:        userService,
		groupNoticeService: groupNoticeService,
		messageService:     messageService,
		memberService:      memberService,
		mq:                 mq,
		config:             config,
	}
}

// Create 创建群聊分组
func (c *Group) Create(ctx *ichat.Context) error {

	params := &web.GroupCreateRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)
	}

	gid, err := c.service.Create(ctx.RequestCtx(), &model.CreateGroupOpts{
		UserId:    ctx.UserId(),
		Name:      params.Name,
		Avatar:    params.Avatar,
		Profile:   params.Profile,
		MemberIds: sliceutil.ParseIds(params.MembersIds),
	})
	if err != nil {
		return ctx.BusinessError("创建群聊失败，请稍后再试！")
	}

	return ctx.Success(entity.H{
		"group_id": gid,
	})
}

// Create 创建聊天室分组
func (c *Group) CreateChat(ctx *ichat.Context) error {
	params := &web.GroupCreateChatRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)

	}
	members, err := c.memberService.FindAdmin()
	if err != nil {
		return ctx.BusinessError("获取管理员账号列表失败")

	}
	var userIds []int
	for _, member := range members {
		userIds = append(userIds, member.Id)
	}
	if len(userIds) > 0 {
		users, err := c.userService.Dao().FindByIds(userIds)
		if err != nil {
			return ctx.BusinessError("创建管理员账号失败")
		}
		password, _ := encrypt.HashPassword("12345689")
		for _, member := range members {
			isExit := false
			for _, user := range users {
				if member.Id == user.Id {
					isExit = true
				}
			}
			if !isExit {
				_, err := c.userService.Dao().Create(&model.Users{
					Id:        member.Id,
					Nickname:  member.UserName,
					Mobile:    member.Mobile,
					Avatar:    member.Avatar,
					Gender:    member.Gender,
					Type:      member.Type,
					Motto:     member.Motto,
					Password:  password,
					CreatedAt: time.Now(),
				})
				if err != nil {
					logrus.Error(err)
				}
			}

		}
	}

	gid, err := c.service.Create(ctx.RequestCtx(), &model.CreateGroupOpts{
		UserId:    params.AnchorId,
		Name:      params.Name,
		Profile:   params.Profile,
		Type:      3, //默认聊天室
		MemberIds: userIds,
	})
	if err != nil {
		fmt.Printf("创建聊天室出错：%s", err.Error())
		return ctx.BusinessError("创建群聊失败，请稍后再试！")
	}

	return ctx.Success(entity.H{
		"group_id": gid,
	})
}

// Dismiss 解散群组
func (c *Group) Dismiss(ctx *ichat.Context) error {

	params := &web.GroupDismissRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)
	}

	uid := ctx.UserId()
	if !c.groupMemberService.Dao().IsMaster(params.GroupId, uid) {
		return ctx.BusinessError("暂无权限解散群组！")
	}

	if err := c.service.Dismiss(ctx.RequestCtx(), params.GroupId, ctx.UserId()); err != nil {
		return ctx.BusinessError("群组解散失败！")
	}

	_ = c.messageService.SendSysMessage(ctx.Context, &service.SysTextMessageOpt{
		UserId:     uid,
		TalkType:   entity.ChatGroupMode,
		ReceiverId: params.GroupId,
		Text:       "群组已被群主或管理员解散！",
	})

	return ctx.Success(nil)
}

// Invite 邀请好友加入群聊
func (c *Group) Invite(ctx *ichat.Context) error {

	params := &web.GroupInviteRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)
	}

	key := fmt.Sprintf("group-join:%d", params.GroupId)
	if !c.redisLock.Lock(ctx.Context, key, 20) {
		return ctx.BusinessError("网络异常，请稍后再试！")
	}

	defer c.redisLock.UnLock(ctx.Context, key)

	uid := ctx.UserId()
	uids := sliceutil.UniqueInt(sliceutil.ParseIds(params.Ids))

	if len(uids) == 0 {
		return ctx.BusinessError("邀请好友列表不能为空！")
	}

	if !c.groupMemberService.Dao().IsMember(params.GroupId, uid, true) {
		return ctx.BusinessError("非群组成员，无权邀请好友！")
	}

	if err := c.service.InviteMembers(ctx.Context, &model.InviteGroupMembersOpt{
		UserId:    uid,
		GroupId:   params.GroupId,
		MemberIds: uids,
	}); err != nil {
		return ctx.BusinessError(err.Error())
	}

	return ctx.Success(nil)
}

// Join 加入群聊
func (c *Group) Join(ctx *ichat.Context) error {
	params := &web.GroupCommonRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)

	}

	// key := fmt.Sprintf("group-join-chat:%d", params.GroupId)
	// if !c.redisLock.Lock(ctx.Context, key, 20) {
	// 	return ctx.BusinessError("网络异常，请稍后再试！")
	// }

	// defer c.redisLock.UnLock(ctx.Context, key)

	uid := ctx.UserId()
	ids := []int{uid}
	if err := c.service.InviteMembers(ctx.Context, &model.InviteGroupMembersOpt{
		UserId:    1, //默认为管理员邀请
		GroupId:   params.GroupId,
		MemberIds: ids,
	}); err != nil {
		fmt.Printf("进群失败：%s", err.Error())
		return ctx.BusinessError(err.Error())
	} else {
		return ctx.Success(nil)
	}
}

// SignOut 退出群聊
func (c *Group) SignOut(ctx *ichat.Context) error {

	params := &web.GroupSecedeRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)
	}

	uid := ctx.UserId()
	if err := c.service.Secede(ctx.RequestCtx(), params.GroupId, uid); err != nil {
		return ctx.BusinessError(err.Error())
	}

	// 删除聊天会话
	sid := c.talkSessionService.Dao().FindBySessionId(uid, params.GroupId, entity.ChatGroupMode)
	_ = c.talkSessionService.Delete(ctx.Context, ctx.UserId(), sid)

	return ctx.Success(nil)
}

// Setting 群设置接口（预留）
func (c *Group) Setting(ctx *ichat.Context) error {

	params := &web.GroupSettingRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)
	}

	uid := ctx.UserId()
	if !c.groupMemberService.Dao().IsLeader(params.GroupId, uid) {
		return ctx.BusinessError("无权限操作")
	}

	if err := c.service.Update(ctx.RequestCtx(), &model.UpdateGroupOpt{
		GroupId: params.GroupId,
		Name:    params.GroupName,
		Avatar:  params.Avatar,
		Profile: params.Profile,
	}); err != nil {
		return ctx.BusinessError(err.Error())
	}

	_ = c.messageService.SendSysMessage(ctx.Context, &service.SysTextMessageOpt{
		UserId:     uid,
		TalkType:   entity.ChatGroupMode,
		ReceiverId: params.GroupId,
		Text:       "群主或管理员修改了群信息！",
	})

	return ctx.Success(nil)
}

// RemoveMembers 移除指定成员(群组&管理员权限)
func (c *Group) RemoveMembers(ctx *ichat.Context) error {

	params := &web.GroupRemoveMembersRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)
	}

	uid := ctx.UserId()

	if !c.groupMemberService.Dao().IsLeader(params.GroupId, uid) {
		return ctx.BusinessError("无权限操作")
	}

	err := c.service.RemoveMembers(ctx.RequestCtx(), &service.RemoveMembersOpt{
		UserId:    uid,
		GroupId:   params.GroupId,
		MemberIds: sliceutil.ParseIds(params.MembersIds),
	})

	if err != nil {
		return ctx.BusinessError(err.Error())
	}

	return ctx.Success(nil)
}

// Detail 获取群组信息
func (c *Group) Detail(ctx *ichat.Context) error {

	params := &web.GroupCommonRequest{}
	if err := ctx.Context.ShouldBindQuery(params); err != nil {
		return ctx.InvalidParams(err)
	}

	uid := ctx.UserId()

	groupInfo, err := c.service.Dao().FindById(params.GroupId)
	if err != nil {
		return ctx.BusinessError(err.Error())
	}

	if groupInfo.Id == 0 {
		return ctx.BusinessError("数据不存在")
	}

	info := entity.H{}
	info["group_id"] = groupInfo.Id
	info["group_name"] = groupInfo.Name
	info["profile"] = groupInfo.Profile
	info["avatar"] = groupInfo.Avatar
	info["created_at"] = timeutil.FormatDatetime(groupInfo.CreatedAt)
	info["is_manager"] = uid == groupInfo.CreatorId
	info["manager_nickname"] = ""
	info["visit_card"] = c.groupMemberService.Dao().GetMemberRemark(params.GroupId, uid)
	info["is_disturb"] = 0
	info["is_top"] = 0
	info["is_show_nickname"] = 0
	info["is_mute"] = groupInfo.IsMute
	info["is_overt"] = groupInfo.IsOvert
	info["notice"] = entity.H{}

	if notice, _ := c.groupNoticeService.Dao().GetLatestNotice(ctx.Context, params.GroupId); err == nil {
		info["notice"] = notice
	}

	if c.talkSessionService.Dao().IsDisturb(uid, groupInfo.Id, 2) {
		info["is_disturb"] = 1
	}

	if userInfo, err := c.userService.Dao().FindById(uid); err == nil {
		info["manager_nickname"] = userInfo.Nickname
	}

	//群成员
	members := c.groupMemberService.Dao().GetMembers(params.GroupId)
	info["members"] = members

	//会话
	talkSesstion, err := c.talkSessionService.FindTalkSession(ctx.Context, params.GroupId, uid)
	if err == nil {
		info["is_top"] = talkSesstion.IsTop
		info["is_disturb"] = talkSesstion.IsDisturb
		info["is_show_nickname"] = talkSesstion.IsShowNickname
	}

	return ctx.Success(info)
}

// EditRemark 修改群备注接口
func (c *Group) EditRemark(ctx *ichat.Context) error {

	params := &web.GroupEditRemarkRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)
	}

	if err := c.groupMemberService.CardEdit(params.GroupId, ctx.UserId(), params.VisitCard); err != nil {
		return ctx.BusinessError("修改群备注失败！")
	}

	return ctx.Success(nil)
}

func (c *Group) GetInviteFriends(ctx *ichat.Context) error {

	params := &web.GetInviteFriendsRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)
	}

	items, err := c.contactService.List(ctx.Context, ctx.UserId())
	if err != nil {
		return ctx.BusinessError(err.Error())
	}

	if params.GroupId <= 0 {
		return ctx.Success(items)
	}

	mids := c.groupMemberService.Dao().GetMemberIds(params.GroupId)
	if len(mids) == 0 {
		return ctx.Success(items)
	}

	data := make([]*model.ContactListItem, 0)
	for i := 0; i < len(items); i++ {
		if !sliceutil.InInt(items[i].Id, mids) {
			data = append(data, items[i])
		}
	}

	return ctx.Success(data)
}

func (c *Group) Groups(ctx *ichat.Context) error {

	items, err := c.service.List(ctx.UserId())
	if err != nil {
		return ctx.BusinessError(err.Error())
	}

	return ctx.Success(entity.H{
		"rows": items,
	})
}

// Members 获取群成员列表
func (c *Group) Members(ctx *ichat.Context) error {

	params := &web.GroupCommonRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)
	}

	if !c.groupMemberService.Dao().IsMember(params.GroupId, ctx.UserId(), false) {
		return ctx.BusinessError("非群成员无权查看成员列表！")
	}

	return ctx.Success(c.groupMemberService.Dao().GetMembers(params.GroupId))
}

// OvertList 公开群列表
func (c *Group) OvertList(ctx *ichat.Context) error {

	params := &web.GroupOvertListRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)
	}

	list, err := c.service.Dao().SearchOvertList(ctx.Context, params.Name, params.Page, 21)
	if err != nil {
		return ctx.BusinessError("查询异常！")
	}

	if len(list) == 0 {
		return ctx.Success(entity.H{
			"items": make([]interface{}, 0),
			"next":  false,
		})
	}

	ids := make([]int, 0)
	for _, val := range list {
		ids = append(ids, val.Id)
	}

	count, err := c.groupMemberService.Dao().CountGroupMemberNum(ids)
	if err != nil {
		return ctx.BusinessError("查询异常！")
	}

	countMap := make(map[int]int)
	for _, member := range count {
		countMap[member.GroupId] = member.Count
	}

	checks, err := c.groupMemberService.Dao().CheckUserGroup(ids, ctx.UserId())
	if err != nil {
		return ctx.BusinessError("查询异常！")
	}

	items := make([]*entity.H, 0)
	for i, value := range list {
		if i >= 20 {
			break
		}

		item := &entity.H{
			"id":         value.Id,
			"type":       value.Type,
			"name":       value.Name,
			"avatar":     value.Avatar,
			"profile":    value.Profile,
			"count":      countMap[value.Id],
			"max_num":    value.MaxNum,
			"is_member":  sliceutil.InInt(value.Id, checks),
			"created_at": timeutil.FormatDatetime(value.CreatedAt),
		}

		items = append(items, item)
	}

	return ctx.Success(entity.H{
		"items": items,
		"next":  len(list) > 20,
	})
}

// Handover 群主交接
func (c *Group) Handover(ctx *ichat.Context) error {

	params := &web.GroupHandoverRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)
	}

	uid := ctx.UserId()
	if !c.groupMemberService.Dao().IsMaster(params.GroupId, uid) {
		return ctx.BusinessError("暂无权限！")
	}

	if uid == params.UserId {
		return ctx.BusinessError("暂无权限！")
	}

	err := c.groupMemberService.Handover(params.GroupId, uid, params.UserId)
	if err != nil {
		return ctx.BusinessError("转让群主失败！")
	}

	return ctx.Success(entity.H{})
}

// AssignAdmin 分配管理员
func (c *Group) AssignAdmin(ctx *ichat.Context) error {

	params := &web.GroupAssignAdminRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)
	}

	uid := ctx.UserId()
	if !c.groupMemberService.Dao().IsMaster(params.GroupId, uid) {
		return ctx.BusinessError("暂无权限！")
	}

	leader := 0
	if params.Mode == 1 {
		leader = 1
	}

	err := c.groupMemberService.UpdateLeaderStatus(params.GroupId, params.UserId, leader)
	if err != nil {
		logger.Error("[Group AssignAdmin] 设置管理员信息失败 err :", err.Error())
		return ctx.BusinessError("设置管理员信息失败！")
	}

	return ctx.Success(entity.H{})
}

// NoSpeak 禁止发言
func (c *Group) NoSpeak(ctx *ichat.Context) error {

	params := &web.GroupNoSpeakRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)
	}

	uid := ctx.UserId()
	if !c.groupMemberService.Dao().IsLeader(params.GroupId, uid) {
		return ctx.BusinessError("暂无权限！")
	}

	status := 1
	if params.Mode == 2 {
		status = 0
	}

	err := c.groupMemberService.UpdateMuteStatus(params.GroupId, params.UserId, status)
	if err != nil {
		return ctx.BusinessError("设置群成员禁言状态失败！")
	}

	return ctx.Success(entity.H{})
}

//全员禁言
func (c *Group) AllNoSpeak(ctx *ichat.Context) error {
	params := &web.GroupAllNoSpeakRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)
	}
	uid := ctx.UserId()
	if !c.groupMemberService.Dao().IsLeader(params.GroupId, uid) {
		return ctx.BusinessError("暂无权限！")
	}

	err := c.groupMemberService.UpdateAllMuteStatus(params.GroupId, params.Mode)
	if err != nil {
		return ctx.BusinessError("设置群成员禁言状态失败！")
	}

	content := jsonutil.Encode(map[string]interface{}{
		"event": entity.EventTalkMuteGroup,
		"data": jsonutil.Encode(map[string]interface{}{
			"group_id": params.GroupId,
			"is_mute":  params.Mode,
		}),
	})

	// 创建一个Channel
	channel, err := c.mq.Channel()
	if err != nil {
		log.Println("Failed to open a channel:", err.Error())

	}
	defer channel.Close()

	// 声明exchange
	if err := channel.ExchangeDeclare(
		c.config.RabbitMQ.ExchangeName, //name
		"fanout",                       //exchangeType
		true,                           //durable
		false,                          //auto-deleted
		false,                          //internal
		false,                          //noWait
		nil,                            //arguments
	); err != nil {
		log.Println("Failed to declare a exchange:", err.Error())
	}

	c.messageService.SendAll(channel, content)

	return ctx.Success(entity.H{})
}

func (c *Group) Open(ctx *ichat.Context) error {
	params := &web.GroupOpenRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)
	}
	uid := ctx.UserId()
	if !c.groupMemberService.Dao().IsLeader(params.GroupId, uid) {
		return ctx.BusinessError("暂无权限！")
	}

	err := c.groupMemberService.UpdateIsOvertStatus(params.GroupId, params.Mode)
	if err != nil {
		return ctx.BusinessError("设置群公开失败！")
	}

	return ctx.Success(entity.H{})
}

//修改群头像
func (c *Group) Avatar(ctx *ichat.Context) error {
	params := &web.GroupAvatarRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)

	}
	uid := ctx.UserId()
	if !c.groupMemberService.Dao().IsLeader(params.GroupId, uid) {
		return ctx.BusinessError("暂无权限！")

	}
	err := c.service.Avatar(ctx.Context, &model.UpdateGroupOpts{
		GroupId: params.GroupId,
		Avatar:  params.Avatar,
	})
	if err != nil {
		logger.Error("[Group NoSpeak] 设置群头像失败 err :", err.Error())
		return ctx.BusinessError("操作失败！")

	}
	return ctx.Success(entity.H{})
}
