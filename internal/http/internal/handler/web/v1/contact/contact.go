package contact

import (
	"errors"
	"strconv"
	"strings"
	"unicode/utf8"

	"go-chat/internal/http/internal/dto/web"
	"go-chat/internal/pkg/ichat"
	"go-chat/internal/repository/cache"
	"go-chat/internal/service/organize"

	"github.com/mozillazg/go-pinyin"
	"gorm.io/gorm"

	"go-chat/internal/entity"
	"go-chat/internal/pkg/strutil"
	"go-chat/internal/service"
)

type Contact struct {
	service            *service.ContactService
	wsClient           *cache.WsClientSession
	userService        *service.UserService
	talkListService    *service.TalkSessionService
	talkMessageService *service.TalkMessageService
	organizeService    *organize.OrganizeService
}

func NewContact(service *service.ContactService, wsClient *cache.WsClientSession, userService *service.UserService, talkListService *service.TalkSessionService, talkMessageService *service.TalkMessageService, organizeService *organize.OrganizeService) *Contact {
	return &Contact{service: service, wsClient: wsClient, userService: userService, talkListService: talkListService, talkMessageService: talkMessageService, organizeService: organizeService}
}

// List 联系人列表
func (c *Contact) List(ctx *ichat.Context) error {

	items, err := c.service.List(ctx.Context, ctx.UserId())

	if err != nil {
		return ctx.BusinessError(err.Error())
	}

	for _, item := range items {
		item.IsOnline = strutil.BoolToInt(c.wsClient.IsOnline(ctx.Context, entity.ImChannelDefault, strconv.Itoa(item.Id)))
	}

	return ctx.Success(items)
}

func (c *Contact) ListByPage(ctx *ichat.Context) error {

	params := &web.ContactQueryRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)

	}
	items, err := c.service.ListByPage(ctx.Context, ctx.UserId(), params.PageIndex, params.Keyword)
	// fmt.Println("PageIndexPageIndex==", params.PageIndex)
	if err != nil {
		return ctx.BusinessError(err)

	}
	var a = pinyin.NewArgs()
	for _, item := range items {
		_, size := utf8.DecodeRuneInString(item.Nickname)
		word := item.Nickname[:size]

		if item.Remark != "" {
			_, size = utf8.DecodeRuneInString(item.Remark)
			word = item.Remark[:size]
		}
		result := pinyin.Pinyin(string(word), a)
		englishIndex := ""
		if len(result) > 0 {
			englishIndex = string(result[0][0][0])
		}
		if englishIndex == "" {
			englishIndex = "#"
		}
		item.EnglishIndex = strings.ToUpper(englishIndex)
	}
	return ctx.Success(items)
}

func (c *Contact) TotalPage(ctx *ichat.Context) error {
	items, err := c.service.TotalPage(ctx.Context, ctx.UserId())

	if err != nil {
		return ctx.BusinessError(err)
	}
	return ctx.Success(items)
}

// Delete 删除联系人
func (c *Contact) Delete(ctx *ichat.Context) error {

	params := &web.ContactDeleteRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)
	}

	uid := ctx.UserId()
	if err := c.service.Delete(ctx.Context, uid, params.FriendId); err != nil {
		return ctx.BusinessError(err.Error())
	}

	// 解除好友关系后需添加一条聊天记录
	// _ = c.talkMessageService.SendSysMessage(ctx.Context, &service.SysTextMessageOpt{
	// 	UserId:     uid,
	// 	TalkType:   entity.ChatPrivateMode,
	// 	ReceiverId: params.FriendId,
	// 	Text:       "你与对方已经解除了好友关系",
	// })

	// 删除聊天会话
	sid := c.talkListService.Dao().FindBySessionId(uid, params.FriendId, entity.ChatPrivateMode)
	if err := c.talkListService.Delete(ctx.Context, ctx.UserId(), sid); err != nil {
		return ctx.BusinessError(err.Error())
	}

	return ctx.Success(nil)
}

// Search 查找联系人
func (c *Contact) Search(ctx *ichat.Context) error {

	params := &web.ContactSearchRequest{}
	if err := ctx.Context.ShouldBindQuery(params); err != nil {
		return ctx.InvalidParams(err)
	}

	user, err := c.userService.Dao().FindByMobile(params.Mobile)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.BusinessError("用户不存在")
		}

		return ctx.BusinessError(err.Error())
	}

	return ctx.Success(entity.H{
		"id":       user.Id,
		"mobile":   user.Mobile,
		"nickname": user.Nickname,
		"avatar":   user.Avatar,
		"gender":   user.Gender,
		"motto":    user.Motto,
	})
}

// EditRemark 编辑联系人备注
func (c *Contact) EditRemark(ctx *ichat.Context) error {

	params := &web.ContactEditRemarkRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)
	}

	if err := c.service.EditRemark(ctx.Context, ctx.UserId(), params.FriendId, params.Remarks); err != nil {
		return ctx.BusinessError(err.Error())
	}

	return ctx.Success(nil)
}

// Detail 联系人详情信息
func (c *Contact) Detail(ctx *ichat.Context) error {

	params := &web.ContactDetailRequest{}
	if err := ctx.Context.ShouldBindQuery(params); err != nil {
		return ctx.InvalidParams(err)
	}

	uid := ctx.UserId()

	user, err := c.userService.Dao().FindById(params.UserId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ctx.BusinessError("用户不存在")
		}

		return ctx.BusinessError(err.Error())
	}

	resp := entity.H{
		"avatar":        user.Avatar,
		"friend_apply":  0,
		"friend_status": 1, // 朋友关系[0:本人;1:陌生人;2:朋友;]
		"gender":        user.Gender,
		"id":            user.Id,
		"mobile":        user.Mobile,
		"motto":         user.Motto,
		"nickname":      user.Nickname,
		"remark_name":   "",
		"member_type":   user.Type, //用户类型(0会员 1主播 2专家 3超管)
		"is_mute":       user.IsMute,
	}

	if uid != params.UserId {
		if c.service.Dao().IsFriend(ctx.RequestCtx(), uid, params.UserId, false) {
			resp["friend_status"] = 2
			resp["remark_name"] = c.service.Dao().GetFriendRemark(ctx.RequestCtx(), uid, params.UserId)
		} else {
			isOk, _ := c.organizeService.Dao().IsQiyeMember(uid, params.UserId)
			if isOk {
				resp["friend_status"] = 2
			}
		}
	} else {
		resp["friend_status"] = 0
	}

	return ctx.Success(&resp)
}
