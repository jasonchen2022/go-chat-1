package dao

import (
	"context"
	"math/rand"
	"time"

	"go-chat/internal/repository/model"
)

type GroupDao struct {
	*BaseDao
}

func NewGroupDao(baseDao *BaseDao) *GroupDao {
	return &GroupDao{BaseDao: baseDao}
}

func (dao *GroupDao) FindById(id int) (*model.Group, error) {
	info := &model.Group{}

	if err := dao.Db().First(&info, id).Error; err != nil {
		return nil, err
	}

	return info, nil
}

func (dao *GroupDao) SearchOvertList(ctx context.Context, name string, page, size int) ([]*model.Group, error) {

	items := make([]*model.Group, 0)
	tx := dao.Db().Table("group")

	if name != "" {
		tx.Where("group_name LIKE ?", "%"+name+"%")
	}

	if err := tx.Where("group.type > ?", 0).Where("is_overt = ?", 1).Where("is_dismiss = 0").Scan(&items).Error; err != nil {
		return nil, err
	}

	if len(items) <= size {
		return items, nil
	}

	rand.Seed(time.Now().UnixNano())

	ids := make([]int, 0)
	//防止随机得到的id不存在  或者和当前用户一样的  循环次数定为100
	for i := 0; i < 100; i++ {
		//根据最大id进行随机
		random := rand.Intn(len(items))
		if !isValueInArr(random, ids) && len(ids) < size {
			ids = append(ids, random)
		}
	}

	resUsers := make([]*model.Group, 0)
	for _, v := range ids {
		resUsers = append(resUsers, items[v])
	}
	return resUsers, nil
}
