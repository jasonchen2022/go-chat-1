package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	"go-chat/config"
	"go-chat/internal/provider"
	model2 "go-chat/internal/repository/model"

	"gorm.io/gorm"

	"go-chat/internal/entity"
	"go-chat/internal/pkg/jsonutil"
)

type ContactApplyCreateOpts struct {
	UserId   int
	Remarks  string
	FriendId int
}

type ContactApplyAcceptOpts struct {
	UserId  int
	Remarks string
	ApplyId int
}

type ContactApplyDeclineOpts struct {
	UserId  int
	Remarks string
	ApplyId int
}

type ContactApplyService struct {
	*BaseService
	talkMessage *TalkMessageService
}

func NewContactsApplyService(base *BaseService, talkMessage *TalkMessageService) *ContactApplyService {
	return &ContactApplyService{BaseService: base, talkMessage: talkMessage}
}

func (s *ContactApplyService) Create(ctx context.Context, opts *ContactApplyCreateOpts) error {

	apply := &model2.ContactApply{
		UserId:   opts.UserId,
		FriendId: opts.FriendId,
		Remark:   opts.Remarks,
	}

	if err := s.db.Create(apply).Error; err != nil {
		return err
	}

	body := map[string]interface{}{
		"event": entity.EventContactApply,
		"data": jsonutil.Encode(map[string]interface{}{
			"apply_id": int64(apply.Id),
			"type":     1,
		}),
	}
	if s.mq == nil {
		conf := config.ReadConfig(config.ParseConfigArg())
		s.mq = provider.NewRabbitMQClient(ctx, conf)
		log.Println("Failed to open a channel:", "并重新初始化")
	}
	// 创建一个Channel
	channel, err := s.mq.Channel()
	if err != nil {
		log.Println("Failed to open a channel:", err.Error())

	}
	defer channel.Close()

	// 声明exchange
	if err := channel.ExchangeDeclare(
		s.config.RabbitMQ.ExchangeName, //name
		"fanout",                       //exchangeType
		true,                           //durable
		false,                          //auto-deleted
		false,                          //internal
		false,                          //noWait
		nil,                            //arguments
	); err != nil {
		log.Println("Failed to declare a exchange:", err.Error())
	}

	s.talkMessage.SendAll(channel, jsonutil.Encode(body))

	return nil
}

// Accept 同意好友申请
func (s *ContactApplyService) Accept(ctx context.Context, opts *ContactApplyAcceptOpts) (*model2.ContactApply, error) {
	var (
		err       error
		applyInfo *model2.ContactApply
	)

	if err := s.db.First(&applyInfo, "id = ? and friend_id = ?", opts.ApplyId, opts.UserId).Error; err != nil {
		return nil, err
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		addFriendFunc := func(uid, fid int, remark string) error {
			var friends *model2.Contact

			err = tx.Where("user_id = ? and friend_id = ?", uid, fid).First(&friends).Error

			// 数据存在则更新
			if err == nil {
				return tx.Model(&model2.Contact{}).Where("id = ?", friends.Id).Updates(&model2.Contact{
					Remark: remark,
					Status: 1,
				}).Error
			}

			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}

			return tx.Create(&model2.Contact{
				UserId:   uid,
				FriendId: fid,
				Remark:   remark,
				Status:   1,
			}).Error
		}

		var user *model2.Users
		if err := tx.Select("id", "nickname").First(&user, applyInfo.FriendId).Error; err != nil {
			return err
		}

		if err := addFriendFunc(applyInfo.UserId, applyInfo.FriendId, user.Nickname); err != nil {
			return err
		}

		if err := addFriendFunc(applyInfo.FriendId, applyInfo.UserId, opts.Remarks); err != nil {
			return err
		}

		return tx.Delete(&model2.ContactApply{}, "user_id = ? and friend_id = ?", applyInfo.UserId, applyInfo.FriendId).Error
	})

	return applyInfo, err
}

// Decline 拒绝好友申请
func (s *ContactApplyService) Decline(ctx context.Context, opts *ContactApplyDeclineOpts) error {
	err := s.db.Delete(&model2.ContactApply{}, "id = ? and friend_id = ?", opts.ApplyId, opts.UserId).Error

	if err == nil {
		body := map[string]interface{}{
			"event": entity.EventContactApply,
			"data": jsonutil.Encode(map[string]interface{}{
				"apply_id": int64(opts.ApplyId),
				"type":     2,
			}),
		}

		// 创建一个Channel
		channel, err := s.mq.Channel()
		if err != nil {
			log.Println("Failed to open a channel:", err.Error())

		}
		defer channel.Close()

		// 声明exchange
		if err := channel.ExchangeDeclare(
			s.config.RabbitMQ.ExchangeName, //name
			"fanout",                       //exchangeType
			true,                           //durable
			false,                          //auto-deleted
			false,                          //internal
			false,                          //noWait
			nil,                            //arguments
		); err != nil {
			log.Println("Failed to declare a exchange:", err.Error())
		}
		s.talkMessage.SendAll(channel, jsonutil.Encode(body))

	}

	return err
}

// List 联系人申请列表
func (s *ContactApplyService) List(ctx context.Context, uid, page, size int) ([]*model2.ApplyItem, error) {
	fields := []string{
		"contact_apply.id",
		"contact_apply.remark",
		"users.nickname",
		"users.avatar",
		"users.mobile",
		"contact_apply.user_id",
		"contact_apply.friend_id",
		"contact_apply.created_at",
	}

	tx := s.db.Debug().Table("contact_apply")
	tx.Joins("left join `users` ON `users`.id = contact_apply.user_id")
	tx.Where("contact_apply.friend_id = ?", uid)
	tx.Order("contact_apply.id desc")

	items := make([]*model2.ApplyItem, 0)
	if err := tx.Select(fields).Scan(&items).Error; err != nil {
		return nil, err
	}

	return items, nil
}

func (s *ContactApplyService) GetApplyUnreadNum(ctx context.Context, uid int) int {

	num, err := s.rds.Get(ctx, fmt.Sprintf("friend-apply:user_%d", uid)).Int()
	if err != nil {
		return 0
	}

	return num
}

func (s *ContactApplyService) ClearApplyUnreadNum(ctx context.Context, uid int) {
	s.rds.Del(ctx, fmt.Sprintf("friend-apply:user_%d", uid))
}
