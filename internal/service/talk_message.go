package service

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/GUAIK-ORG/go-snowflake/snowflake"
	"gorm.io/gorm"

	"go-chat/config"
	"go-chat/internal/cache"
	"go-chat/internal/dao"
	"go-chat/internal/entity"
	"go-chat/internal/model"
	"go-chat/internal/pkg/filesystem"
	"go-chat/internal/pkg/jsonutil"
	"go-chat/internal/pkg/strutil"
	"go-chat/internal/pkg/timeutil"
)

type SysTextMessageOpts struct {
	UserId     int
	TalkType   int
	ReceiverId int
	Text       string
}

type TextMessageOpts struct {
	UserId     int
	TalkType   int
	ReceiverId int
	Text       string
}

type LoginMessageOpts struct {
	UserId   int
	Ip       string
	Address  string
	Platform string
	Agent    string
}

type FileMessageOpts struct {
	UserId     int
	TalkType   int
	ReceiverId int
	File       *multipart.FileHeader
}

type ImageMessageOpts struct {
	UserId     int
	TalkType   int
	ReceiverId int
	File       *multipart.FileHeader
}

type LocationMessageOpts struct {
	UserId     int
	TalkType   int
	ReceiverId int
	Longitude  string
	Latitude   string
}

type CodeMessageOpts struct {
	UserId     int
	TalkType   int
	ReceiverId int
	Lang       string
	Code       string
}

type CardMessageOpts struct {
	UserId     int
	TalkType   int
	ReceiverId int
	ContactId  int
}

type VoteMessageOpts struct {
	UserId     int
	ReceiverId int
	Mode       int
	Anonymous  int
	Title      string
	Options    []string
}

type EmoticonMessageOpts struct {
	UserId     int
	TalkType   int
	ReceiverId int
	EmoticonId int
}

type VoteMessageHandleOpts struct {
	UserId   int
	RecordId int
	Options  string
}

type TalkMessageService struct {
	*BaseService
	config             *config.Config
	unreadTalkCache    *cache.UnreadTalkCache
	lastMessage        *cache.LastMessage
	talkRecordsVoteDao *dao.TalkRecordsVoteDao
	groupMemberDao     *dao.GroupMemberDao
	sidServer          *cache.SidServer
	client             *cache.WsClientSession
	fileSystem         *filesystem.Filesystem
	splitUploadDao     *dao.SplitUploadDao
}

func NewTalkMessageService(baseService *BaseService, config *config.Config, unreadTalkCache *cache.UnreadTalkCache, lastMessage *cache.LastMessage, talkRecordsVoteDao *dao.TalkRecordsVoteDao, groupMemberDao *dao.GroupMemberDao, sidServer *cache.SidServer, client *cache.WsClientSession, fileSystem *filesystem.Filesystem, splitUploadDao *dao.SplitUploadDao) *TalkMessageService {
	return &TalkMessageService{BaseService: baseService, config: config, unreadTalkCache: unreadTalkCache, lastMessage: lastMessage, talkRecordsVoteDao: talkRecordsVoteDao, groupMemberDao: groupMemberDao, sidServer: sidServer, client: client, fileSystem: fileSystem, splitUploadDao: splitUploadDao}
}

// SendSysMessage 发送文本消息
func (s *TalkMessageService) SendSysMessage(ctx context.Context, opts *SysTextMessageOpts) error {
	record := &model.TalkRecords{
		TalkType:   opts.TalkType,
		MsgType:    entity.MsgTypeSystemText,
		UserId:     opts.UserId,
		ReceiverId: opts.ReceiverId,
		Content:    opts.Text,
	}

	if err := s.db.Debug().Create(record).Error; err != nil {
		return err
	}

	s.afterHandle(ctx, record, map[string]string{
		"text": strutil.MtSubstr(record.Content, 0, 30),
	})

	return nil
}

// SendTextMessage 发送文本消息
func (s *TalkMessageService) SendTextMessage(ctx context.Context, opts *TextMessageOpts) error {
	record := &model.TalkRecords{
		TalkType:   opts.TalkType,
		MsgType:    entity.MsgTypeText,
		UserId:     opts.UserId,
		ReceiverId: opts.ReceiverId,
		Content:    opts.Text,
	}

	//校验权限
	c := s.checkUserAuth(record.UserId)
	if c != nil {
		return c
	}

	if err := s.db.Create(record).Error; err != nil {
		return err
	}

	if e := s.afterHandle(ctx, record, map[string]string{
		"text": strutil.MtSubstr(record.Content, 0, 30),
	}); e != nil {
		return e
	}

	return nil
}

// SendCodeMessage 发送代码消息
func (s *TalkMessageService) SendCodeMessage(ctx context.Context, opts *CodeMessageOpts) error {
	var (
		err    error
		record = &model.TalkRecords{
			TalkType:   opts.TalkType,
			MsgType:    entity.MsgTypeCode,
			UserId:     opts.UserId,
			ReceiverId: opts.ReceiverId,
		}
	)

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err = s.db.Create(record).Error; err != nil {
			return err
		}

		if err = s.db.Create(&model.TalkRecordsCode{
			RecordId: record.Id,
			UserId:   opts.UserId,
			Lang:     opts.Lang,
			Code:     opts.Code,
		}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	if e := s.afterHandle(ctx, record, map[string]string{"text": "[代码消息]"}); e != nil {
		return e
	}

	return nil
}

// SendImageMessage 发送图片消息
func (s *TalkMessageService) SendImageMessage(ctx context.Context, opts *ImageMessageOpts) error {
	var (
		err    error
		record = &model.TalkRecords{
			TalkType:   opts.TalkType,
			MsgType:    entity.MsgTypeFile,
			UserId:     opts.UserId,
			ReceiverId: opts.ReceiverId,
		}
	)
	//校验权限
	c := s.checkUserAuth(record.UserId)
	if c != nil {
		return c
	}
	stream, err := filesystem.ReadMultipartStream(opts.File)
	if err != nil {
		return err
	}

	ext := strutil.FileSuffix(opts.File.Filename)
	sn, _ := snowflake.NewSnowflake(int64(0), int64(0))
	val := sn.NextVal()
	fileName := fmt.Sprintf("chat/image/%s/%s%s", time.Now().Format("20060102"), strconv.FormatInt(val, 10), ext)

	if err := s.fileSystem.Oss.UploadByte(fileName, stream); err != nil {
		return err
	}

	filePath := s.fileSystem.Oss.PublicUrl(fileName)

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err = s.db.Create(record).Error; err != nil {
			return err
		}

		if err = s.db.Create(&model.TalkRecordsFile{
			RecordId:     record.Id,
			UserId:       opts.UserId,
			Source:       1,
			Type:         entity.GetMediaType(ext),
			Drive:        entity.FileDriveMode(s.fileSystem.Driver()),
			OriginalName: opts.File.Filename,
			Suffix:       ext,
			Size:         int(opts.File.Size),
			Path:         filePath,
			Url:          filePath,
		}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	if e := s.afterHandle(ctx, record, map[string]string{"text": "[图片消息]"}); e != nil {
		return e
	}

	return nil
}

// SendFileMessage 发送文件消息
func (s *TalkMessageService) SendFileMessage(ctx context.Context, opts *FileMessageOpts) error {

	var (
		err    error
		record = &model.TalkRecords{
			TalkType:   opts.TalkType,
			MsgType:    entity.MsgTypeFile,
			UserId:     opts.UserId,
			ReceiverId: opts.ReceiverId,
		}
	)

	//校验权限
	c := s.checkUserAuth(record.UserId)
	if c != nil {
		return c
	}

	stream, err := filesystem.ReadMultipartStream(opts.File)
	if err != nil {
		return err
	}

	ext := strutil.FileSuffix(opts.File.Filename)
	sn, _ := snowflake.NewSnowflake(int64(0), int64(0))
	val := sn.NextVal()
	fileName := fmt.Sprintf("chat/file/%s/%s%s", time.Now().Format("20060102"), strconv.FormatInt(val, 10), ext)

	if err := s.fileSystem.Oss.UploadByte(fileName, stream); err != nil {
		return err
	}

	filePath := s.fileSystem.Oss.PublicUrl(fileName)

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err = s.db.Create(record).Error; err != nil {
			return err
		}

		if err = s.db.Create(&model.TalkRecordsFile{
			RecordId:     record.Id,
			UserId:       opts.UserId,
			Source:       1,
			Type:         entity.GetMediaType(ext),
			Drive:        entity.FileDriveMode(s.fileSystem.Driver()),
			OriginalName: opts.File.Filename,
			Suffix:       ext,
			Size:         int(opts.File.Size),
			Path:         filePath,
			Url:          filePath,
		}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	if e := s.afterHandle(ctx, record, map[string]string{"text": "[文件消息]"}); e != nil {
		return e
	}

	return nil
}

// SendCardMessage 发送用户名片消息
func (s *TalkMessageService) SendCardMessage(ctx context.Context, opts *CardMessageOpts) error {
	// todo 发送用户名片消息待开发
	return nil
}

// SendVoteMessage 发送投票消息
func (s *TalkMessageService) SendVoteMessage(ctx context.Context, opts *VoteMessageOpts) error {
	var (
		err    error
		record = &model.TalkRecords{
			TalkType:   entity.ChatGroupMode,
			MsgType:    entity.MsgTypeVote,
			UserId:     opts.UserId,
			ReceiverId: opts.ReceiverId,
		}
	)

	options := make(map[string]string)
	for i, value := range opts.Options {
		options[fmt.Sprintf("%c", 65+i)] = value
	}

	num := s.groupMemberDao.CountMemberTotal(opts.ReceiverId)

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err = s.db.Create(record).Error; err != nil {
			return err
		}

		if err = s.db.Create(&model.TalkRecordsVote{
			RecordId:     record.Id,
			UserId:       opts.UserId,
			Title:        opts.Title,
			AnswerMode:   opts.Mode,
			AnswerOption: jsonutil.Encode(options),
			AnswerNum:    int(num),
			IsAnonymous:  opts.Anonymous,
		}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	s.afterHandle(ctx, record, map[string]string{"text": "[投票消息]"})

	return nil
}

// SendEmoticonMessage 发送表情包消息
func (s *TalkMessageService) SendEmoticonMessage(ctx context.Context, opts *EmoticonMessageOpts) error {
	var (
		err      error
		emoticon model.EmoticonItem
		record   = &model.TalkRecords{
			TalkType:   opts.TalkType,
			MsgType:    entity.MsgTypeFile,
			UserId:     opts.UserId,
			ReceiverId: opts.ReceiverId,
		}
	)

	if err = s.db.Model(&model.EmoticonItem{}).Where("id = ?", opts.EmoticonId).First(&emoticon).Error; err != nil {
		return err
	}

	if emoticon.UserId > 0 && emoticon.UserId != opts.UserId {
		return errors.New("表情包不存在！")
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err = s.db.Create(record).Error; err != nil {
			return err
		}

		if err = s.db.Create(&model.TalkRecordsFile{
			RecordId:     record.Id,
			UserId:       opts.UserId,
			Source:       2,
			Type:         entity.GetMediaType(emoticon.FileSuffix),
			OriginalName: "图片表情",
			Suffix:       emoticon.FileSuffix,
			Size:         emoticon.FileSize,
			Path:         emoticon.Url,
			Url:          emoticon.Url,
		}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	s.afterHandle(ctx, record, map[string]string{"text": "[图片消息]"})

	return nil
}

// SendLocationMessage 发送位置消息
func (s *TalkMessageService) SendLocationMessage(ctx context.Context, opts *LocationMessageOpts) error {

	var (
		err    error
		record = &model.TalkRecords{
			TalkType:   opts.TalkType,
			MsgType:    entity.MsgTypeLocation,
			UserId:     opts.UserId,
			ReceiverId: opts.ReceiverId,
		}
	)

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err = s.db.Create(record).Error; err != nil {
			return err
		}

		if err = s.db.Create(&model.TalkRecordsLocation{
			RecordId:  record.Id,
			UserId:    opts.UserId,
			Longitude: opts.Longitude,
			Latitude:  opts.Latitude,
		}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	s.afterHandle(ctx, record, map[string]string{"text": "[位置消息]"})

	return nil
}

// SendRevokeRecordMessage 撤销推送消息
func (s *TalkMessageService) SendRevokeRecordMessage(ctx context.Context, uid int, recordId int) error {
	var (
		err    error
		record model.TalkRecords
	)

	if err = s.db.First(&record, recordId).Error; err != nil {
		return err
	}

	if record.IsRevoke == 1 {
		return nil
	}
	//私聊只能撤回自己发的消息
	if record.TalkType == 1 {
		if record.UserId != uid {
			return errors.New("无权撤回消息")
		}
	}
	//如果是群聊，管理员可以撤回所有人发的消息
	if record.TalkType == 2 {
		if !(s.groupMemberDao.IsMember(record.ReceiverId, uid, true)) {
			return errors.New("无权撤回群聊消息")
		}
	}
	///无时间限制
	// if time.Now().Unix() > record.CreatedAt.Add(3*time.Minute).Unix() {
	// 	return errors.New("超出有效撤回时间范围，无法进行撤销！")
	// }

	if err = s.db.Model(&model.TalkRecords{Id: recordId}).Update("is_revoke", 1).Error; err != nil {
		return err
	}

	body := map[string]interface{}{
		"event": entity.EventTalkRevoke,
		"data": jsonutil.Encode(map[string]interface{}{
			"record_id": record.Id,
		}),
	}

	s.rds.Publish(ctx, entity.IMGatewayAll, jsonutil.Encode(body))

	return nil
}

// VoteHandle 投票处理
func (s *TalkMessageService) VoteHandle(ctx context.Context, opts *VoteMessageHandleOpts) (int, error) {
	var (
		err  error
		vote *model.QueryVoteModel
	)

	tx := s.db.Table("talk_records")
	tx.Select([]string{
		"talk_records.receiver_id", "talk_records.talk_type", "talk_records.msg_type",
		"vote.id as vote_id", "vote.id as record_id", "vote.answer_mode", "vote.answer_option",
		"vote.answer_num", "vote.status as vote_status",
	})
	tx.Joins("left join talk_records_vote as vote on vote.record_id = talk_records.id")
	tx.Where("talk_records.id = ?", opts.RecordId)

	res := tx.Take(&vote)
	if err := res.Error; err != nil {
		return 0, err
	}

	if res.RowsAffected == 0 {
		return 0, fmt.Errorf("投票信息不存在[%d]", opts.RecordId)
	}

	if vote.MsgType != entity.MsgTypeVote {
		return 0, fmt.Errorf("当前记录属于投票信息[%d]", vote.MsgType)
	}

	// 判断是否有投票权限

	var count int64
	s.db.Table("talk_records_vote_answer").Where("vote_id = ? and user_id = ？", vote.VoteId, opts.UserId).Count(&count)
	if count > 0 { // 判断是否已投票
		return 0, fmt.Errorf("不能重复投票[%d]", vote.VoteId)
	}

	options := strings.Split(opts.Options, ",")
	sort.Strings(options)

	var answerOptions map[string]interface{}
	if err = jsonutil.Decode(vote.AnswerOption, &answerOptions); err != nil {
		return 0, err
	}

	for _, option := range options {
		if _, ok := answerOptions[option]; !ok {
			return 0, fmt.Errorf("的投票选项不存在[%s]", option)
		}
	}

	// 判断是否单选
	if vote.AnswerMode == 0 {
		options = options[:1]
	}

	answers := make([]*model.TalkRecordsVoteAnswer, 0, len(options))

	for _, option := range options {
		answers = append(answers, &model.TalkRecordsVoteAnswer{
			VoteId: vote.VoteId,
			UserId: opts.UserId,
			Option: option,
		})
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err = tx.Table("talk_records_vote").Where("id = ?", vote.VoteId).Updates(map[string]interface{}{
			"answered_num": gorm.Expr("answered_num + 1"),
			"status":       gorm.Expr("if(answered_num >= answer_num, 1, 0)"),
		}).Error; err != nil {
			return err
		}

		if err = tx.Create(answers).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	_, _ = s.talkRecordsVoteDao.SetVoteAnswerUser(ctx, vote.VoteId)
	_, _ = s.talkRecordsVoteDao.SetVoteStatistics(ctx, vote.VoteId)

	return vote.VoteId, nil
}

// SendLoginMessage 添加登录消息
func (s *TalkMessageService) SendLoginMessage(ctx context.Context, opts *LoginMessageOpts) error {
	var (
		err    error
		record = &model.TalkRecords{
			TalkType:   entity.ChatPrivateMode,
			MsgType:    entity.MsgTypeLogin,
			UserId:     1,
			ReceiverId: opts.UserId,
		}
	)

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err = s.db.Create(record).Error; err != nil {
			return err
		}

		if err = s.db.Create(&model.TalkRecordsLogin{
			RecordId: record.Id,
			UserId:   opts.UserId,
			Ip:       opts.Ip,
			Platform: opts.Platform,
			Agent:    opts.Agent,
			Address:  opts.Address,
			Reason:   "常用设备登录",
		}).Error; err != nil {
			return err
		}

		return nil
	})

	if err == nil {
		if e := s.afterHandle(ctx, record, map[string]string{"text": "[系统通知] 账号登录提醒！"}); e != nil {
			return e
		}

	}

	return err
}

func (s *TalkMessageService) SendDefaultMessage(ctx context.Context, receiverId int) error {
	var (
		err    error
		record = &model.TalkRecords{
			TalkType:   entity.ChatPrivateMode,
			MsgType:    entity.MsgTypeText,
			UserId:     7715,
			ReceiverId: receiverId,
			Content:    "欢迎加入11直播,如在使用过程中发现任何问题,全程为您提供服务",
		}
	)
	if err = s.db.Create(record).Error; err == nil {
		return err
	}
	return nil
}

func (s *TalkMessageService) checkUserAuth(userId int) error {
	//1.检测发送消息用户账号是否被禁止发言
	user := &model.QueryUserTypeItem{}
	if err := s.db.Table("users").Where(&model.Users{Id: userId}).First(user).Error; err != nil {
		return err
	}
	if user.IsMute == 1 {
		return errors.New("你已被禁言，请文明聊天！")
	}
	if user.Type == -1 {
		return errors.New("游客请先登录后再发言！")
	}
	return nil
}

//公开发送信息接口
func (s *TalkMessageService) HandleSendMessage(ctx context.Context, record *model.TalkRecords, opts map[string]string) error {
	return s.afterHandle(ctx, record, opts)
}

// 发送消息后置处理
func (s *TalkMessageService) afterHandle(ctx context.Context, record *model.TalkRecords, opts map[string]string) error {

	if record.TalkType == entity.ChatPrivateMode {
		s.unreadTalkCache.Increment(ctx, record.UserId, record.ReceiverId)

		if record.MsgType == entity.MsgTypeSystemText {
			s.unreadTalkCache.Increment(ctx, record.ReceiverId, record.UserId)
		}
	}

	_ = s.lastMessage.Set(ctx, record.TalkType, record.UserId, record.ReceiverId, &cache.LastCacheMessage{
		Content:  opts["text"],
		Datetime: timeutil.DateTime(),
	})

	content := jsonutil.Encode(map[string]interface{}{
		"event": entity.EventTalk,
		"data": jsonutil.Encode(map[string]interface{}{
			"sender_id":   record.UserId,
			"receiver_id": record.ReceiverId,
			"talk_type":   record.TalkType,
			"record_id":   record.Id,
		}),
	})
	// 点对点消息采用精确投递
	if record.TalkType == entity.ChatPrivateMode {
		sids := s.sidServer.All(ctx, 1)
		// 小于两台服务器则采用全局广播
		if len(sids) <= 3 {
			s.rds.Publish(ctx, entity.IMGatewayAll, content)
		} else {
			to := []int{record.UserId, record.ReceiverId}
			for _, sid := range s.sidServer.All(ctx, 1) {
				for _, uid := range to {
					if s.client.IsCurrentServerOnline(ctx, sid, entity.ImChannelDefault, strconv.Itoa(uid)) {
						s.rds.Publish(ctx, entity.GetIMGatewayPrivate(sid), content)
					}
				}
			}
		}
	} else {
		s.rds.Publish(ctx, entity.IMGatewayAll, content)
	}
	return nil
}
