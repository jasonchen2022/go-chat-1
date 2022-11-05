// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"context"
	"github.com/google/wire"
	"go-chat/config"
	"go-chat/internal/provider"
	"go-chat/internal/repository/cache"
	"go-chat/internal/repository/dao"
	"go-chat/internal/service"
	"go-chat/internal/websocket/internal/handler"
	"go-chat/internal/websocket/internal/process"
	"go-chat/internal/websocket/internal/process/handle"
	"go-chat/internal/websocket/internal/process/server"
	"go-chat/internal/websocket/internal/router"
)

// Injectors from wire.go:

func Initialize(ctx context.Context, conf *config.Config) *AppProvider {
	client := provider.NewRedisClient(ctx, conf)
	connection := provider.NewRabbitMQClient(ctx, conf)
	sidServer := cache.NewSid(client)
	wsClientSession := cache.NewWsClientSession(client, conf, sidServer)
	clientService := service.NewClientService(wsClientSession)
	roomStorage := cache.NewRoomStorage(client)
	db := provider.NewMySQLClient(conf)
	baseService := service.NewBaseService(db, client, connection, conf)
	baseDao := dao.NewBaseDao(db, client)
	relation := cache.NewRelation(client)
	groupMemberDao := dao.NewGroupMemberDao(baseDao, relation)
	groupMemberService := service.NewGroupMemberService(baseService, groupMemberDao)
	unreadStorage := cache.NewUnreadStorage(client)
	messageStorage := cache.NewMessageStorage(client)
	talkVote := cache.NewTalkVote(client)
	talkRecordsVoteDao := dao.NewTalkRecordsVoteDao(baseDao, talkVote)
	filesystem := provider.NewFilesystem(conf)
	splitUploadDao := dao.NewFileSplitUploadDao(baseDao)
	sensitiveMatchService := service.NewSensitiveMatchService(db, client)
	contactRemark := cache.NewContactRemark(client)
	contactDao := dao.NewContactDao(baseDao, contactRemark, relation)
	talkMessageService := service.NewTalkMessageService(baseService, conf, unreadStorage, messageStorage, talkRecordsVoteDao, groupMemberDao, sidServer, wsClientSession, filesystem, splitUploadDao, sensitiveMatchService, contactDao)
	defaultWebSocket := handler.NewDefaultWebSocket(client, connection, conf, clientService, roomStorage, groupMemberService, talkMessageService)
	exampleWebsocket := handler.NewExampleWebsocket()
	handlerHandler := &handler.Handler{
		DefaultWebSocket: defaultWebSocket,
		ExampleWebsocket: exampleWebsocket,
	}
	sessionStorage := cache.NewSessionStorage(client)
	engine := router.NewRouter(conf, handlerHandler, sessionStorage)
	websocketServer := provider.NewWebsocketServer(conf, engine)
	health := server.NewHealth(conf, sidServer)
	talkRecordsDao := dao.NewTalkRecordsDao(baseDao)
	talkRecordsService := service.NewTalkRecordsService(baseService, talkVote, talkRecordsVoteDao, groupMemberDao, talkRecordsDao, sensitiveMatchService)
	contactService := service.NewContactService(baseService, contactDao)
	subscribeConsume := handle.NewSubscribeConsume(conf, wsClientSession, roomStorage, talkRecordsService, contactService)
	wsSubscribe := server.NewWsSubscribe(client, connection, conf, subscribeConsume)
	subServers := &process.SubServers{
		Health:    health,
		Subscribe: wsSubscribe,
	}
	processServer := process.NewServer(subServers)
	appProvider := &AppProvider{
		Config:    conf,
		Server:    websocketServer,
		Coroutine: processServer,
	}
	return appProvider
}

// wire.go:

var providerSet = wire.NewSet(provider.NewMySQLClient, provider.NewRedisClient, provider.NewRabbitMQClient, provider.NewWebsocketServer, provider.NewFilesystem, router.NewRouter, wire.Struct(new(process.SubServers), "*"), process.NewServer, server.NewHealth, server.NewWsSubscribe, handle.NewSubscribeConsume, cache.NewSessionStorage, cache.NewSid, cache.NewRedisLock, cache.NewWsClientSession, cache.NewRoomStorage, cache.NewTalkVote, cache.NewRelation, cache.NewContactRemark, cache.NewUnreadStorage, cache.NewMessageStorage, dao.NewBaseDao, dao.NewTalkRecordsDao, dao.NewTalkRecordsVoteDao, dao.NewGroupMemberDao, dao.NewContactDao, dao.NewFileSplitUploadDao, service.NewBaseService, service.NewTalkRecordsService, service.NewClientService, service.NewGroupMemberService, service.NewContactService, service.NewSensitiveMatchService, service.NewTalkMessageService, handler.NewDefaultWebSocket, handler.NewExampleWebsocket, wire.Struct(new(handler.Handler), "*"), wire.Struct(new(AppProvider), "*"))
