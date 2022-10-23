package service

import (
	"go-chat/internal/repository/dao"
	"go-chat/internal/repository/model"
)

type NavigationService struct {
	dao *dao.NavigationDao
}

func NewNavigationService(navigationDao *dao.NavigationDao) *NavigationService {
	return &NavigationService{dao: navigationDao}
}

func (s *NavigationService) FindList() ([]*model.Navigation, error) {
	items, err := s.dao.FindList()
	return items, err
}
