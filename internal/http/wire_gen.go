// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"context"
	"github.com/google/wire"
	"go-chat/config"
	"go-chat/internal/cache"
	"go-chat/internal/dao"
	"go-chat/internal/dao/note"
	"go-chat/internal/dao/organize"
	"go-chat/internal/http/internal/handler"
	"go-chat/internal/http/internal/handler/api/v1"
	"go-chat/internal/http/internal/handler/api/v1/article"
	"go-chat/internal/http/internal/handler/api/v1/contact"
	"go-chat/internal/http/internal/handler/api/v1/group"
	"go-chat/internal/http/internal/handler/api/v1/talk"
	"go-chat/internal/http/internal/router"
	"go-chat/internal/pkg/client"
	"go-chat/internal/provider"
	"go-chat/internal/service"
	note2 "go-chat/internal/service/note"
	organize2 "go-chat/internal/service/organize"
)

import (
	_ "github.com/urfave/cli/v2"
	_ "go-chat/internal/pkg/validation"
)

// Injectors from wire.go:

func Initialize(ctx context.Context, conf *config.Config) *AppProvider {
	redisClient := provider.NewRedisClient(ctx, conf)
	smsCodeCache := cache.NewSmsCodeCache(redisClient)
	smsService := service.NewSmsService(smsCodeCache)
	db := provider.NewMySQLClient(conf)
	baseDao := dao.NewBaseDao(db, redisClient)
	usersDao := dao.NewUserDao(baseDao)
	userService := service.NewUserService(usersDao)
	common := v1.NewCommonHandler(conf, smsService, userService)
	memberDao := dao.NewMemberDao(baseDao)
	memberService := service.NewMemberService(memberDao)
	session := cache.NewSession(redisClient)
	redisLock := cache.NewRedisLock(redisClient)
	baseService := service.NewBaseService(db, redisClient)
	unreadTalkCache := cache.NewUnreadTalkCache(redisClient)
	lastMessage := cache.NewLastMessage(redisClient)
	talkVote := cache.NewTalkVote(redisClient)
	talkRecordsVoteDao := dao.NewTalkRecordsVoteDao(baseDao, talkVote)
	relation := cache.NewRelation(redisClient)
	groupMemberDao := dao.NewGroupMemberDao(baseDao, relation)
	sidServer := cache.NewSid(redisClient)
	wsClientSession := cache.NewWsClientSession(redisClient, conf, sidServer)
	filesystem := provider.NewFilesystem(conf)
	splitUploadDao := dao.NewFileSplitUploadDao(baseDao)
	talkMessageService := service.NewTalkMessageService(baseService, conf, unreadTalkCache, lastMessage, talkRecordsVoteDao, groupMemberDao, sidServer, wsClientSession, filesystem, splitUploadDao)
	httpClient := provider.NewHttpClient()
	clientHttpClient := client.NewHttpClient(httpClient)
	ipAddressService := service.NewIpAddressService(baseService, conf, clientHttpClient)
	talkSessionDao := dao.NewTalkSessionDao(baseDao)
	talkSessionService := service.NewTalkSessionService(baseService, talkSessionDao)
	articleClassDao := note.NewArticleClassDao(baseDao)
	articleClassService := note2.NewArticleClassService(baseService, articleClassDao)
	auth := v1.NewAuthHandler(conf, userService, memberService, smsService, session, redisLock, talkMessageService, ipAddressService, talkSessionService, articleClassService)
	organizeDao := organize.NewOrganizeDao(baseDao)
	organizeService := organize2.NewOrganizeService(baseService, organizeDao)
	user := v1.NewUserHandler(userService, smsService, organizeService)
	departmentDao := organize.NewDepartmentDao(baseDao)
	organizeDeptService := organize2.NewOrganizeDeptService(baseService, departmentDao)
	positionDao := organize.NewPositionDao(baseDao)
	positionService := organize2.NewPositionService(baseService, positionDao)
	v1Organize := v1.NewOrganizeHandler(organizeDeptService, organizeService, positionService)
	talkService := service.NewTalkService(baseService, groupMemberDao)
	talkMessageForwardService := service.NewTalkMessageForwardService(baseService)
	splitUploadService := service.NewSplitUploadService(baseService, splitUploadDao, conf, filesystem)
	contactDao := dao.NewContactDao(baseDao, relation)
	contactService := service.NewContactService(baseService, contactDao)
	groupMemberService := service.NewGroupMemberService(baseService, groupMemberDao)
	message := talk.NewTalkMessageHandler(talkMessageService, talkService, talkRecordsVoteDao, talkMessageForwardService, splitUploadService, contactService, groupMemberService, organizeService)
	groupDao := dao.NewGroupDao(baseDao)
	groupService := service.NewGroupService(baseService, groupDao, groupMemberDao, relation)
	authPermissionService := service.NewAuthPermissionService(contactDao, groupMemberDao, organizeDao)
	talkTalk := talk.NewTalkHandler(talkService, talkSessionService, redisLock, userService, wsClientSession, lastMessage, unreadTalkCache, contactService, groupService, authPermissionService)
	talkRecordsDao := dao.NewTalkRecordsDao(baseDao)
	talkRecordsService := service.NewTalkRecordsService(baseService, talkVote, talkRecordsVoteDao, groupMemberDao, talkRecordsDao)
	records := talk.NewTalkRecordsHandler(talkRecordsService, groupMemberService, filesystem, authPermissionService)
	emoticonDao := dao.NewEmoticonDao(baseDao)
	emoticonService := service.NewEmoticonService(baseService, emoticonDao, filesystem)
	emoticon := v1.NewEmoticonHandler(emoticonService, filesystem, redisLock)
	upload := v1.NewUploadHandler(conf, filesystem, splitUploadService)
	groupNoticeDao := dao.NewGroupNoticeDao(baseDao)
	groupNoticeService := service.NewGroupNoticeService(groupNoticeDao)
	groupGroup := group.NewGroupHandler(groupService, groupMemberService, talkSessionService, redisLock, contactService, userService, groupNoticeService, talkMessageService)
	notice := group.NewGroupNoticeHandler(groupNoticeService, groupMemberService)
	groupApplyDao := dao.NewGroupApply(baseDao)
	groupApplyService := service.NewGroupApplyService(baseService, groupApplyDao)
	apply := group.NewApplyHandler(groupApplyService, groupMemberService, groupService)
	contactContact := contact.NewContactHandler(contactService, wsClientSession, userService, talkSessionService, talkMessageService, organizeService)
	contactApplyService := service.NewContactsApplyService(baseService)
	contactApply := contact.NewContactsApplyHandler(contactApplyService, userService, talkMessageService, contactService)
	articleService := note2.NewArticleService(baseService)
	articleAnnexDao := note.NewArticleAnnexDao(baseDao)
	articleAnnexService := note2.NewArticleAnnexService(baseService, articleAnnexDao, filesystem)
	articleArticle := article.NewArticleHandler(articleService, filesystem, articleAnnexService)
	annex := article.NewAnnexHandler(articleAnnexService, filesystem)
	class := article.NewClassHandler(articleClassService)
	articleTagService := note2.NewArticleTagService(baseService)
	tag := article.NewTagHandler(articleTagService)
	apiHandler := &handler.ApiHandler{
		Common:        common,
		Auth:          auth,
		User:          user,
		Organize:      v1Organize,
		TalkMessage:   message,
		Talk:          talkTalk,
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
	}
	adminHandler := &handler.AdminHandler{}
	openHandler := &handler.OpenHandler{}
	handlerHandler := &handler.Handler{
		Api:   apiHandler,
		Admin: adminHandler,
		Open:  openHandler,
	}
	engine := router.NewRouter(conf, handlerHandler, session)
	httpServer := provider.NewHttpServer(conf, engine)
	appProvider := &AppProvider{
		Config: conf,
		Server: httpServer,
	}
	return appProvider
}

// wire.go:

var providerSet = wire.NewSet(provider.NewMySQLClient, provider.NewRedisClient, provider.NewHttpClient, provider.NewHttpServer, provider.NewFilesystem, client.NewHttpClient, router.NewRouter, wire.Struct(new(handler.ApiHandler), "*"), wire.Struct(new(handler.AdminHandler), "*"), wire.Struct(new(handler.OpenHandler), "*"), wire.Struct(new(handler.Handler), "*"), wire.Struct(new(AppProvider), "*"))

var cacheProviderSet = wire.NewSet(cache.NewSession, cache.NewSid, cache.NewUnreadTalkCache, cache.NewRedisLock, cache.NewWsClientSession, cache.NewLastMessage, cache.NewTalkVote, cache.NewRoom, cache.NewRelation, cache.NewSmsCodeCache)

var daoProviderSet = wire.NewSet(dao.NewBaseDao, dao.NewContactDao, dao.NewGroupMemberDao, dao.NewUserDao, dao.NewMemberDao, dao.NewGroupDao, dao.NewGroupApply, dao.NewTalkRecordsDao, dao.NewGroupNoticeDao, dao.NewTalkSessionDao, dao.NewEmoticonDao, dao.NewTalkRecordsVoteDao, dao.NewFileSplitUploadDao, note.NewArticleClassDao, note.NewArticleAnnexDao, organize.NewDepartmentDao, organize.NewOrganizeDao, organize.NewPositionDao)

var serviceProviderSet = wire.NewSet(service.NewBaseService, service.NewUserService, service.NewSmsService, service.NewTalkService, service.NewTalkMessageService, service.NewClientService, service.NewGroupService, service.NewGroupMemberService, service.NewGroupNoticeService, service.NewGroupApplyService, service.NewTalkSessionService, service.NewTalkMessageForwardService, service.NewEmoticonService, service.NewTalkRecordsService, service.NewContactService, service.NewContactsApplyService, service.NewSplitUploadService, service.NewIpAddressService, service.NewAuthPermissionService, service.NewMemberService, note2.NewArticleService, note2.NewArticleTagService, note2.NewArticleClassService, note2.NewArticleAnnexService, organize2.NewOrganizeDeptService, organize2.NewOrganizeService, organize2.NewPositionService)
