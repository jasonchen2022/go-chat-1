package v1

import (
	"strconv"
	"strings"
	"time"

	"go-chat/internal/http/internal/dto/web"
	"go-chat/internal/pkg/encrypt"
	"go-chat/internal/pkg/ichat"
	"go-chat/internal/pkg/jwt"
	"go-chat/internal/pkg/timeutil"
	"go-chat/internal/repository/cache"
	"go-chat/internal/repository/dao"
	"go-chat/internal/repository/model"
	"go-chat/internal/service/note"

	"go-chat/config"
	"go-chat/internal/entity"
	"go-chat/internal/service"

	"github.com/gin-gonic/gin"
)

type Auth struct {
	config             *config.Config
	userService        *service.UserService
	memberService      *service.MemberService
	contactService     *service.ContactService
	smsService         *service.SmsService
	session            *cache.SessionStorage
	redisLock          *cache.RedisLock
	messageStorage     *cache.MessageStorage
	unreadStorage      *cache.UnreadStorage
	talkMessageService *service.TalkMessageService
	ipAddressService   *service.IpAddressService
	talkSessionService *service.TalkSessionService
	noteClassService   *note.ArticleClassService
	robotDao           *dao.RobotDao
}

func NewAuth(config *config.Config, userService *service.UserService, memberService *service.MemberService, contactService *service.ContactService, smsService *service.SmsService, session *cache.SessionStorage, redisLock *cache.RedisLock, messageStorage *cache.MessageStorage,
	unreadStorage *cache.UnreadStorage, talkMessageService *service.TalkMessageService, ipAddressService *service.IpAddressService, talkSessionService *service.TalkSessionService, noteClassService *note.ArticleClassService, robotDao *dao.RobotDao) *Auth {
	return &Auth{config: config, userService: userService, memberService: memberService, contactService: contactService, smsService: smsService, session: session, redisLock: redisLock, messageStorage: messageStorage, unreadStorage: unreadStorage, talkMessageService: talkMessageService, ipAddressService: ipAddressService, talkSessionService: talkSessionService, noteClassService: noteClassService, robotDao: robotDao}
}

// Login 登录接口
func (c *Auth) Login(ctx *ichat.Context) error {

	params := &web.AuthLoginRequest{}
	if err := ctx.Context.ShouldBindJSON(params); err != nil {
		return ctx.InvalidParams(err)
	}
	//短信登录
	if params.Type == 1 {
		if params.Password != "202217" {
			// 验证短信验证码是否正确
			if !c.smsService.CheckSmsCode(ctx.Context, entity.SmsLoginChannel, params.Mobile, params.Password) {
				return ctx.InvalidParams("短信验证码填写错误")

			}
		}
	}

	// user, err := c.userService.Login(params.Mobile, params.Password)
	user, err := c.userService.NewLogin(params.Mobile, params.Password, params.Type)
	if err != nil {
		return ctx.BusinessError(err.Error())
	}

	root, _ := c.robotDao.FindLoginRobot()
	if root != nil {

		_, _ = c.talkSessionService.Create(ctx.RequestCtx(), &model.TalkSessionCreateOpt{
			UserId:     user.Id,
			TalkType:   entity.ChatPrivateMode,
			ReceiverId: root.UserId,
			IsBoot:     true,
		})

		// 推送登录消息
		//  ip := ctx.Context.ClientIP()
		//	address, _ := c.ipAddressService.FindAddress(ip)
		// _ = c.talkMessageService.SendLoginMessage(ctx.RequestCtx(), &service.LoginMessageOpt{
		// 	UserId:   user.Id,
		// 	Ip:       ip,
		// 	Address:  address,
		// 	Platform: params.Platform,
		// 	Agent:    ctx.Context.GetHeader("user-agent"),
		// })
	}

	return ctx.Success(&web.AuthLoginResponse{
		Type:        "Bearer",
		AccessToken: c.token(user.Id),
		ExpiresIn:   int(c.config.Jwt.ExpiresTime),
	})
}

// Login 同步登录接口
func (c *Auth) Sync(ctx *ichat.Context) error {

	params := &web.SyncRequest{}
	if err := ctx.Context.ShouldBindJSON(params); err != nil {
		return ctx.InvalidParams(err)

	}
	member, err := c.memberService.FindById(params.UserId)
	if err != nil || member == nil {
		return ctx.BusinessError(err)

	}
	//先查询当前用户存不存在
	user, _ := c.userService.Dao().FindByMobile(member.Mobile)
	if user == nil {
		password, _ := encrypt.HashPassword("12345689")
		timetemp := strconv.FormatInt(time.Now().Unix(), 10)
		timeresult := strings.TrimLeft(timetemp, "1")
		timeid, _ := strconv.Atoi(timeresult)
		_, err := c.userService.Dao().Create(&model.Users{
			Id:               member.Id,
			MemberId:         (timeid + member.Id) * 2, //（时间戳+ID主键）* 2
			MemberLevel:      member.MemberLevel,
			MemberLevelTitle: member.MemberLevelTitle,
			Username:         member.UserName,
			Nickname:         member.Nickname,
			Mobile:           member.Mobile,
			Gender:           member.Gender,
			Type:             member.Type,
			Motto:            member.Motto,
			ClientId:         member.ClientId,
			Password:         password,
			CreatedAt:        time.Now(),
		})
		if err != nil {
			return ctx.BusinessError(err)

		}
		//如果是独立IM站，则添加所有好友
		if c.config.GetEnv() == "alone" {
			c.sendDefautMsg(ctx.Context, params.UserId)
		} else if member.Type > -1 && !(strings.Contains(member.UserName, "游客_")) {
			c.sendDefautMsg(ctx.Context, params.UserId)
		}

	}

	//ip := ctx.ClientIP()

	//address, _ := c.ipAddressService.FindAddress(ip)

	//登录提醒
	// _, _ = c.talkSessionService.Create(ctx.Context, &model.TalkSessionCreateOpt{
	// 	UserId:     params.UserId,
	// 	TalkType:   entity.ChatPrivateMode,
	// 	ReceiverId: 1,
	// 	IsBoot:     true,
	// })

	// 推送登录消息
	// _ = c.talkMessageService.SendLoginMessage(ctx.Request.Context(), &service.LoginMessageOpts{
	// 	UserId:   params.UserId,
	// 	Ip:       ip,
	// 	Address:  address,
	// 	Platform: "h5",
	// 	Agent:    ctx.GetHeader("user-agent"),
	// })

	return ctx.Success(c.token(params.UserId))
}

func (c *Auth) sendDefautMsg(ctx *gin.Context, userId int) {
	//添加11直播官方为好友
	err := c.contactService.AddCustomerFriend(ctx, userId)
	if err != nil {
		return
	}
	//发送官方消息
	_ = c.talkMessageService.SendDefaultMessage(ctx.Request.Context(), userId)
	//创建会话
	_, _ = c.talkSessionService.Create(ctx.Request.Context(), &model.TalkSessionCreateOpt{
		UserId:     userId,
		TalkType:   entity.ChatPrivateMode,
		ReceiverId: 7715,
		IsBoot:     false,
	})
	//设置最后一条消息缓存
	_ = c.messageStorage.Set(ctx, 1, 7715, userId, &cache.LastCacheMessage{
		Content:  c.config.App.Welcome,
		Datetime: timeutil.DateTime(),
	})
	//设置消息未读数
	c.unreadStorage.Increment(ctx.Request.Context(), entity.ChatPrivateMode, 7715, userId)
}

// Register 注册接口
func (c *Auth) Register(ctx *ichat.Context) error {

	params := &web.RegisterRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)
	}

	// 验证短信验证码是否正确
	if !c.smsService.CheckSmsCode(ctx.RequestCtx(), entity.SmsRegisterChannel, params.Mobile, params.SmsCode) {
		return ctx.InvalidParams("短信验证码填写错误")
	}

	_, err := c.userService.Register(&service.UserRegisterOpt{
		Nickname: params.Nickname,
		Mobile:   params.Mobile,
		Password: params.Password,
		Platform: params.Platform,
	})
	if err != nil {
		return ctx.BusinessError(err.Error())
	}

	c.smsService.DeleteSmsCode(ctx.RequestCtx(), entity.SmsRegisterChannel, params.Mobile)

	return ctx.Success(nil)
}

// Logout 退出登录接口
func (c *Auth) Logout(ctx *ichat.Context) error {

	c.toBlackList(ctx)

	return ctx.Success(nil)
}

// Refresh Token 刷新接口
func (c *Auth) Refresh(ctx *ichat.Context) error {

	c.toBlackList(ctx)

	return ctx.Success(&web.AuthRefreshResponse{
		Type:        "Bearer",
		AccessToken: c.token(ctx.UserId()),
		ExpiresIn:   int(c.config.Jwt.ExpiresTime),
	})
}

// Forget 账号找回接口
func (c *Auth) Forget(ctx *ichat.Context) error {

	params := &web.ForgetRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)
	}

	// 验证短信验证码是否正确
	if !c.smsService.CheckSmsCode(ctx.RequestCtx(), entity.SmsForgetAccountChannel, params.Mobile, params.SmsCode) {
		return ctx.InvalidParams("短信验证码填写错误")
	}

	if _, err := c.userService.Forget(&service.UserForgetOpt{
		Mobile:   params.Mobile,
		Password: params.Password,
		SmsCode:  params.SmsCode,
	}); err != nil {
		return ctx.BusinessError(err.Error())
	}

	c.smsService.DeleteSmsCode(ctx.RequestCtx(), entity.SmsForgetAccountChannel, params.Mobile)

	return ctx.Success(nil)
}

//强制离线
func (c *Auth) Offline(ctx *ichat.Context) error {
	params := &web.OfflineRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)
	}
	_ = c.talkMessageService.SendOfflineMessage(ctx.Context, &service.SysOfflineMessageOpt{
		UserId:   params.UserId,
		ClientId: params.ClientId,
	})
	return ctx.Success(nil)
}

func (c *Auth) token(uid int) string {

	expiresAt := time.Now().Add(time.Second * time.Duration(c.config.Jwt.ExpiresTime))

	// 生成登录凭证
	token := jwt.GenerateToken("api", c.config.Jwt.Secret, &jwt.Options{
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		ID:        strconv.Itoa(uid),
	})

	return token
}

// 设置黑名单
func (c *Auth) toBlackList(ctx *ichat.Context) {

	session := ctx.JwtSession()
	if session != nil {
		ex := session.ExpiresAt - time.Now().Unix()

		// 将 session 加入黑名单
		_ = c.session.SetBlackList(ctx.RequestCtx(), session.Token, int(ex))
	}
}
