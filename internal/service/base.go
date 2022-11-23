package service

import (
	"go-chat/config"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type BaseService struct {
	db     *gorm.DB
	rds    *redis.Client
	mq     rocketmq.Producer
	config *config.Config
}

func NewBaseService(db *gorm.DB, rds *redis.Client, mq rocketmq.Producer, config *config.Config) *BaseService {
	return &BaseService{db, rds, mq, config}
}

func (base *BaseService) Db() *gorm.DB {
	return base.db
}
