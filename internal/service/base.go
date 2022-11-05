package service

import (
	"go-chat/config"

	"github.com/go-redis/redis/v8"
	"github.com/streadway/amqp"
	"gorm.io/gorm"
)

type BaseService struct {
	db     *gorm.DB
	rds    *redis.Client
	mq     *amqp.Connection
	config *config.Config
}

func NewBaseService(db *gorm.DB, rds *redis.Client, mq *amqp.Connection, config *config.Config) *BaseService {
	return &BaseService{db, rds, mq, config}
}

func (base *BaseService) Db() *gorm.DB {
	return base.db
}
