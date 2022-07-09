package service

import (
	"context"
	"go-chat/internal/model"
	"go-chat/internal/pkg/jsonutil"
	"time"

	match "github.com/dongweifly/sensitive-words-match"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type SensitiveMatchService struct {
	db  *gorm.DB
	rds *redis.Client
}

func NewSensitiveMatchService(db *gorm.DB, rds *redis.Client) *SensitiveMatchService {
	return &SensitiveMatchService{db: db, rds: rds}
}

func (s *SensitiveMatchService) GetService() *match.MatchService {
	ctx := context.Background()
	key := "sensitive-stop-words"
	json := s.rds.Get(ctx, key).Val()
	dict := make([]string, 0)
	service := match.NewMatchService()
	if json == "" {
		if s.db.Model(&model.DictData{}).Where("dict_id = 3 and status = 1").Select("code").Scan(&dict).Error == nil {
			s.rds.Set(ctx, key, jsonutil.Encode(dict), time.Duration(60*15)*time.Second)
		}
	} else {
		jsonutil.Decode(json, &dict)
	}
	if len(dict) > 0 {
		service.Build(dict)
	}
	return service
}