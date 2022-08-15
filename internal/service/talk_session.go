package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"go-chat/internal/dao"
	"go-chat/internal/model"
	"go-chat/internal/pkg/jsonutil"
)

type TalkSessionCreateOpts struct {
	UserId     int
	TalkType   int
	ReceiverId int
	IsBoot     bool
}

type TalkSessionTopOpts struct {
	UserId int
	Id     int
	Type   int
}

type TalkSessionDisturbOpts struct {
	UserId     int
	TalkType   int
	ReceiverId int
	IsDisturb  int
}

type TalkSessionService struct {
	*BaseService
	dao *dao.TalkSessionDao
}

func NewTalkSessionService(base *BaseService, dao *dao.TalkSessionDao) *TalkSessionService {
	return &TalkSessionService{base, dao}
}

func (s *TalkSessionService) Dao() *dao.TalkSessionDao {
	return s.dao
}

func (s *TalkSessionService) List(ctx context.Context, uid int) ([]*model.SearchTalkSession, error) {
	var (
		err   error
		items = make([]*model.SearchTalkSession, 0)
	)

	fields := []string{
		"list.id", "list.talk_type", "list.receiver_id", "list.updated_at",
		"list.is_disturb", "list.is_top", "list.is_robot",
		"`users`.avatar as user_avatar", "`users`.nickname",
		"`group`.group_name", "`group`.avatar as group_avatar",
	}
	user := &model.QueryUserTypeItem{}
	if err := s.db.Table("users").Where(&model.Users{Id: uid}).First(user).Error; err != nil {
		return nil, err
	}
	fmt.Printf("用户：%s", jsonutil.Encode(user))
	query := s.db.Table("talk_session list")
	query.Joins("left join `users` ON list.receiver_id = `users`.id AND list.talk_type = 1")
	//只有管理员用户才拥有聊天室权限
	if user.Type < 1 {
		query.Joins("left join `group` ON list.receiver_id = `group`.id AND list.talk_type = 2 and  group.type = 1")
	} else {
		query.Joins("left join `group` ON list.receiver_id = `group`.id AND list.talk_type = 2 and  group.type > 0")
	}
	query.Where("list.user_id = ? and list.is_delete = 0", uid)
	query.Order("list.updated_at desc")

	if err = query.Select(fields).Scan(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// Create 创建会话列表
func (s *TalkSessionService) Create(ctx context.Context, opts *TalkSessionCreateOpts) (*model.TalkSession, error) {
	var (
		err    error
		result *model.TalkSession
	)

	err = s.db.Where(&model.TalkSession{
		TalkType:   opts.TalkType,
		UserId:     opts.UserId,
		ReceiverId: opts.ReceiverId,
	}).First(&result).Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		result = &model.TalkSession{
			TalkType:   opts.TalkType,
			UserId:     opts.UserId,
			ReceiverId: opts.ReceiverId,
		}

		if opts.IsBoot {
			result.IsRobot = 1
		}

		s.db.Create(result)
	} else {
		result.IsTop = 0
		result.IsDelete = 0
		result.IsDisturb = 0

		if opts.IsBoot {
			result.IsRobot = 1
		}

		s.db.Save(result)
	}

	return result, nil
}

// Delete 删除会话
func (s *TalkSessionService) Delete(ctx context.Context, uid int, id int) error {
	return s.db.Model(&model.TalkSession{}).Where("id = ? and user_id = ?", id, uid).Updates(map[string]interface{}{
		"is_delete":  1,
		"updated_at": time.Now(),
	}).Error
}

// Top 会话置顶
func (s *TalkSessionService) Top(ctx context.Context, opts *TalkSessionTopOpts) error {

	isTop := 0

	if opts.Type == 1 {
		isTop = 1
	}

	err := s.db.Model(&model.TalkSession{}).Where("id = ? and user_id = ?", opts.Id, opts.UserId).
		Updates(map[string]interface{}{
			"is_top":     isTop,
			"updated_at": time.Now(),
		}).Error

	return err
}

// Top 会话是否置顶
func (s *TalkSessionService) FindTalkSession(ctx context.Context, groupId int, uid int) (*model.TalkSession, error) {
	talkSession := &model.TalkSession{}
	if err := s.db.Model(&model.TalkSession{}).Where("id = ? and user_id = ?", groupId, uid).Limit(1).Scan(talkSession).Error; err != nil {
		return talkSession, err
	}
	return talkSession, nil
}

// Disturb 会话免打扰
func (s *TalkSessionService) Disturb(ctx context.Context, opts *TalkSessionDisturbOpts) error {
	err := s.db.Model(&model.TalkSession{}).
		Where("user_id = ? and receiver_id = ? and talk_type = ?", opts.UserId, opts.ReceiverId, opts.TalkType).
		Updates(map[string]interface{}{
			"is_disturb": opts.IsDisturb,
			"updated_at": time.Now(),
		}).Error

	return err
}
