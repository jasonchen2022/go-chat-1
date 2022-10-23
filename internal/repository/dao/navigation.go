package dao

import (
	"go-chat/internal/repository/model"
)

type NavigationDao struct {
	*BaseDao
}

func NewNavigationDao(baseDao *BaseDao) *NavigationDao {
	return &NavigationDao{BaseDao: baseDao}
}

func (dao *NavigationDao) FindList() ([]*model.Navigation, error) {
	items := make([]*model.Navigation, 0)
	if err := dao.Db().Model(&model.Navigation{}).Select("id", "title", "sport_id", "logo", "url", "sort").Scan(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}
