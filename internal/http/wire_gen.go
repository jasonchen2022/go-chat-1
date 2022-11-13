// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"context"
	"github.com/google/wire"
	"go-chat/config"
	"go-chat/internal/http/internal/handler"
	"go-chat/internal/http/internal/handler/admin"
	v1_2 "go-chat/internal/http/internal/handler/admin/v1"
	"go-chat/internal/http/internal/handler/open"
	v1_3 "go-chat/internal/http/internal/handler/open/v1"
	"go-chat/internal/http/internal/handler/web"
	"go-chat/internal/http/internal/handler/web/v1"
	"go-chat/internal/http/internal/handler/web/v1/article"
	"go-chat/internal/http/internal/handler/web/v1/contact"
	"go-chat/internal/http/internal/handler/web/v1/group"
	"go-chat/internal/http/internal/handler/web/v1/site"
	"go-chat/internal/http/internal/handler/web/v1/talk"
	"go-chat/internal/http/internal/router"
	"go-chat/internal/provider"
	"go-chat/internal/repository/cache"
	"go-chat/internal/repository/dao"
	"go-chat/internal/repository/dao/note"
	"go-chat/internal/repository/dao/organize"
	"go-chat/internal/service"
	note2 "go-chat/internal/service/note"
	organize2 "go-chat/internal/service/organize"
	"go-chat/internal/service/push"
)

import (
	_ "github.com/urfave/cli/v2"
	_ "go-chat/internal/pkg/validation"
)

// Injectors from wire.go:

func Initialize(ctx context.Context, conf *config.Config) *AppProvider {
	db := provider.NewMySQLClient(conf)
	client := provider.NewRedisClient(ctx, conf)
	connection := provider.NewRabbitMQClient(ctx, conf)
	baseService := service.NewBaseService(db, client, connection, conf)
	smsCodeCache := cache.NewSmsCodeCache(client)
	smsService := service.NewSmsService(baseService, smsCodeCache)
	baseDao := dao.NewBaseDao(db, client)
	usersDao := dao.NewUserDao(baseDao)
	userService := service.NewUserService(usersDao)
	common := v1.NewCommon(conf, smsService, userService)
	memberDao := dao.NewMemberDao(baseDao)
	memberService := service.NewMemberService(memberDao)
	contactRemark := cache.NewContactRemark(client)
	relation := cache.NewRelation(client)
	contactDao := dao.NewContactDao(baseDao, contactRemark, relation)
	contactService := service.NewContactService(baseService, contactDao)
	sessionStorage := cache.NewSessionStorage(client)
	redisLock := cache.NewRedisLock(client)
	messageStorage := cache.NewMessageStorage(client)
	unreadStorage := cache.NewUnreadStorage(client)
	talkVote := cache.NewTalkVote(client)
	talkRecordsVoteDao := dao.NewTalkRecordsVoteDao(baseDao, talkVote)
	groupMemberDao := dao.NewGroupMemberDao(baseDao, relation)
	sidServer := cache.NewSid(client)
	wsClientSession := cache.NewWsClientSession(client, conf, sidServer)
	filesystem := provider.NewFilesystem(conf)
	splitUploadDao := dao.NewFileSplitUploadDao(baseDao)
	sensitiveMatchService := service.NewSensitiveMatchService(db, client)
	talkMessageService := service.NewTalkMessageService(baseService, conf, unreadStorage, messageStorage, talkRecordsVoteDao, groupMemberDao, sidServer, wsClientSession, filesystem, splitUploadDao, sensitiveMatchService, contactDao)
	httpClient := provider.NewHttpClient()
	requestClient := provider.NewRequestClient(httpClient)
	ipAddressService := service.NewIpAddressService(baseService, conf, requestClient)
	talkSessionDao := dao.NewTalkSessionDao(baseDao)
	talkSessionService := service.NewTalkSessionService(baseService, talkSessionDao, contactDao)
	articleClassDao := note.NewArticleClassDao(baseDao)
	articleClassService := note2.NewArticleClassService(baseService, articleClassDao)
	robotDao := dao.NewRobotDao(baseDao)
	auth := v1.NewAuth(conf, userService, memberService, contactService, smsService, sessionStorage, redisLock, messageStorage, unreadStorage, talkMessageService, ipAddressService, talkSessionService, articleClassService, robotDao)
	organizeDao := organize.NewOrganizeDao(baseDao)
	organizeService := organize2.NewOrganizeService(baseService, organizeDao)
	user := v1.NewUser(userService, smsService, organizeService, talkMessageService)
	departmentDao := organize.NewDepartmentDao(baseDao)
	deptService := organize2.NewOrganizeDeptService(baseService, departmentDao)
	positionDao := organize.NewPositionDao(baseDao)
	positionService := organize2.NewPositionService(baseService, positionDao)
	v1Organize := v1.NewOrganize(deptService, organizeService, positionService)
	talkService := service.NewTalkService(baseService, groupMemberDao)
	groupDao := dao.NewGroupDao(baseDao)
	groupService := service.NewGroupService(baseService, groupDao, groupMemberDao, relation, talkMessageService)
	authPermissionService := service.NewAuthPermissionService(contactDao, groupMemberDao, organizeDao)
	talkTalk := talk.NewTalk(talkService, talkSessionService, redisLock, userService, wsClientSession, messageStorage, contactService, unreadStorage, contactRemark, groupService, authPermissionService)
	talkMessageForwardService := service.NewTalkMessageForwardService(baseService, talkMessageService)
	splitUploadService := service.NewSplitUploadService(baseService, splitUploadDao, conf, filesystem)
	groupMemberService := service.NewGroupMemberService(baseService, groupMemberDao)
	message := talk.NewMessage(talkMessageService, talkService, talkRecordsVoteDao, talkMessageForwardService, splitUploadService, contactService, groupMemberService, organizeService)
	talkRecordsDao := dao.NewTalkRecordsDao(baseDao)
	talkRecordsService := service.NewTalkRecordsService(baseService, talkVote, talkRecordsVoteDao, groupMemberDao, talkRecordsDao, sensitiveMatchService, contactService)
	records := talk.NewRecords(talkRecordsService, groupMemberService, filesystem, authPermissionService)
	emoticonDao := dao.NewEmoticonDao(baseDao)
	emoticonService := service.NewEmoticonService(baseService, emoticonDao, filesystem)
	emoticon := v1.NewEmoticon(filesystem, emoticonService, redisLock)
	upload := v1.NewUpload(conf, filesystem, splitUploadService)
	groupNoticeDao := dao.NewGroupNoticeDao(baseDao)
	groupNoticeService := service.NewGroupNoticeService(groupNoticeDao)
	jpushService := push.NewJpushService(conf, client)
	groupGroup := group.NewGroup(groupService, groupMemberService, talkSessionService, redisLock, contactService, userService, groupNoticeService, talkMessageService, memberService, jpushService, connection, conf, wsClientSession, sidServer)
	notice := group.NewNotice(groupNoticeService, groupMemberService)
	groupApplyDao := dao.NewGroupApply(baseDao)
	groupApplyService := service.NewGroupApplyService(baseService, groupApplyDao)
	apply := group.NewApply(groupApplyService, groupMemberService, groupService)
	contactContact := contact.NewContact(contactService, wsClientSession, userService, talkSessionService, talkMessageService, organizeService)
	contactApplyService := service.NewContactsApplyService(baseService, talkMessageService)
	contactApply := contact.NewApply(contactApplyService, userService, talkMessageService, contactService)
	articleService := note2.NewArticleService(baseService)
	articleAnnexDao := note.NewArticleAnnexDao(baseDao)
	articleAnnexService := note2.NewArticleAnnexService(baseService, articleAnnexDao, filesystem)
	articleArticle := article.NewArticle(articleService, filesystem, articleAnnexService)
	annex := article.NewAnnex(articleAnnexService, filesystem)
	class := article.NewClass(articleClassService)
	articleTagService := note2.NewArticleTagService(baseService)
	tag := article.NewTag(articleTagService)
	navigationDao := dao.NewNavigationDao(baseDao)
	navigationService := service.NewNavigationService(navigationDao)
	navigation := site.NewNavigation(navigationService)
	webV1 := &web.V1{
		Common:        common,
		Auth:          auth,
		User:          user,
		Organize:      v1Organize,
		Talk:          talkTalk,
		TalkMessage:   message,
		TalkRecords:   records,
		Emoticon:      emoticon,
		Upload:        upload,
		Group:         groupGroup,
		GroupNotice:   notice,
		GroupApply:    apply,
		Contact:       contactContact,
		ContactsApply: contactApply,
		Article:       articleArticle,
		ArticleAnnex:  annex,
		ArticleClass:  class,
		ArticleTag:    tag,
		Navigation:    navigation,
	}
	webHandler := &web.Handler{
		V1: webV1,
	}
	index := v1_2.NewIndex()
	v1Auth := v1_2.NewAuth()
	adminV1 := &admin.V1{
		Index: index,
		Auth:  v1Auth,
	}
	v2 := &admin.V2{}
	adminHandler := &admin.Handler{
		V1: adminV1,
		V2: v2,
	}
	v1Index := v1_3.NewIndex()
	openV1 := &open.V1{
		Index: v1Index,
	}
	openHandler := &open.Handler{
		V1: openV1,
	}
	handlerHandler := &handler.Handler{
		Api:   webHandler,
		Admin: adminHandler,
		Open:  openHandler,
	}
	engine := router.NewRouter(conf, handlerHandler, sessionStorage)
	httpServer := provider.NewHttpServer(conf, engine)
	appProvider := &AppProvider{
		Config: conf,
		Server: httpServer,
	}
	return appProvider
}

// wire.go:

var providerSet = wire.NewSet(provider.NewMySQLClient, provider.NewRedisClient, provider.NewRabbitMQClient, provider.NewHttpClient, provider.NewEmailClient, provider.NewHttpServer, provider.NewFilesystem, provider.NewRequestClient, router.NewRouter, wire.Struct(new(web.Handler), "*"), wire.Struct(new(admin.Handler), "*"), wire.Struct(new(open.Handler), "*"), wire.Struct(new(handler.Handler), "*"), wire.Struct(new(AppProvider), "*"))

var cacheProviderSet = wire.NewSet(cache.NewSessionStorage, cache.NewSid, cache.NewUnreadStorage, cache.NewRedisLock, cache.NewWsClientSession, cache.NewMessageStorage, cache.NewTalkVote, cache.NewRoomStorage, cache.NewRelation, cache.NewSmsCodeCache, cache.NewContactRemark)

var daoProviderSet = wire.NewSet(dao.NewBaseDao, dao.NewContactDao, dao.NewGroupMemberDao, dao.NewUserDao, dao.NewMemberDao, dao.NewGroupDao, dao.NewGroupApply, dao.NewTalkRecordsDao, dao.NewGroupNoticeDao, dao.NewTalkSessionDao, dao.NewEmoticonDao, dao.NewTalkRecordsVoteDao, dao.NewFileSplitUploadDao, note.NewArticleClassDao, note.NewArticleAnnexDao, organize.NewDepartmentDao, organize.NewOrganizeDao, organize.NewPositionDao, dao.NewRobotDao, dao.NewNavigationDao)

var serviceProviderSet = wire.NewSet(service.NewBaseService, service.NewUserService, service.NewSmsService, service.NewTalkService, service.NewTalkMessageService, service.NewClientService, service.NewGroupService, service.NewGroupMemberService, service.NewGroupNoticeService, service.NewGroupApplyService, service.NewTalkSessionService, service.NewTalkMessageForwardService, service.NewEmoticonService, service.NewTalkRecordsService, service.NewContactService, service.NewSensitiveMatchService, service.NewContactsApplyService, service.NewSplitUploadService, service.NewIpAddressService, service.NewAuthPermissionService, service.NewMemberService, note2.NewArticleService, note2.NewArticleTagService, note2.NewArticleClassService, note2.NewArticleAnnexService, organize2.NewOrganizeDeptService, organize2.NewOrganizeService, organize2.NewPositionService, service.NewTemplateService, service.NewNavigationService, push.NewJpushService)
