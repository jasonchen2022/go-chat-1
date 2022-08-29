package cache

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-redis/redis/v8"
)

type UnreadTalkCache struct {
	rds *redis.Client
}

func NewUnreadTalkCache(rds *redis.Client) *UnreadTalkCache {
	return &UnreadTalkCache{rds}
}

func (c *UnreadTalkCache) key(sender, receive int) string {
	return fmt.Sprintf("%d_%d", sender, receive)
}
func (u *UnreadTalkCache) name(receive int) string {
	return fmt.Sprintf("talk:unread_msg:uid_%d", receive)
}

// Increment 消息未读数自增
// @params sender  发送者ID
// @params receive 接收者ID
func (c *UnreadTalkCache) Increment(ctx context.Context, sender, receive int) {
	c.rds.HIncrBy(ctx, "talk:unread:msg", c.key(sender, receive), 1)
}

// Get 获取消息未读数
// @params sender  发送者ID
// @params receive 接收者ID
func (c *UnreadTalkCache) Get(ctx context.Context, sender, receive int) int {
	val, _ := c.rds.HGet(ctx, "talk:unread:msg", c.key(sender, receive)).Int()

	return val
}

func (c *UnreadTalkCache) Reset(ctx context.Context, sender, receive int) {
	c.rds.HSet(ctx, "talk:unread:msg", c.key(sender, receive), 0)
}

//获取所有未读
func (u *UnreadTalkCache) GetAll(ctx context.Context, receive int) map[string]int {
	items := make(map[string]int)
	for k, v := range u.rds.HGetAll(ctx, u.name(receive)).Val() {
		items[k], _ = strconv.Atoi(v)
	}

	return items
}
