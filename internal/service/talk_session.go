package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-chat/internal/pkg/timeutil"
	"go-chat/internal/repository/dao"
	"go-chat/internal/repository/model"

	"gorm.io/gorm"
)

type TalkSessionService struct {
	*BaseService
	dao        *dao.TalkSessionDao
	contactDao *dao.ContactDao
}

func NewTalkSessionService(base *BaseService, dao *dao.TalkSessionDao, contactDao *dao.ContactDao) *TalkSessionService {
	return &TalkSessionService{base, dao, contactDao}
}

func (s *TalkSessionService) Dao() *dao.TalkSessionDao {
	return s.dao
}

func (s *TalkSessionService) List(ctx context.Context, uid int) ([]*model.SearchTalkSession, error) {
	var (
		err   error
		items = make([]*model.SearchTalkSession, 0)
	)
	member_type := s.contactDao.GetMemberType(ctx, uid)

	fields := []string{
		"list.id", "list.talk_type", "list.receiver_id", "list.updated_at",
		"list.is_disturb", "list.is_top", "list.is_robot",
		"`users`.avatar as user_avatar", "`users`.nickname",
		"`group`.group_name", "`group`.avatar as group_avatar",
	}

	query := s.db.Table("talk_session list")
	query.Joins("left join `users` ON list.receiver_id = `users`.id AND list.talk_type = 1")
	if member_type < 1 {
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
func (s *TalkSessionService) Create(ctx context.Context, opts *model.TalkSessionCreateOpt) (*model.TalkSession, error) {
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
	return s.db.Delete(&model.TalkSession{}, "id = ? and user_id = ?", id, uid).Error
}

type TalkSessionTopOpt struct {
	UserId int
	Id     int
	Type   int
}

// Top 会话置顶
func (s *TalkSessionService) Top(ctx context.Context, opts *TalkSessionTopOpt) error {

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

type TalkSessionDisturbOpt struct {
	UserId     int
	TalkType   int
	ReceiverId int
	IsDisturb  int
}

// Top 会话是否置顶
func (s *TalkSessionService) FindTalkSession(ctx context.Context, groupId int, uid int) (*model.TalkSession, error) {
	talkSession := &model.TalkSession{}
	if err := s.db.Model(&model.TalkSession{}).Where("receiver_id = ? and user_id = ? and talk_type = 2", groupId, uid).Limit(1).Scan(talkSession).Error; err != nil {
		return talkSession, err
	}
	return talkSession, nil
}

// Disturb 会话免打扰
func (s *TalkSessionService) Disturb(ctx context.Context, opts *TalkSessionDisturbOpt) error {
	err := s.db.Model(&model.TalkSession{}).
		Where("user_id = ? and receiver_id = ? and talk_type = ?", opts.UserId, opts.ReceiverId, opts.TalkType).
		Updates(map[string]interface{}{
			"is_disturb": opts.IsDisturb,
			"updated_at": time.Now(),
		}).Error

	return err
}

// BatchAddList 批量添加会话列表
func (s *TalkSessionService) BatchAddList(ctx context.Context, uid int, values map[string]int) {

	ctime := timeutil.DateTime()

	data := make([]string, 0)
	for k, v := range values {
		if v == 0 {
			continue
		}

		value := strings.Split(k, "_")
		if len(value) != 2 {
			continue
		}

		talkType, _ := strconv.Atoi(value[0])
		receiverId, _ := strconv.Atoi(value[1])

		data = append(data, fmt.Sprintf("(%d, %d, %d, '%s', '%s')", talkType, uid, receiverId, ctime, ctime))
	}

	if len(data) == 0 {
		return
	}

	sql := fmt.Sprintf("INSERT INTO talk_session ( `talk_type`, `user_id`, `receiver_id`, created_at, updated_at ) VALUES %s ON DUPLICATE KEY UPDATE is_delete = 0, updated_at = '%s';", strings.Join(data, ","), ctime)

	s.db.Exec(sql)
}
