package cron

import (
	"context"
	"fmt"

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
	for _, sid := range c.server.GetExpireServerAll(ctx) {

		iter := c.server.Redis().Scan(ctx, 0, fmt.Sprintf("ws:%s:*", sid), 100).Iterator()

		for iter.Next(ctx) {
			c.server.Redis().Del(ctx, iter.Val())
		}

		_ = c.server.DelExpireServer(ctx, sid)
	}

	logrus.Info("结束清除redis缓存")
	return nil
}
