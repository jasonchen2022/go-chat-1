package service

import (
	"errors"

	"go-chat/internal/dao"
	"go-chat/internal/model"

	"gorm.io/gorm"
)

type MemberService struct {
	dao *dao.MemberDao
}

func NewMemberService(memberDao *dao.MemberDao) *MemberService {
	return &MemberService{dao: memberDao}
}

// 根据用户ID查询数据
func (s *MemberService) FindById(userId int) (*model.Member, error) {
	member, err := s.dao.FindById(userId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("账号不存在! ")
		}

		return nil, err
	}
	return member, nil
}

// 查询管理员用户
func (s *MemberService) FindAdmin() ([]*model.Member, error) {
	userIds, err := s.dao.FindAdmin()
	if err != nil {
		return nil, err
	}
	return userIds, nil
}
