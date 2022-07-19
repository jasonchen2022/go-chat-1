package dao

import (
	"context"
	"fmt"
	"time"

	"go-chat/internal/cache"
	"go-chat/internal/model"
)

type IContactDao interface {
	IBaseDao
	IsFriend(ctx context.Context, uid int, friendId int, cache bool) bool
	GetFriendRemarks(ctx context.Context, uid int, fids []int) (map[int]string, error)
	GetFriendRemark(ctx context.Context, uid int, friendId int, isCache bool) string
	SetFriendRemark(ctx context.Context, uid int, friendId int, remark string) error
}

type ContactDao struct {
	*BaseDao
	relation *cache.Relation
}

func NewContactDao(baseDao *BaseDao, relation *cache.Relation) *ContactDao {
	return &ContactDao{BaseDao: baseDao, relation: relation}
}

//判断当前发送者是否管理员
func (dao *ContactDao) IsLeader(userId int) bool {
	var member_type int
	dao.db.Table("users").Where("id = ?", userId).Select([]string{"type"}).Limit(1).Scan(&member_type)
	return member_type > 0
}

// IsFriend 判断是否为好友关系
func (dao *ContactDao) IsFriend(ctx context.Context, uid int, friendId int, cache bool) bool {

	if dao.IsLeader(friendId) || dao.IsLeader(uid) {
		return true
	}

	if cache && dao.relation.IsContactRelation(ctx, uid, friendId) == nil {
		return true
	}

	sql := `SELECT count(1) from contact where ((user_id = ? and friend_id = ?) or (user_id = ? and friend_id = ?)) and status = 1`

	var count int
	if err := dao.Db().Raw(sql, uid, friendId, friendId, uid).Scan(&count).Error; err != nil {
		return false
	}

	if count == 2 {
		dao.relation.SetContactRelation(ctx, uid, friendId)
	} else {
		dao.relation.DelContactRelation(ctx, uid, friendId)
	}

	return count == 2
}

func (dao *ContactDao) GetFriendRemarks(ctx context.Context, uid int, fids []int) (map[int]string, error) {
	if len(fids) == 0 {
		return map[int]string{}, nil
	}

	sql := `SELECT user_id, friend_id, remark FROM contact WHERE user_id = ? and friend_id in (?) and status = 1`

	var contacts []*model.Contact
	if err := dao.Db().Raw(sql, uid, fids).Scan(&contacts).Error; err != nil {
		return nil, err
	}

	items := make(map[int]string)
	for _, contact := range contacts {
		items[contact.FriendId] = contact.Remark
	}

	return items, nil
}

func (dao *ContactDao) GetFriendRemark(ctx context.Context, uid int, friendId int, isCache bool) string {

	if isCache {
		remark := dao.rds.HGet(ctx, fmt.Sprintf("rds:hash:friend-remark:%d", uid), fmt.Sprintf("%d_%d", uid, friendId)).Val()
		if remark != "" {
			return remark
		}
	}

	remark := ""
	err := dao.Db().Model(&model.Contact{}).Select("remark").Where("user_id = ? and friend_id = ?", uid, friendId).Scan(&remark).Error
	if err != nil {
		_ = dao.SetFriendRemark(ctx, uid, friendId, remark)
	}

	return remark
}

func (dao *ContactDao) SetFriendRemark(ctx context.Context, uid int, friendId int, remark string) error {
	err := dao.rds.HSet(ctx, fmt.Sprintf("rds:hash:friend-remark:%d", uid), fmt.Sprintf("%d_%d", uid, friendId), remark).Err()
	if err == nil {
		dao.rds.Expire(ctx, fmt.Sprintf("rds:hash:friend-remark:%d", uid), 72*time.Hour)
	}

	return err
}
