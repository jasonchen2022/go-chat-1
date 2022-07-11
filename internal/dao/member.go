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
	if err := dao.Db().Select("id", "type", "mobile", "nickname", "username", "avatar", "gender", "password", "motto").First(member, userId).Error; err != nil {
		return nil, err
	}
	return member, nil
}

// 查询管理员
func (dao *MemberDao) FindAdmin() ([]*model.Member, error) {
	members := make([]*model.Member, 0)
	if err := dao.Db().Model(&model.Member{}).Select("id", "type", "mobile", "nickname", "username", "avatar", "gender", "password", "motto").Where("type = 3 and status = 1 and mark = 1 ").Scan(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}
