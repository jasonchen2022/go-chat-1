package dao

import (
	"go-chat/internal/model"
)

type MemberDao struct {
	*BaseDao
}

func NewMemberDao(baseDao *BaseDao) *MemberDao {
	return &MemberDao{BaseDao: baseDao}
}

// FindById ID查询
func (dao *MemberDao) FindById(userId int) (*model.Member, error) {
	member := &model.Member{}
	if err := dao.Db().Select("id", "mobile", "nickname", "avatar", "gender", "password", "motto").First(member, userId).Error; err != nil {
		return nil, err
	}
	return member, nil
}
