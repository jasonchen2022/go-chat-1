package cron

import (
	"context"

	"go-chat/internal/repository/cache"

	"github.com/sirupsen/logrus"
)

type ClearWsCache struct {
	server *cache.SidServer
}

func NewClearWsCache(server *cache.SidServer) *ClearWsCache {
	return &ClearWsCache{server: server}
}

// Spec 配置定时任务规则
// 每隔5分钟处理 websocket 缓存
func (c *ClearWsCache) Spec() string {
	return "*/5 * * * *"
}

func (c *ClearWsCache) Handle(ctx context.Context) error {

	logrus.Info("开始清除redis缓存")
	iter := c.server.Redis().Scan(ctx, 0, "ws:*", 5000).Iterator()

	for iter.Next(ctx) {
		c.server.Redis().Del(ctx, iter.Val())
		logrus.Info("删除Redis：", iter.Val())
	}
	logrus.Info("结束清除redis缓存")
	return nil
}
