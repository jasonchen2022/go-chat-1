package talk

import (
	"errors"

	"go-chat/internal/http/internal/dto/web"
	"go-chat/internal/pkg/ichat"
	"go-chat/internal/repository/dao"
	"go-chat/internal/repository/model"
	"go-chat/internal/service/organize"

	"gorm.io/gorm"

	"go-chat/internal/entity"
	"go-chat/internal/pkg/sliceutil"
	"go-chat/internal/pkg/strutil"
	"go-chat/internal/service"
)

type Message struct {
	service            *service.TalkMessageService
	talkService        *service.TalkService
	talkRecordsVoteDao *dao.TalkRecordsVoteDao
	forwardService     *service.TalkMessageForwardService
	splitUploadService *service.SplitUploadService
	contactService     *service.ContactService
	groupMemberService *service.GroupMemberService
	organizeService    *organize.OrganizeService
}

func NewMessage(service *service.TalkMessageService, talkService *service.TalkService, talkRecordsVoteDao *dao.TalkRecordsVoteDao, forwardService *service.TalkMessageForwardService, splitUploadService *service.SplitUploadService, contactService *service.ContactService, groupMemberService *service.GroupMemberService, organizeService *organize.OrganizeService) *Message {
	return &Message{service: service, talkService: talkService, talkRecordsVoteDao: talkRecordsVoteDao, forwardService: forwardService, splitUploadService: splitUploadService, contactService: contactService, groupMemberService: groupMemberService, organizeService: organizeService}
}

type AuthorityOpts struct {
	TalkType   int // 对话类型
	UserId     int // 发送者ID
	ReceiverId int // 接收者ID
}

//判断当前发送者是否管理员
func (dao *Message) IsLeader(userId int) bool {
	var member_type int
	dao.service.Db().Table("users").Where("id = ?", userId).Select([]string{"type"}).Limit(1).Scan(&member_type)
	return member_type > 0
}

// 权限验证
func (c *Message) authority(ctx *ichat.Context, opt *AuthorityOpts) error {

	if opt.TalkType == entity.ChatPrivateMode {
		// 这里需要判断双方是否都是企业成员，如果是则无需添加好友即可聊天
		// if isOk, err := c.organizeService.Dao().IsQiyeMember(opt.UserId, opt.ReceiverId); err != nil {
		// 	return errors.New("系统繁忙，请稍后再试")
		// } else if isOk {
		// 	return nil
		// }

		if c.IsLeader(opt.UserId) || c.IsLeader(opt.ReceiverId) {
			return nil
		}

		isOk := c.contactService.Dao().IsFriend(ctx.RequestCtx(), opt.UserId, opt.ReceiverId, false)
		if isOk {
			return nil
		}

		return errors.New("暂无权限发送消息")
	} else {
		groupMemberInfo := &model.GroupMember{}
		err := c.groupMemberService.Db().First(groupMemberInfo, "group_id = ? and user_id = ?", opt.ReceiverId, opt.UserId).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return errors.New("暂无权限发送消息")
			}

			return errors.New("系统繁忙，请稍后再试")
		}

		if groupMemberInfo.IsQuit == 1 {
			return errors.New("暂无权限发送消息")
		}

		if groupMemberInfo.IsMute == 1 {
			return errors.New("已被群主或管理员禁言")
		}
	}

	return nil
}

// Text 发送文本消息
func (c *Message) Text(ctx *ichat.Context) error {

	params := &web.TextMessageRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)
	}

	uid := ctx.UserId()
	if err := c.authority(ctx, &AuthorityOpts{
		TalkType:   params.TalkType,
		UserId:     uid,
		ReceiverId: params.ReceiverId,
	}); err != nil {
		return ctx.BusinessError(err.Error())
	}
	id, err := c.service.SendTextMessage(ctx.RequestCtx(), &service.TextMessageOpt{
		UserId:     uid,
		TalkType:   params.TalkType,
		ReceiverId: params.ReceiverId,
		Text:       params.Text,
	})
	if err != nil {
		return ctx.BusinessError(err.Error())
	}
	return ctx.Success(id)
}

// Code 发送代码块消息
func (c *Message) Code(ctx *ichat.Context) error {

	params := &web.CodeMessageRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)
	}

	uid := ctx.UserId()
	if err := c.authority(ctx, &AuthorityOpts{
		TalkType:   params.TalkType,
		UserId:     uid,
		ReceiverId: params.ReceiverId,
	}); err != nil {
		return ctx.BusinessError(err.Error())
	}

	if err := c.service.SendCodeMessage(ctx.RequestCtx(), &service.CodeMessageOpt{
		UserId:     uid,
		TalkType:   params.TalkType,
		ReceiverId: params.ReceiverId,
		Lang:       params.Lang,
		Code:       params.Code,
	}); err != nil {
		return ctx.BusinessError(err.Error())
	}

	return ctx.Success(nil)
}

// Image 发送图片消息
func (c *Message) Image(ctx *ichat.Context) error {

	params := &web.ImageMessageRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)
	}

	file, err := ctx.Context.FormFile("image")
	if err != nil {
		return ctx.InvalidParams("image 字段必传")
	}

	if !sliceutil.InStr(strutil.FileSuffix(file.Filename), []string{"png", "jpg", "jpeg", "gif"}) {
		return ctx.InvalidParams("上传文件格式不正确,仅支持 png、jpg、jpeg 和 gif")
	}

	// 判断上传文件大小（5M）
	if file.Size > 5<<20 {
		return ctx.InvalidParams("上传文件大小不能超过5M")
	}

	uid := ctx.UserId()
	if err := c.authority(ctx, &AuthorityOpts{
		TalkType:   params.TalkType,
		UserId:     uid,
		ReceiverId: params.ReceiverId,
	}); err != nil {
		return ctx.BusinessError(err.Error())
	}
	id, err := c.service.SendImageMessage(ctx.RequestCtx(), &service.ImageMessageOpt{
		UserId:     uid,
		TalkType:   params.TalkType,
		ReceiverId: params.ReceiverId,
		File:       file,
	})

	if err != nil {
		return ctx.BusinessError(err.Error())
	}

	return ctx.Success(id)
}

// Image 发送图片消息
func (c *Message) ImageByUrl(ctx *ichat.Context) error {
	params := &web.ImageMessageRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)
	}

	if params.ImageUrl == "" {
		return ctx.InvalidParams("image_url 字段必传")
	}

	if !sliceutil.InStr(strutil.FileSuffix(params.ImageUrl), []string{"png", "jpg", "jpeg", "gif"}) {
		return ctx.InvalidParams("上传文件格式不正确,仅支持 png、jpg、jpeg 和 gif")
	}

	uid := ctx.UserId()
	if uid == params.ReceiverId {
		return ctx.InvalidParams("不能给自己发送消息")
	}
	if err := c.authority(ctx, &AuthorityOpts{
		TalkType:   params.TalkType,
		UserId:     uid,
		ReceiverId: params.ReceiverId,
	}); err != nil {
		return ctx.BusinessError(err.Error())
	}
	id, err := c.service.SendImageMessage(ctx.RequestCtx(), &service.ImageMessageOpt{
		UserId:     uid,
		TalkType:   params.TalkType,
		ReceiverId: params.ReceiverId,
		ImageUrl:   params.ImageUrl,
		File:       nil,
	})
	if err != nil {
		return ctx.BusinessError(err.Error())
	}
	return ctx.Success(id)
}

// File 发送文件消息
func (c *Message) File(ctx *ichat.Context) error {

	params := &web.FileMessageRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)
	}

	uid := ctx.UserId()
	if err := c.authority(ctx, &AuthorityOpts{
		TalkType:   params.TalkType,
		UserId:     uid,
		ReceiverId: params.ReceiverId,
	}); err != nil {
		return ctx.BusinessError(err.Error())
	}
	id, err := c.service.SendFileMessage(ctx.RequestCtx(), &service.FileMessageOpt{
		UserId:     uid,
		TalkType:   params.TalkType,
		ReceiverId: params.ReceiverId,
		UploadId:   params.UploadId,
	})
	if err != nil {
		return ctx.BusinessError(err.Error())
	}

	return ctx.Success(id)
}

// Vote 发送投票消息
func (c *Message) Vote(ctx *ichat.Context) error {

	params := &web.VoteMessageRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)
	}

	if len(params.Options) <= 1 {
		return ctx.InvalidParams("options 选项必须大于1")
	}

	if len(params.Options) > 6 {
		return ctx.InvalidParams("options 选项不能超过6个")
	}

	uid := ctx.UserId()
	if err := c.authority(ctx, &AuthorityOpts{
		TalkType:   entity.ChatGroupMode,
		UserId:     uid,
		ReceiverId: params.ReceiverId,
	}); err != nil {
		return ctx.BusinessError(err.Error())
	}
	id, err := c.service.SendVoteMessage(ctx.RequestCtx(), &service.VoteMessageOpt{
		UserId:     uid,
		ReceiverId: params.ReceiverId,
		Mode:       params.Mode,
		Anonymous:  params.Anonymous,
		Title:      params.Title,
		Options:    params.Options,
	})
	if err != nil {
		return ctx.BusinessError(err.Error())
	}

	return ctx.Success(id)
}

// Emoticon 发送表情包消息
func (c *Message) Emoticon(ctx *ichat.Context) error {

	params := &web.EmoticonMessageRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)
	}

	uid := ctx.UserId()
	if err := c.authority(ctx, &AuthorityOpts{
		TalkType:   params.TalkType,
		UserId:     uid,
		ReceiverId: params.ReceiverId,
	}); err != nil {
		return ctx.BusinessError(err.Error())
	}
	id, err := c.service.SendEmoticonMessage(ctx.RequestCtx(), &service.EmoticonMessageOpt{
		UserId:     uid,
		TalkType:   params.TalkType,
		ReceiverId: params.ReceiverId,
		EmoticonId: params.EmoticonId,
	})
	if err != nil {
		return ctx.BusinessError(err.Error())
	}

	return ctx.Success(id)
}

// Forward 发送转发消息
func (c *Message) Forward(ctx *ichat.Context) error {

	params := &web.ForwardMessageRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)
	}

	if params.ReceiveGroupIds == "" && params.ReceiveUserIds == "" {
		return ctx.InvalidParams("receive_user_ids 和 receive_group_ids 不能都为空")
	}

	uid := ctx.UserId()
	if err := c.authority(ctx, &AuthorityOpts{
		TalkType:   params.TalkType,
		UserId:     uid,
		ReceiverId: params.ReceiverId,
	}); err != nil {
		return ctx.BusinessError(err.Error())
	}

	forward := &service.TalkForwardOpt{
		Mode:       params.ForwardMode,
		UserId:     uid,
		ReceiverId: params.ReceiverId,
		TalkType:   params.TalkType,
		RecordsIds: sliceutil.ParseIds(params.RecordsIds),
		UserIds:    sliceutil.ParseIds(params.ReceiveUserIds),
		GroupIds:   sliceutil.ParseIds(params.ReceiveGroupIds),
	}

	if err := c.forwardService.SendForwardMessage(ctx.RequestCtx(), forward); err != nil {
		return ctx.BusinessError(err.Error())
	}

	return ctx.Success(nil)
}

// Card 发送用户名片消息
func (c *Message) Card(ctx *ichat.Context) error {

	params := &web.CardMessageRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)
	}

	uid := ctx.UserId()
	if err := c.authority(ctx, &AuthorityOpts{
		TalkType:   params.TalkType,
		UserId:     uid,
		ReceiverId: params.ReceiverId,
	}); err != nil {
		return ctx.BusinessError(err.Error())
	}
	id, err := c.service.SendCardMessage(ctx.RequestCtx(), &service.CardMessageOpt{
		UserId:     uid,
		TalkType:   params.TalkType,
		ReceiverId: params.ReceiverId,
		ContactId:  0,
	})
	// todo SendCardMessage
	if err != nil {
		return ctx.BusinessError(err.Error())
	}

	return ctx.Success(id)
}

// Collect 收藏聊天图片
func (c *Message) Collect(ctx *ichat.Context) error {

	params := &web.CollectMessageRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)
	}

	if err := c.talkService.CollectRecord(ctx.RequestCtx(), ctx.UserId(), params.RecordId); err != nil {
		return ctx.BusinessError(err.Error())
	}

	return ctx.Success(nil)
}

// Revoke 撤销聊天记录
func (c *Message) Revoke(ctx *ichat.Context) error {

	params := &web.RevokeMessageRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)
	}

	if err := c.service.SendRevokeRecordMessage(ctx.RequestCtx(), ctx.UserId(), params.RecordId); err != nil {
		return ctx.BusinessError(err.Error())
	}

	return ctx.Success(nil)
}

// Delete 删除聊天记录
func (c *Message) Delete(ctx *ichat.Context) error {

	params := &web.DeleteMessageRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)
	}

	if err := c.talkService.RemoveRecords(ctx.RequestCtx(), &service.TalkMessageDeleteOpt{
		UserId:     ctx.UserId(),
		TalkType:   params.TalkType,
		ReceiverId: params.ReceiverId,
		RecordIds:  params.RecordIds,
	}); err != nil {
		return ctx.BusinessError(err.Error())
	}

	return ctx.Success(nil)
}

// HandleVote 投票处理
func (c *Message) HandleVote(ctx *ichat.Context) error {

	params := &web.VoteMessageHandleRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)
	}

	vid, err := c.service.VoteHandle(ctx.RequestCtx(), &service.VoteMessageHandleOpt{
		UserId:   ctx.UserId(),
		RecordId: params.RecordId,
		Options:  params.Options,
	})
	if err != nil {
		return ctx.BusinessError(err.Error())
	}

	res, _ := c.talkRecordsVoteDao.GetVoteStatistics(ctx.RequestCtx(), vid)

	return ctx.Success(res)
}

// Location 发送位置消息
func (c *Message) Location(ctx *ichat.Context) error {

	params := &web.LocationMessageRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)
	}

	uid := ctx.UserId()
	if err := c.authority(ctx, &AuthorityOpts{
		TalkType:   params.TalkType,
		UserId:     uid,
		ReceiverId: params.ReceiverId,
	}); err != nil {
		return ctx.BusinessError(err.Error())
	}
	id, err := c.service.SendLocationMessage(ctx.RequestCtx(), &service.LocationMessageOpt{
		UserId:     uid,
		TalkType:   params.TalkType,
		ReceiverId: params.ReceiverId,
		Longitude:  params.Longitude,
		Latitude:   params.Latitude,
	})
	if err != nil {
		return ctx.BusinessError(err.Error())
	}

	return ctx.Success(id)
}
