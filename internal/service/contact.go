package service

import (
	"context"
	"math"

	"go-chat/internal/dao"
	"go-chat/internal/model"

	"gorm.io/gorm"
)

type ContactService struct {
	*BaseService
	dao *dao.ContactDao
}

func NewContactService(baseService *BaseService, dao *dao.ContactDao) *ContactService {
	return &ContactService{BaseService: baseService, dao: dao}
}

func (s *ContactService) Dao() dao.IContactDao {
	return s.dao
}

// EditRemark 编辑联系人备注
// @params uid      用户ID
// @params friendId 联系人ID
func (s *ContactService) EditRemark(ctx context.Context, uid int, friendId int, remark string) error {
	err := s.db.Model(&model.Contact{}).Where("user_id = ? and friend_id = ?", uid, friendId).Update("remark", remark).Error

	_ = s.dao.SetFriendRemark(ctx, uid, friendId, remark)

	return err
}

// Delete 删除联系人
// @params uid      用户ID
// @params friendId 联系人ID
func (s *ContactService) Delete(ctx context.Context, uid, friendId int) error {
	return s.db.Model(&model.Contact{}).Where("user_id = ? and friend_id = ?", uid, friendId).Update("status", 0).Error
}

// List 获取联系人列表
// @params uid      用户ID
func (s *ContactService) List(ctx context.Context, uid int) ([]*model.ContactListItem, error) {

	tx := s.db.Model(&model.Contact{})
	tx.Select([]string{
		"users.id",
		"users.nickname",
		"users.avatar",
		"users.motto",
		"users.gender",
		"contact.remark",
	})

	tx.Joins("inner join `users` ON `users`.id = contact.friend_id")
	tx.Where("`contact`.user_id = ? and contact.status = ?", uid, 1)

	items := make([]*model.ContactListItem, 0)
	if err := tx.Scan(&items).Error; err != nil {
		return nil, err
	}

	return items, nil
}

func (s *ContactService) ListByPage(ctx context.Context, uid int, page int) ([]*model.ContactListItem, error) {

	pageIndex := (page - 1) * 20
	tx := s.db.Model(&model.Contact{})
	tx.Select([]string{
		"users.id",
		"users.nickname",
		"users.avatar",
		"users.motto",
		"users.gender",
		"contact.remark",
	})

	tx.Joins("inner join `users` ON `users`.id = contact.friend_id")
	tx.Where("`contact`.user_id = ? and contact.status = ? LIMIT ?,20;", uid, 1, pageIndex)

	items := make([]*model.ContactListItem, 0)
	if err := tx.Scan(&items).Error; err != nil {
		return nil, err
	}

	return items, nil
}

func (s *ContactService) TotalPage(ctx context.Context, uid int) (int, error) {

	tx := s.db.Model(&model.Contact{})
	tx.Select([]string{
		"users.id",
		"users.nickname",
		"users.avatar",
		"users.motto",
		"users.gender",
		"contact.remark",
	})

	tx.Joins("inner join `users` ON `users`.id = contact.friend_id")
	tx.Where("`contact`.user_id = ? and contact.status = ? ", uid, 1)

	items := make([]*model.ContactListItem, 0)
	if err := tx.Scan(&items).Error; err != nil {
		return 0, err
	}
	//    page:=math.Ceil(len(items)/50)
	page := int(math.Ceil(float64(len(items)) / float64(20)))
	return page, nil
}

func (s *ContactService) GetContactIds(ctx context.Context, uid int) []int64 {
	ids := make([]int64, 0)

	s.db.Model(&model.Contact{}).Where("user_id = ? and status = ?", uid, 1).Pluck("friend_id", &ids)

	return ids
}

//建立好友关系
func (s *ContactService) Create(ctx context.Context, opts *ContactApplyCreateOpts) error {
	apply := &model.Contact{
		UserId:   opts.UserId,
		FriendId: opts.FriendId,
		Remark:   opts.Remarks,
		Status:   1,
	}

	if err := s.db.Create(apply).Error; err != nil {
		return err
	}
	return nil
}

//添加11直播官方为双向好友
func (s *ContactService) AddCustomerFriend(ctx context.Context, uid int) error {

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if !(s.Dao().IsFriend(ctx, uid, 7715, false)) {
			apply := &model.Contact{
				UserId:   uid,
				FriendId: 7715,
				Status:   1,
			}

			if err := s.db.Create(apply).Error; err != nil {
				return err
			}

			apply_friend := &model.Contact{
				UserId:   7715,
				FriendId: uid,
				Status:   1,
			}

			if err := s.db.Create(apply_friend).Error; err != nil {
				return err
			}

		}
		return nil
	})
	return err
}
