package dao

import (
	"fmt"
	"go-chat/internal/model"
	"go-chat/internal/pkg/jsonutil"
	"math/rand"
	"time"
)

type UsersDao struct {
	*BaseDao
}

func NewUserDao(baseDao *BaseDao) *UsersDao {
	return &UsersDao{BaseDao: baseDao}
}

// Create 创建数据
func (dao *UsersDao) Create(user *model.Users) (*model.Users, error) {
	if err := dao.Db().Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// FindById ID查询
func (dao *UsersDao) FindById(userId int) (*model.Users, error) {
	user := &model.Users{}

	if err := dao.Db().Where(&model.Users{Id: userId}).First(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// GetNickName ID查询
func (dao *UsersDao) GetNickName(userId int) (string, error) {
	var nickname string
	if err := dao.Db().Table("users").Where(&model.Users{Id: userId}).Select("nickname").Limit(1).Scan(&nickname).Error; err != nil {
		return "", err
	}
	return nickname, nil
}

// FindByIds ID查询
func (dao *UsersDao) FindByIds(userIds []int) ([]*model.Users, error) {
	user := make([]*model.Users, 0)

	if err := dao.Db().Model(&model.Users{}).Where("id in ?", userIds).Scan(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// FindByMobile 手机号查询
func (dao *UsersDao) FindByMobile(mobile string) (*model.Users, error) {
	user := &model.Users{}

	if err := dao.Db().Where(&model.Users{Mobile: mobile}).First(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// IsMobileExist 判断手机号是否存在
func (dao *UsersDao) IsMobileExist(mobile string) bool {
	user := &model.Users{}

	rowsAffects := dao.Db().Select("id").Where(&model.Users{Mobile: mobile}).First(user).RowsAffected

	return rowsAffects != 0
}

/*
*发现好友  （除登录用户外）
*userId:登录用户id
*index:查询用户数
 */
func (dao *UsersDao) RandomUser(userId, index int) ([]*model.Users, error) {

	users := make([]*model.Users, 0)
	if err := dao.Db().Model(&model.Users{}).Where("type = ?", 1).Scan(&users).Error; err != nil {
		return nil, err
	}
	// fmt.Println(jsonutil.Encode(users))
	if len(users) <= 6 {
		return users, nil
	}

	ids := make([]int, 0)

	rand.Seed(time.Now().UnixNano())

	//防止随机得到的id不存在  或者和当前用户一样的  循环次数定为100
	for i := 0; i < index; i++ {
		//根据最大id进行随机
		random := rand.Intn(len(users))
		if !isValueInArr(random, ids) {
			ids = append(ids, random)
		}
	}

	resUsers := make([]*model.Users, 0)
	for _, v := range ids {
		resUsers = append(resUsers, users[v])
	}

	// fmt.Println(len(users))
	fmt.Println(jsonutil.Encode(resUsers))

	return resUsers, nil
}

// func (dao *UsersDao) RandomUser(userId, index int) ([]*model.Users, error) {

// 	userNew := &model.Users{}
// 	//查出表中  最大id
// 	dao.Db().Last(userNew)

// 	ids := make([]int, 0)

// 	rand.Seed(time.Now().UnixNano())

// 	users := make([]*model.Users, 0)
// 	//防止随机得到的id不存在  或者和当前用户一样的  循环次数定为100
// 	for i := 0; i < 100; i++ {
// 		//根据最大id进行随机
// 		random := rand.Intn(userNew.Id)
// 		if random != userId && !isValueInArr(random, ids) {
// 			ids = append(ids, random)
// 		}
// 	}

// 	if err := dao.Db().Debug().Model(&model.Users{}).Where("id in ?", ids).Scan(&users).Error; err != nil {
// 		return nil, err
// 	}

// 	resUsers := make([]*model.Users, 0)

// 	for _, value := range users {
// 		if len(resUsers) < index {
// 			resUsers = append(resUsers, value)
// 		}
// 		if len(resUsers) == index {
// 			break
// 		}
// 	}

// 	// fmt.Println("@@@@====index=", index)
// 	// fmt.Println("@@@@====", jsonutil.Encode(resUsers))
// 	return resUsers, nil
// }

//判断数组中是否包含某个值
func isValueInArr(value int, arr []int) bool {
	for _, v := range arr {
		if v == value {
			return true
		}
	}
	return false
}
