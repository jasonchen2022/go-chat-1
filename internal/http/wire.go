//go:build wireinject
// +build wireinject

package main

import (
	"context"

	"go-chat/config"
	"go-chat/internal/http/internal/handler/admin"
	"go-chat/internal/http/internal/handler/open"
	"go-chat/internal/http/internal/handler/web"
	"go-chat/internal/provider"
	"go-chat/internal/repository/cache"
	"go-chat/internal/repository/dao"
	note3 "go-chat/internal/repository/dao/note"
	organize3 "go-chat/internal/repository/dao/organize"
	"go-chat/internal/service/note"
	"go-chat/internal/service/organize"

	"github.com/google/wire"
	"go-chat/internal/http/internal/handler"
	"go-chat/internal/http/internal/router"
	"go-chat/internal/service"
	"go-chat/internal/service/push"
)

var providerSet = wire.NewSet(
	// 基础服务
	provider.NewMySQLClient,
	provider.NewRedisClient,
	provider.NewRabbitMQClient,
	provider.NewHttpClient,
	provider.NewEmailClient,
	provider.NewHttpServer,
	provider.NewFilesystem,
	provider.NewRequestClient,

	// 注册路由
	router.NewRouter,
	wire.Struct(new(web.Handler), "*"),
	wire.Struct(new(admin.Handler), "*"),
	wire.Struct(new(open.Handler), "*"),
	wire.Struct(new(handler.Handler), "*"),

	// AppProvider
	wire.Struct(new(AppProvider), "*"),
)

var cacheProviderSet = wire.NewSet(
	cache.NewSessionStorage,
	cache.NewSid,
	cache.NewUnreadStorage,
	cache.NewRedisLock,
	cache.NewWsClientSession,
	cache.NewMessageStorage,
	cache.NewTalkVote,
	cache.NewRoomStorage,
	cache.NewRelation,
	cache.NewSmsCodeCache,
	cache.NewContactRemark,
)

var daoProviderSet = wire.NewSet(
	dao.NewBaseDao,
	dao.NewContactDao,
	dao.NewGroupMemberDao,
	dao.NewUserDao,
	dao.NewMemberDao,
	dao.NewGroupDao,
	dao.NewGroupApply,
	dao.NewTalkRecordsDao,
	dao.NewGroupNoticeDao,
	dao.NewTalkSessionDao,
	dao.NewEmoticonDao,
	dao.NewTalkRecordsVoteDao,
	dao.NewFileSplitUploadDao,
	note3.NewArticleClassDao,
	note3.NewArticleAnnexDao,
	organize3.NewDepartmentDao,
	organize3.NewOrganizeDao,
	organize3.NewPositionDao,
	dao.NewRobotDao,
	dao.NewNavigationDao,
)

var serviceProviderSet = wire.NewSet(
	service.NewBaseService,
	service.NewUserService,
	service.NewSmsService,
	service.NewTalkService,
	service.NewTalkMessageService,
	service.NewClientService,
	service.NewGroupService,
	service.NewGroupMemberService,
	service.NewGroupNoticeService,
	service.NewGroupApplyService,
	service.NewTalkSessionService,
	service.NewTalkMessageForwardService,
	service.NewEmoticonService,
	service.NewTalkRecordsService,
	service.NewContactService,
	service.NewSensitiveMatchService,
	service.NewContactsApplyService,
	service.NewSplitUploadService,
	service.NewIpAddressService,
	service.NewAuthPermissionService,
	service.NewMemberService,
	note.NewArticleService,
	note.NewArticleTagService,
	note.NewArticleClassService,
	note.NewArticleAnnexService,
	organize.NewOrganizeDeptService,
	organize.NewOrganizeService,
	organize.NewPositionService,
	service.NewTemplateService,
	service.NewNavigationService,
	push.NewJpushService,
)

func Initialize(ctx context.Context, conf *config.Config) *AppProvider {
	panic(
		wire.Build(
			providerSet,
			cacheProviderSet,   // 注入 Cache 依赖
			daoProviderSet,     // 注入 Dao 依赖
			serviceProviderSet, // 注入 Service 依赖
			web.ProviderSet,    // 注入 Web Handler 依赖
			admin.ProviderSet,  // 注入 Admin Handler 依赖
			open.ProviderSet,   // 注入 Open Handler 依赖
		),
	)
}
