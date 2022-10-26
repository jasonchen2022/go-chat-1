package dao

import (
	"go-chat/internal/repository/model"
	"math/rand"
	"strconv"
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
	if err := dao.Db().Model(&model.Users{}).Where("id in ?", userIds).Scan(&user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// FindByMobile 手机号查询
func (dao *UsersDao) FindByMobile(mobile string) (*model.Users, error) {
	user := &model.Users{}

	if err := dao.Db().Where(&model.Users{Mobile: mobile}).Or(&model.Users{Nickname: mobile}).First(user).Error; err != nil {
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
func (dao *UsersDao) RandomUser(userId, index int, userName string) ([]*model.UserTemp, error) {

	//查出当前用户关注过的主播
	anchors := make([]*model.Fans, 0)
	dao.Db().Model(&model.Fans{}).Where("mark = ?", 1).Scan(&anchors)

	// fmt.Println(jsonutil.Encode(anchors))
	users := make([]*model.UserTemp, 0)
	if userName == "" {
		//只随机主播  type=1
		if err := dao.Db().Model(&model.Users{}).Where("type = ?", 1).Where("Id <> ?", userId).Scan(&users).Error; err != nil {
			return nil, err
		}
	} else {
		//只随机主播  type=0
		value, err := strconv.Atoi(userName)
		if err == nil {
			if len(userName) > 10 {
				// if err := dao.Db().Model(&model.Users{}).Where("type != ?", -1).Where("Id <> ?", userId).Where("mobile = ?", value).Scan(&users).Error; err != nil {
				// 	return nil, err
				// }
				if err := dao.Db().Model(&model.Users{}).Where("type != ?", -1).Where("mobile = ?", value).Scan(&users).Error; err != nil {
					return nil, err
				}
			} else {
				// if err := dao.Db().Model(&model.Users{}).Where("type != ?", -1).Where("Id <> ?", userId).Where("id = ?", value).Scan(&users).Error; err != nil {
				// 	return nil, err
				// }
				if err := dao.Db().Model(&model.Users{}).Where("type != ?", -1).Where("member_id = ?", value).Scan(&users).Error; err != nil {
					return nil, err
				}
			}

		} else {
			// if err := dao.Db().Model(&model.Users{}).Where("type != ?", -1).Where("Id <> ?", userId).Where("nickname like ?", "%"+userName+"%").Scan(&users).Error; err != nil {
			// 	return nil, err
			// }
			if err := dao.Db().Model(&model.Users{}).Where("type != ?", -1).Where("nickname like ?", "%"+userName+"%").Scan(&users).Error; err != nil {
				return nil, err
			}
		}

	}

	for _, v := range users {
		v.FansCount = fansCount(anchors, v.Id)
	}
	// fmt.Println(jsonutil.Encode(users))
	if len(users) <= 6 {
		return users, nil
	}

	ids := make([]int, 0)

	rand.Seed(time.Now().UnixNano())

	//防止随机得到的id不存在  或者和当前用户一样的  循环次数定为100
	for i := 0; i < 100; i++ {
		//根据最大id进行随机
		random := rand.Intn(len(users))
		if !isValueInArr(random, ids) && len(ids) < index {
			ids = append(ids, random)
		}
	}

	resUsers := make([]*model.UserTemp, 0)
	for _, v := range ids {
		if isGuanZhu(anchors, users[v].Id) {
			users[v].IsGaunZhu = 1
		} else {
			users[v].IsGaunZhu = 0
		}
		resUsers = append(resUsers, users[v])
	}

	// fmt.Println(len(users))
	// fmt.Println(jsonutil.Encode(resUsers))

	return resUsers, nil
}

//判断数组中是否包含某个值
func isValueInArr(value int, arr []int) bool {
	for _, v := range arr {
		if v == value {
			return true
		}
	}
	return false
}

//判断是否关注过
func isGuanZhu(anchors []*model.Fans, anchorId int) bool {
	for _, v := range anchors {
		if v.AnchorId == anchorId {
			return true
		}
	}
	return false
}

func fansCount(anchors []*model.Fans, user_id int) int {
	i := 0
	for _, v := range anchors {
		if v.AnchorId == user_id {
			i = i + 1
		}
	}
	return i
}
