package service

import (
	"github.com/go-redis/redis/v8"
	"github.com/streadway/amqp"
	"gorm.io/gorm"
)

type BaseService struct {
	db  *gorm.DB
	rds *redis.Client
	mq  *amqp.Connection
}

func NewBaseService(db *gorm.DB, rds *redis.Client, mq *amqp.Connection) *BaseService {
	return &BaseService{db, rds, mq}
}

func (base *BaseService) Db() *gorm.DB {
	return base.db
}
