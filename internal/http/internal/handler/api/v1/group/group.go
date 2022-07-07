package group

import (
	"fmt"
	"time"

	"go-chat/internal/cache"
	"go-chat/internal/entity"
	"go-chat/internal/http/internal/request"
	"go-chat/internal/http/internal/response"
	"go-chat/internal/model"
	"go-chat/internal/pkg/encrypt"
	"go-chat/internal/pkg/jwtutil"
	"go-chat/internal/pkg/logger"
	"go-chat/internal/pkg/sliceutil"
	"go-chat/internal/pkg/timeutil"
	"go-chat/internal/service"

	"github.com/gin-gonic/gin"
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
}

func NewGroupHandler(
	service *service.GroupService,
	groupMemberService *service.GroupMemberService,
	talkSessionService *service.TalkSessionService,
	redisLock *cache.RedisLock,
	contactService *service.ContactService,
	userService *service.UserService,
	groupNoticeService *service.GroupNoticeService,
	messageService *service.TalkMessageService,
	memberService *service.MemberService,
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
	}
}

// Create 创建群聊分组
func (c *Group) Create(ctx *gin.Context) {
	params := &request.GroupCreateRequest{}
	if err := ctx.ShouldBind(params); err != nil {
		response.InvalidParams(ctx, err)
		return
	}

	gid, err := c.service.Create(ctx.Request.Context(), &service.CreateGroupOpts{
		UserId:    jwtutil.GetUid(ctx),
		Name:      params.Name,
		Avatar:    params.Avatar,
		Profile:   params.Profile,
		Type:      1, //暂时默认普通群
		MemberIds: sliceutil.ParseIds(params.MembersIds),
	})
	if err != nil {
		response.BusinessError(ctx, "创建群聊失败，请稍后再试！")
		return
	}

	response.Success(ctx, entity.H{
		"group_id": gid,
	})
}

// Create 创建聊天室分组
func (c *Group) CreateChat(ctx *gin.Context) {
	params := &request.GroupCreateChatRequest{}
	if err := ctx.ShouldBind(params); err != nil {
		response.InvalidParams(ctx, err)
		return
	}
	members, err := c.memberService.FindAdmin()
	if err != nil {
		response.BusinessError(ctx, "获取管理员账号列表失败")
		return
	}
	var userIds []int
	for _, member := range members {
		userIds = append(userIds, member.Id)
	}
	if len(userIds) > 0 {
		users, err := c.userService.Dao().FindByIds(userIds)
		if err != nil {
			response.BusinessError(ctx, "创建管理员账号失败")
			return
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
				c.userService.Dao().Create(&model.Users{
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
			}

		}
	}

	gid, err := c.service.Create(ctx.Request.Context(), &service.CreateGroupOpts{
		UserId:    params.AnchorId,
		Name:      params.Name,
		Profile:   params.Profile,
		Type:      -1, //默认聊天室
		MemberIds: userIds,
	})
	if err != nil {
		fmt.Printf("创建聊天室出错：%s", err.Error())
		response.BusinessError(ctx, "创建群聊失败，请稍后再试！")
		return
	}

	response.Success(ctx, entity.H{
		"group_id": gid,
	})
}

// Dismiss 解散群组
func (c *Group) Dismiss(ctx *gin.Context) {
	params := &request.GroupDismissRequest{}
	if err := ctx.ShouldBind(params); err != nil {
		response.InvalidParams(ctx, err)
		return
	}

	uid := jwtutil.GetUid(ctx)
	if !c.groupMemberService.Dao().IsMaster(params.GroupId, uid) {
		response.BusinessError(ctx, "暂无权限解散群组！")
		return
	}

	if err := c.service.Dismiss(ctx.Request.Context(), params.GroupId, jwtutil.GetUid(ctx)); err != nil {
		response.BusinessError(ctx, "群组解散失败！")
	} else {
		_ = c.messageService.SendSysMessage(ctx, &service.SysTextMessageOpts{
			UserId:     uid,
			TalkType:   entity.ChatGroupMode,
			ReceiverId: params.GroupId,
			Text:       "群组已被群主或管理员解散！",
		})

		response.Success(ctx, nil)
	}
}

// Invite 邀请好友加入群聊
func (c *Group) Invite(ctx *gin.Context) {
	params := &request.GroupInviteRequest{}
	if err := ctx.ShouldBind(params); err != nil {
		response.InvalidParams(ctx, err)
		return
	}

	key := fmt.Sprintf("group-join:%d", params.GroupId)
	if !c.redisLock.Lock(ctx, key, 20) {
		response.BusinessError(ctx, "网络异常，请稍后再试！")
		return
	}

	defer c.redisLock.UnLock(ctx, key)

	uid := jwtutil.GetUid(ctx)
	uids := sliceutil.UniqueInt(sliceutil.ParseIds(params.Ids))

	if len(uids) == 0 {
		response.BusinessError(ctx, "邀请好友列表不能为空！")
		return
	}

	if !c.groupMemberService.Dao().IsMember(params.GroupId, uid, true) {
		response.BusinessError(ctx, "非群组成员，无权邀请好友！")
		return
	}

	if err := c.service.InviteMembers(ctx, &service.InviteGroupMembersOpts{
		UserId:    uid,
		GroupId:   params.GroupId,
		MemberIds: uids,
	}); err != nil {
		response.BusinessError(ctx, "邀请好友加入群聊失败！")
	} else {
		response.Success(ctx, nil)
	}
}

// Join 加入群聊
func (c *Group) Join(ctx *gin.Context) {
	params := &request.GroupCommonRequest{}
	if err := ctx.ShouldBind(params); err != nil {
		response.InvalidParams(ctx, err)
		return
	}

	key := fmt.Sprintf("group-join:%d", params.GroupId)
	if !c.redisLock.Lock(ctx, key, 20) {
		response.BusinessError(ctx, "网络异常，请稍后再试！")
		return
	}

	defer c.redisLock.UnLock(ctx, key)

	uid := jwtutil.GetUid(ctx)
	ids := []int{uid}
	if err := c.service.InviteMembers(ctx, &service.InviteGroupMembersOpts{
		UserId:    1, //默认为管理员邀请
		GroupId:   params.GroupId,
		MemberIds: ids,
	}); err != nil {
		fmt.Printf("进群失败：%s", err.Error())
		response.BusinessError(ctx, "邀请好友加入群聊失败！")
	} else {
		response.Success(ctx, nil)
	}
}

// SignOut 退出群聊
func (c *Group) SignOut(ctx *gin.Context) {
	params := &request.GroupSecedeRequest{}
	if err := ctx.ShouldBind(params); err != nil {
		response.InvalidParams(ctx, err)
		return
	}

	uid := jwtutil.GetUid(ctx)
	if err := c.service.Secede(ctx.Request.Context(), params.GroupId, uid); err != nil {
		response.BusinessError(ctx, err.Error())
		return
	}

	// 删除聊天会话
	sid := c.talkSessionService.Dao().FindBySessionId(uid, params.GroupId, entity.ChatGroupMode)
	_ = c.talkSessionService.Delete(ctx, jwtutil.GetUid(ctx), sid)

	response.Success(ctx, nil)
}

// Setting 群设置接口（预留）
func (c *Group) Setting(ctx *gin.Context) {
	params := &request.GroupSettingRequest{}
	if err := ctx.ShouldBind(params); err != nil {
		response.InvalidParams(ctx, err)
		return
	}

	uid := jwtutil.GetUid(ctx)
	if !c.groupMemberService.Dao().IsLeader(params.GroupId, uid) {
		response.BusinessError(ctx, "无权限操作")
		return
	}

	if err := c.service.Update(ctx.Request.Context(), &service.UpdateGroupOpts{
		GroupId: params.GroupId,
		Name:    params.GroupName,
		Avatar:  params.Avatar,
		Profile: params.Profile,
	}); err != nil {
		response.BusinessError(ctx, err)
	} else {
		_ = c.messageService.SendSysMessage(ctx, &service.SysTextMessageOpts{
			UserId:     uid,
			TalkType:   entity.ChatGroupMode,
			ReceiverId: params.GroupId,
			Text:       "群主或管理员修改了群信息！",
		})

		response.Success(ctx, nil)
	}
}

// RemoveMembers 移除指定成员(群组&管理员权限)
func (c *Group) RemoveMembers(ctx *gin.Context) {
	params := &request.GroupRemoveMembersRequest{}
	if err := ctx.ShouldBind(params); err != nil {
		response.InvalidParams(ctx, err)
		return
	}

	uid := jwtutil.GetUid(ctx)

	if !c.groupMemberService.Dao().IsLeader(params.GroupId, uid) {
		response.BusinessError(ctx, "无权限操作")
		return
	}

	err := c.service.RemoveMembers(ctx.Request.Context(), &service.RemoveMembersOpts{
		UserId:    uid,
		GroupId:   params.GroupId,
		MemberIds: sliceutil.ParseIds(params.MembersIds),
	})

	if err != nil {
		response.BusinessError(ctx, err)
	} else {
		response.Success(ctx, nil)
	}
}

// Detail 获取群组信息
func (c *Group) Detail(ctx *gin.Context) {
	params := &request.GroupCommonRequest{}
	if err := ctx.ShouldBindQuery(params); err != nil {
		response.InvalidParams(ctx, err)
		return
	}

	uid := jwtutil.GetUid(ctx)

	groupInfo, err := c.service.Dao().FindById(params.GroupId)
	if err != nil {
		response.BusinessError(ctx, err)
		return
	}

	if groupInfo.Id == 0 {
		response.BusinessError(ctx, "数据不存在")
		return
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
	info["notice"] = entity.H{} //公告

	if notice, _ := c.groupNoticeService.Dao().GetLatestNotice(ctx, params.GroupId); err == nil {
		info["notice"] = notice
	}

	if nickname, err := c.userService.Dao().GetNickName(uid); err == nil {
		info["manager_nickname"] = nickname
	}
	//群成员
	members := c.groupMemberService.Dao().GetMembers(params.GroupId)
	info["members"] = members

	//会话
	talkSesstion, err := c.talkSessionService.FindTalkSession(ctx, params.GroupId, uid)
	if err == nil {
		info["is_top"] = talkSesstion.IsTop
		info["is_disturb"] = talkSesstion.IsDisturb
		info["is_show_nickname"] = talkSesstion.IsShowNickname
	}
	response.Success(ctx, info)
}

// EditRemark 修改群备注接口
func (c *Group) EditRemark(ctx *gin.Context) {
	params := &request.GroupEditRemarkRequest{}
	if err := ctx.ShouldBind(params); err != nil {
		response.InvalidParams(ctx, err)
		return
	}

	if err := c.groupMemberService.CardEdit(params.GroupId, jwtutil.GetUid(ctx), params.VisitCard); err != nil {
		response.BusinessError(ctx, "修改群备注失败！")
		return
	}

	response.Success(ctx, nil)
}

func (c *Group) GetInviteFriends(ctx *gin.Context) {
	params := &request.GetInviteFriendsRequest{}
	if err := ctx.ShouldBind(params); err != nil {
		response.InvalidParams(ctx, err)
		return
	}

	items, err := c.contactService.List(ctx, jwtutil.GetUid(ctx))
	if err != nil {
		response.BusinessError(ctx, err)
		return
	}

	if params.GroupId <= 0 {
		response.Success(ctx, items)
		return
	}

	mids := c.groupMemberService.Dao().GetMemberIds(params.GroupId)
	if len(mids) == 0 {
		response.Success(ctx, items)
		return
	}

	data := make([]*model.ContactListItem, 0)
	for i := 0; i < len(items); i++ {
		if !sliceutil.InInt(items[i].Id, mids) {
			data = append(data, items[i])
		}
	}

	response.Success(ctx, data)
}

func (c *Group) GetGroups(ctx *gin.Context) {
	items, err := c.service.List(jwtutil.GetUid(ctx))
	if err != nil {
		response.BusinessError(ctx, items)
		return
	}

	response.Success(ctx, entity.H{
		"rows": items,
	})
}

// GetMembers 获取群成员列表
func (c *Group) GetMembers(ctx *gin.Context) {
	params := &request.GroupCommonRequest{}
	if err := ctx.ShouldBind(params); err != nil {
		response.InvalidParams(ctx, err)
		return
	}

	if !c.groupMemberService.Dao().IsMember(params.GroupId, jwtutil.GetUid(ctx), false) {
		response.BusinessError(ctx, "非群成员无权查看成员列表！")
	} else {
		response.Success(ctx, c.groupMemberService.Dao().GetMembers(params.GroupId))
	}
}

// OvertList 公开群列表
func (c *Group) OvertList(ctx *gin.Context) {
	params := &request.GroupOvertListRequest{}
	if err := ctx.ShouldBind(params); err != nil {
		response.InvalidParams(ctx, err)
		return
	}

	list, err := c.service.Dao().SearchOvertList(ctx, params.Name, params.Page, 21)
	if err != nil {
		response.BusinessError(ctx, "查询异常！")
		return
	}

	if len(list) == 0 {
		response.Success(ctx, entity.H{
			"items": make([]interface{}, 0),
			"next":  false,
		})
		return
	}

	ids := make([]int, 0)
	for _, val := range list {
		ids = append(ids, val.Id)
	}

	count, err := c.groupMemberService.Dao().CountGroupMemberNum(ids)
	if err != nil {
		response.BusinessError(ctx, "查询异常！")
		return
	}

	countMap := make(map[int]int)
	for _, member := range count {
		countMap[member.GroupId] = member.Count
	}

	checks, err := c.groupMemberService.Dao().CheckUserGroup(ids, jwtutil.GetUid(ctx))
	if err != nil {
		response.BusinessError(ctx, "查询异常！")
		return
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

	response.Success(ctx, entity.H{
		"items": items,
		"next":  len(list) > 20,
	})
}

// Handover 群主交接
func (c *Group) Handover(ctx *gin.Context) {
	params := &request.GroupHandoverRequest{}
	if err := ctx.ShouldBind(params); err != nil {
		response.InvalidParams(ctx, err)
		return
	}

	uid := jwtutil.GetUid(ctx)
	if !c.groupMemberService.Dao().IsMaster(params.GroupId, uid) {
		response.BusinessError(ctx, "暂无权限！")
		return
	}

	if uid == params.UserId {
		response.BusinessError(ctx, "暂无权限！")
		return
	}

	err := c.groupMemberService.Handover(params.GroupId, uid, params.UserId)
	if err != nil {
		logger.Error("[Group Handover] 转让群主失败 err :", err.Error())
		response.BusinessError(ctx, "转让群主失败！")
		return
	}

	response.Success(ctx, entity.H{})
}

// AssignAdmin 分配管理员
func (c *Group) AssignAdmin(ctx *gin.Context) {
	params := &request.GroupAssignAdminRequest{}
	if err := ctx.ShouldBind(params); err != nil {
		response.InvalidParams(ctx, err)
		return
	}

	uid := jwtutil.GetUid(ctx)
	if !c.groupMemberService.Dao().IsMaster(params.GroupId, uid) {
		response.BusinessError(ctx, "暂无权限！")
		return
	}

	leader := 0
	if params.Mode == 1 {
		leader = 1
	}

	err := c.groupMemberService.UpdateLeaderStatus(params.GroupId, params.UserId, leader)
	if err != nil {
		logger.Error("[Group AssignAdmin] 设置管理员信息失败 err :", err.Error())
		response.BusinessError(ctx, "设置管理员信息失败！")
		return
	}

	response.Success(ctx, entity.H{})
}

// NoSpeak 禁止发言
func (c *Group) NoSpeak(ctx *gin.Context) {
	params := &request.GroupNoSpeakRequest{}
	if err := ctx.ShouldBind(params); err != nil {
		response.InvalidParams(ctx, err)
		return
	}

	uid := jwtutil.GetUid(ctx)
	if !c.groupMemberService.Dao().IsLeader(params.GroupId, uid) {
		response.BusinessError(ctx, "暂无权限！")
		return
	}

	status := 1
	if params.Mode == 2 {
		status = 0
	}

	err := c.groupMemberService.UpdateMuteStatus(params.GroupId, params.UserId, status)
	if err != nil {
		logger.Error("[Group NoSpeak] 设置群成员禁言状态失败 err :", err.Error())
		response.BusinessError(ctx, "操作失败！")
		return
	}

	response.Success(ctx, entity.H{})
}

// 修改群名
func (c *Group) Rename(ctx *gin.Context) {
	params := &request.GroupRenameRequest{}
	if err := ctx.ShouldBind(params); err != nil {
		response.InvalidParams(ctx, err)
		return
	}

	uid := jwtutil.GetUid(ctx)
	if !c.groupMemberService.Dao().IsLeader(params.GroupId, uid) {
		response.BusinessError(ctx, "暂无权限！")
		return
	}
	err := c.service.Rename(ctx.Request.Context(), &service.UpdateGroupOpts{
		GroupId: params.GroupId,
		Name:    params.GroupName,
	})
	if err != nil {
		logger.Error("[Group NoSpeak] 设置群名称失败 err :", err.Error())
		response.BusinessError(ctx, "操作失败！")
		return
	}

	response.Success(ctx, entity.H{})
}
