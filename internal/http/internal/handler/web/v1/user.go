package v1

import (
	"go-chat/internal/entity"
	"go-chat/internal/http/internal/dto/web"
	"go-chat/internal/pkg/encrypt"
	"go-chat/internal/pkg/ichat"
	"go-chat/internal/repository/model"
	"go-chat/internal/service"
	"go-chat/internal/service/organize"

	"github.com/sirupsen/logrus"
)

type User struct {
	service            *service.UserService
	smsService         *service.SmsService
	organizeServ       *organize.OrganizeService
	talkMessageService *service.TalkMessageService
}

func NewUser(service *service.UserService, smsService *service.SmsService, organizeServ *organize.OrganizeService, talkMessageService *service.TalkMessageService) *User {
	return &User{service: service, smsService: smsService, organizeServ: organizeServ, talkMessageService: talkMessageService}
}

// Detail 个人用户信息
func (u *User) Detail(ctx *ichat.Context) error {

	user, err := u.service.Dao().FindById(ctx.UserId())
	if err != nil {
		return ctx.Error(err.Error())
	}

	return ctx.Success(user)
}

//设置app状态
func (u *User) AppStatus(ctx *ichat.Context) error {
	params := &web.AppStatusRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)
	}
	err := u.service.Dao().SetAppStatus(params.UserId, params.ClientId)
	if err != nil {
		// err := u.talkMessageService.SendOfflineMessage(ctx.Context, &service.SysOfflineMessageOpt{UserId: params.UserId, ClientId: params.ClientId})
		// if err != nil {
		// 	logrus.Error("单点登录错误：", err.Error())
		// }
		logrus.Error("绑定用户client_id错误：", err.Error())
	}
	return ctx.Success(nil)
}

// Setting 用户设置
func (u *User) Setting(ctx *ichat.Context) error {

	uid := ctx.UserId()

	user, _ := u.service.Dao().FindById(uid)

	isOk, _ := u.organizeServ.Dao().IsQiyeMember(uid)

	return ctx.Success(entity.H{
		"user_info": entity.H{
			"uid":      user.Id,
			"nickname": user.Nickname,
			"avatar":   user.Avatar,
			"motto":    user.Motto,
			"gender":   user.Gender,
			"is_qiye":  isOk,
			"mobile":   user.Mobile,
			"email":    user.Email,
		},
		"setting": entity.H{
			"theme_mode":            "",
			"theme_bag_img":         "",
			"theme_color":           "",
			"notify_cue_tone":       "",
			"keyboard_event_notify": "",
		},
	})
}

// ChangeDetail 修改个人用户信息
func (u *User) ChangeDetail(ctx *ichat.Context) error {

	params := &web.ChangeUserDetailRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)
	}

	_, _ = u.service.Dao().BaseUpdate(&model.Users{}, entity.MapStrAny{
		"id": ctx.UserId(),
	}, entity.MapStrAny{
		"nickname": params.Nickname,
		"avatar":   params.Avatar,
		"gender":   params.Gender,
		"motto":    params.Motto,
	})

	return ctx.Success(nil, "个人信息修改成功")
}

// ChangePassword 修改密码接口
func (u *User) ChangePassword(ctx *ichat.Context) error {

	params := &web.ChangeUserPasswordRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)
	}

	uid := ctx.UserId()

	if uid == 2054 || uid == 2055 {
		return ctx.BusinessError("预览账号不支持修改密码")
	}

	if err := u.service.UpdatePassword(ctx.UserId(), params.OldPassword, params.NewPassword); err != nil {
		return ctx.BusinessError("密码修改失败")
	}

	return ctx.Success(nil, "密码修改成功")
}

// ChangeMobile 修改手机号接口
func (u *User) ChangeMobile(ctx *ichat.Context) error {

	params := &web.ChangeUserMobileRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)
	}

	uid := ctx.UserId()

	if uid == 2054 || uid == 2055 {
		return ctx.BusinessError("预览账号不支持修改手机号")
	}

	if !u.smsService.CheckSmsCode(ctx.RequestCtx(), entity.SmsChangeAccountChannel, params.Mobile, params.SmsCode) {
		return ctx.BusinessError("短信验证码填写错误")
	}

	user, _ := u.service.Dao().FindById(uid)

	if user.Mobile != params.Mobile {
		return ctx.BusinessError("手机号与原手机号一致无需修改")
	}

	if !encrypt.VerifyPassword(user.Password, params.Password) {
		return ctx.BusinessError("账号密码填写错误")
	}

	_, err := u.service.Dao().BaseUpdate(&model.Users{}, entity.MapStrAny{"id": user.Id}, entity.MapStrAny{"mobile": params.Mobile})
	if err != nil {
		return ctx.BusinessError("手机号修改失败")
	}

	return ctx.Success(nil, "手机号修改成功")
}

// ChangeEmail 修改邮箱接口
func (u *User) ChangeEmail(ctx *ichat.Context) error {

	params := &web.ChangeUserEmailRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)
	}

	// todo 1.验证邮件激活码是否正确

	return nil
}

/*
*发现好友  （除登录用户外）
*userId:登录用户id
*index:查询用户数
 */
func (u *User) RandomUser(ctx *ichat.Context) error {
	params := &web.RandUserRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)
	}

	users, err := u.service.RandomUser(ctx.UserId(), params.Index, params.UserName)
	if err != nil {
		return ctx.BusinessError("取发现用户列表出错")
	}
	return ctx.Success(users, "成功")
}

// Mute 禁言
func (u *User) Mute(ctx *ichat.Context) error {
	params := &web.ChangeMuteRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)

	}
	user, err := u.service.Dao().FindById(params.UserId)
	if err == nil {
		if user.IsMute == 1 {
			_, _ = u.service.Dao().BaseUpdate(&model.Users{}, entity.MapStrAny{
				"id": params.UserId,
			}, entity.MapStrAny{
				"is_mute": 0,
			})
			return ctx.Success(nil, "解禁成功")
		} else {
			_, _ = u.service.Dao().BaseUpdate(&model.Users{}, entity.MapStrAny{
				"id": params.UserId,
			}, entity.MapStrAny{
				"is_mute": 1,
			})
			return ctx.Success(nil, "禁言成功")
		}
	} else {
		return ctx.BusinessError(err.Error())
	}

}
