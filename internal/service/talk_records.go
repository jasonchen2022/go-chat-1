package service

import (
	"context"
	"fmt"
	"sort"
	"time"

	"go-chat/internal/entity"
	"go-chat/internal/pkg/jsonutil"
	"go-chat/internal/pkg/logger"
	"go-chat/internal/pkg/sliceutil"
	"go-chat/internal/pkg/timeutil"
	"go-chat/internal/repository/cache"
	"go-chat/internal/repository/dao"
	"go-chat/internal/repository/model"

	"github.com/wxnacy/wgo/arrays"
)

type TalkRecordsItem struct {
	Id                int         `json:"id"`
	TalkType          int         `json:"talk_type"`
	MsgType           int         `json:"msg_type"`
	UserId            int         `json:"user_id"`
	ReceiverId        int         `json:"receiver_id"`
	Nickname          string      `json:"nickname"`
	Remarkname        string      `json:"remark_name"`
	Avatar            string      `json:"avatar"`
	IsRevoke          int         `json:"is_revoke"`
	IsMark            int         `json:"is_mark"`
	IsRead            int         `json:"is_read"`
	IsLeader          int         `json:"is_leader"`
	FanLevel          int         `json:"fan_level"`
	FanLabel          string      `json:"fan_label"`
	MemberId          int         `json:"member_id"`
	MemberLevel       int         `json:"member_level"`
	MemberLevelTitle  string      `json:"member_level_title"`
	MemberType        int         `json:"member_type"`
	IsMute            int         `json:"is_mute"`
	Content           string      `json:"content,omitempty"`
	File              interface{} `json:"file,omitempty"`
	CodeBlock         interface{} `json:"code_block,omitempty"`
	Forward           interface{} `json:"forward,omitempty"`
	Invite            interface{} `json:"invite,omitempty"`
	Vote              interface{} `json:"vote,omitempty"`
	Login             interface{} `json:"login,omitempty"`
	Location          interface{} `json:"location,omitempty"`
	CreatedAt         string      `json:"created_at"`
	GroupName         string      `json:"group_name"`
	GroupAvatar       string      `json:"group_avatar"`
	GroupType         int         `json:"group_type"`
	RedPacketsStadus  int         `json:"red_packets_stadus"`
	RedPacketsRemarks string      `json:"red_packets_remarks"`
	ReceiverNickname  string      `json:"receiver_nickname"`
	IsRedPackets      int         `json:"is_red_packets"`
}

type TalkRecordsService struct {
	*BaseService
	talkVoteCache         *cache.TalkVote
	talkRecordsVoteDao    *dao.TalkRecordsVoteDao
	groupMemberDao        *dao.GroupMemberDao
	dao                   *dao.TalkRecordsDao
	sensitiveMatchService *SensitiveMatchService
	contactService        *ContactService
}

func NewTalkRecordsService(baseService *BaseService, talkVoteCache *cache.TalkVote, talkRecordsVoteDao *dao.TalkRecordsVoteDao, groupMemberDao *dao.GroupMemberDao, dao *dao.TalkRecordsDao, sensitiveMatchService *SensitiveMatchService, contactService *ContactService) *TalkRecordsService {
	return &TalkRecordsService{BaseService: baseService, talkVoteCache: talkVoteCache, talkRecordsVoteDao: talkRecordsVoteDao, groupMemberDao: groupMemberDao, dao: dao, sensitiveMatchService: sensitiveMatchService, contactService: contactService}
}

func (s *TalkRecordsService) Dao() *dao.TalkRecordsDao {
	return s.dao
}

type QueryTalkRecordsOpt struct {
	TalkType   int    // 对话类型
	UserId     int    // 获取消息的用户
	ReceiverId int    // 接收者ID
	MsgType    []int  // 消息类型
	RecordId   int    // 上次查询的最小消息ID
	Limit      int    // 数据行数
	Keyword    string //搜搜关键字
}

// GetTalkRecords 获取对话消息
func (s *TalkRecordsService) GetTalkRecords(ctx context.Context, opts *QueryTalkRecordsOpt) ([]*TalkRecordsItem, error) {
	var (
		err    error
		items  = make([]*model.QueryTalkRecordsItem, 0)
		fields = []string{
			"talk_records.id",
			"talk_records.talk_type",
			"talk_records.msg_type",
			"talk_records.user_id",
			"talk_records.receiver_id",
			"talk_records.is_revoke",
			"talk_records.is_read",
			"talk_records.content",
			"talk_records.created_at",
			"users.nickname",
			"users.avatar as avatar",
			"1 as fan_level",
			"null as fan_label",
			"users.member_id",
			"users.member_level",
			"users.member_level_title",
			"users.type as member_type",
			"users.is_mute",
			"0 as is_leader",
		}
	)

	query := s.db.Table("talk_records")
	query.Joins("left join users on talk_records.user_id = users.id")

	if opts.RecordId > 0 {
		query.Where("talk_records.id < ?", opts.RecordId)
	}

	if opts.Keyword != "" {
		query.Where("talk_records.content like ?", fmt.Sprintf("%%%s%%", opts.Keyword))
	}

	if opts.TalkType == entity.ChatPrivateMode {
		subQuery := s.db.Where("talk_records.user_id = ? and talk_records.receiver_id = ?", opts.UserId, opts.ReceiverId)
		subQuery.Or("talk_records.user_id = ? and talk_records.receiver_id = ?", opts.ReceiverId, opts.UserId)

		query.Where(subQuery)
	} else {
		query.Where("talk_records.receiver_id = ?", opts.ReceiverId)
	}

	if opts.MsgType != nil && len(opts.MsgType) > 0 {
		query.Where("talk_records.msg_type in ?", opts.MsgType)
	}

	query.Where("talk_records.talk_type = ?", opts.TalkType)
	query.Where("NOT EXISTS (SELECT 1 FROM `talk_records_delete` WHERE talk_records_delete.record_id = talk_records.id AND talk_records_delete.user_id = ? LIMIT 1)", opts.UserId)
	query.Select(fields).Order("talk_records.id desc").Limit(opts.Limit)

	if err = query.Scan(&items).Error; err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return make([]*TalkRecordsItem, 0), err
	} else {
		//如果群聊则查询当前聊天记录内用户信息
		if opts.TalkType == entity.ChatGroupMode {
			var userIds []int64
			for _, item := range items {
				if arrays.ContainsInt(userIds, int64(item.UserId)) < 0 {
					userIds = append(userIds, int64(item.UserId))
				}

			}
			memberQuery := s.db.Table("group_member")
			memberQuery.Joins("left join users on group_member.user_id = users.id")
			memberQuery.Joins("left join `group` on `group`.id = group_member.group_id")
			memberQuery.Where("group_member.group_id = ?", opts.ReceiverId)
			memberQuery.Where("group_member.user_id in ?", userIds)
			var memberFields = []string{
				"group_member.user_id",
				"users.type as member_type",
				"users.is_mute",
				"users.member_id",
				"users.member_level",
				"users.member_level_title",
				"group_member.leader as is_leader",
				"`group`.group_name",
				"`group`.avatar as group_avatar",
				"`group`.type as group_type",
			}
			var memberItems = make([]*model.QueryGroupMemberItem, 0)

			//更新用户信息
			if err = memberQuery.Select(memberFields).Scan(&memberItems).Error; err == nil {
				if len(memberItems) > 0 {
					for _, item := range memberItems {
						for _, record := range items {
							if item.UserId == record.UserId {
								record.IsLeader = item.IsLeader
								record.MemberType = item.MemberType
								record.MemberLevel = item.MemberLevel
								record.MemberId = item.MemberId
								record.MemberLevelTitle = item.MemberLevelTitle
								record.IsMute = item.IsMute
								record.GroupName = item.GroupName
								record.GroupAvatar = item.GroupAvatar
								record.GroupType = item.GroupType
							}
						}
					}
				}
			}

		} else {
			// 获取好友备注
			remarks, _ := s.contactService.Dao().Remarks(ctx, opts.UserId, []int{opts.ReceiverId})
			for _, item := range items {
				if len(remarks) > 0 {
					item.Remarkname = remarks[item.UserId]
				}
			}
		}
	}

	return s.HandleTalkRecords(ctx, opts.UserId, items)
}

// SearchTalkRecords 对话搜索消息
func (s *TalkRecordsService) SearchTalkRecords() {

}

func (s *TalkRecordsService) GetTalkRecord(ctx context.Context, recordId int64) (*TalkRecordsItem, error) {
	var (
		err    error
		record *model.QueryTalkRecordsItem
		fields = []string{
			"talk_records.id",
			"talk_records.talk_type",
			"talk_records.msg_type",
			"talk_records.user_id",
			"talk_records.receiver_id",
			"talk_records.is_revoke",
			"talk_records.content",
			"talk_records.created_at",
			"users.nickname",
			"users.avatar as avatar",
			"1 as fan_level",
			"null as fan_label",
			"users.member_id",
			"users.member_level",
			"users.member_level_title",
			"users.type as member_type",
			"users.is_mute",
			"0 as is_leader",
		}
	)

	query := s.db.Table("talk_records")
	query.Joins("left join users on talk_records.user_id = users.id")
	query.Where("talk_records.id = ?", recordId)

	if err = query.Select(fields).Take(&record).Error; err != nil {
		return nil, err
	}

	//如果群聊则查询当前聊天记录内用户信息
	if record.TalkType == entity.ChatGroupMode {
		memberQuery := s.db.Table("group_member")
		memberQuery.Joins("left join users on group_member.user_id = users.id")
		memberQuery.Joins("left join `group` on `group`.id = group_member.group_id")
		memberQuery.Where("group_member.group_id = ?  and is_quit =0 ", record.ReceiverId)
		memberQuery.Where("group_member.user_id = ? ", record.UserId)
		var memberFields = []string{
			"group_member.user_id",
			"users.type as member_type",
			"users.is_mute",
			"users.member_id",
			"users.member_level",
			"users.member_level_title",
			"group_member.leader as is_leader",
			"`group`.group_name",
			"`group`.avatar as group_avatar",
			"`group`.type as group_type",
		}
		var memberItems = make([]*model.QueryGroupMemberItem, 0)
		if err = memberQuery.Select(memberFields).Scan(&memberItems).Error; err == nil {
			if len(memberItems) > 0 {
				for _, item := range memberItems {
					if item.UserId == record.UserId {
						record.IsLeader = item.IsLeader
						record.MemberType = item.MemberType
						record.MemberLevel = item.MemberLevel
						record.MemberId = item.MemberId
						record.MemberLevelTitle = item.MemberLevelTitle
						record.IsMute = item.IsMute
						record.GroupName = item.GroupName
						record.GroupAvatar = item.GroupAvatar
						record.GroupType = item.GroupType
					}
				}
			}
		}

	} else {
		// 获取好友备注
		remarks, _ := s.contactService.Dao().Remarks(ctx, record.UserId, []int{record.ReceiverId})
		if len(remarks) > 0 {
			record.Remarkname = remarks[record.UserId]
		}

	}

	list, err := s.HandleTalkRecords(ctx, record.UserId, []*model.QueryTalkRecordsItem{record})
	if err != nil {
		return nil, err
	}

	return list[0], nil
}

// GetForwardRecords 获取转发消息记录
func (s *TalkRecordsService) GetForwardRecords(ctx context.Context, uid int, recordId int64) ([]*TalkRecordsItem, error) {
	record := &model.TalkRecords{}
	if err := s.db.First(&record, recordId).Error; err != nil {
		return nil, err
	}

	if record.TalkType == entity.ChatPrivateMode {
		if record.UserId != uid && record.ReceiverId != uid {
			return nil, entity.ErrPermissionDenied
		}
	} else if record.TalkType == entity.ChatGroupMode {
		if !s.groupMemberDao.IsMember(record.ReceiverId, uid, true) {
			return nil, entity.ErrPermissionDenied
		}
	} else {
		return nil, entity.ErrPermissionDenied
	}

	forward := &model.TalkRecordsForward{}
	if err := s.db.Where("record_id = ?", recordId).First(forward).Error; err != nil {
		return nil, err
	}

	var (
		items  = make([]*model.QueryTalkRecordsItem, 0)
		fields = []string{
			"talk_records.id",
			"talk_records.talk_type",
			"talk_records.msg_type",
			"talk_records.user_id",
			"talk_records.receiver_id",
			"talk_records.is_revoke",
			"talk_records.content",
			"talk_records.created_at",
			"users.nickname",
			"users.avatar as avatar",
		}
	)

	query := s.db.Table("talk_records")
	query.Select(fields)
	query.Joins("left join users on talk_records.user_id = users.id")
	query.Where("talk_records.id in ?", sliceutil.ParseIds(forward.RecordsId))

	if err := query.Scan(&items).Error; err != nil {
		return nil, err
	}

	return s.HandleTalkRecords(ctx, uid, items)
}

func (s *TalkRecordsService) HandleTalkRecords(ctx context.Context, uid int, items []*model.QueryTalkRecordsItem) ([]*TalkRecordsItem, error) {
	var (
		files              []int
		codes              []int
		forwards           []int
		invites            []int
		votes              []int
		logins             []int
		locations          []int
		notices            []string
		fileItems          []*model.TalkRecordsFile
		codeItems          []*model.TalkRecordsCode
		forwardItems       []*model.TalkRecordsForward
		inviteItems        []*model.TalkRecordsInvite
		voteItems          []*model.TalkRecordsVote
		loginItems         []*model.TalkRecordsLogin
		locationItems      []*model.TalkRecordsLocation
		recordIds          []string
		red_packets        []*model.RedPackets
		red_packets_record []*model.RedPacketsRecord
		users              []*model.Users
		usersId            []int
	)

	for _, item := range items {
		switch item.MsgType {
		case entity.MsgTypeFile:
			files = append(files, item.Id)
		case entity.MsgTypeForward:
			forwards = append(forwards, item.Id)
		case entity.MsgTypeCode:
			codes = append(codes, item.Id)
		case entity.MsgTypeVote:
			votes = append(votes, item.Id)
		case entity.MsgTypeGroupNotice:
		case entity.MsgTypeFriendApply:
		case entity.MsgTypeLogin:
			logins = append(logins, item.Id)
		case entity.MsgTypeGroupInvite:
			invites = append(invites, item.Id)
		case entity.MsgTypeLocation:
			locations = append(locations, item.Id)
		case entity.MsgTypeRedPackets:
			recordIds = append(recordIds, item.Content)
		case entity.MsgTypeSysRedPackets:
			usersId = append(usersId, item.ReceiverId)
		}

		//已撤回的消息不能显示
		if item.IsRevoke == 1 {
			item.Content = ""
		}
	}

	hashFiles := make(map[int]*model.TalkRecordsFile)
	if len(files) > 0 {
		s.db.Model(&model.TalkRecordsFile{}).Where("record_id in ?", files).Scan(&fileItems)
		for i := range fileItems {
			hashFiles[fileItems[i].RecordId] = fileItems[i]
		}
	}

	hashForwards := make(map[int]*model.TalkRecordsForward)
	if len(forwards) > 0 {
		s.db.Model(&model.TalkRecordsForward{}).Where("record_id in ?", forwards).Scan(&forwardItems)
		for i := range forwardItems {
			hashForwards[forwardItems[i].RecordId] = forwardItems[i]
		}
	}

	hashCodes := make(map[int]*model.TalkRecordsCode)
	if len(codes) > 0 {
		s.db.Model(&model.TalkRecordsCode{}).Where("record_id in ?", codes).Select("record_id", "lang", "code").Scan(&codeItems)
		for i := range codeItems {
			hashCodes[codeItems[i].RecordId] = codeItems[i]
		}
	}

	hashVotes := make(map[int]*model.TalkRecordsVote)
	if len(votes) > 0 {
		s.db.Model(&model.TalkRecordsVote{}).Where("record_id in ?", votes).Scan(&voteItems)
		for i := range voteItems {
			hashVotes[voteItems[i].RecordId] = voteItems[i]
		}
	}

	hashLogins := make(map[int]*model.TalkRecordsLogin)
	if len(logins) > 0 {
		s.db.Model(&model.TalkRecordsLogin{}).Where("record_id in ?", logins).Scan(&loginItems)
		for i := range loginItems {
			hashLogins[loginItems[i].RecordId] = loginItems[i]
		}
	}

	hashInvites := make(map[int]*model.TalkRecordsInvite)
	var noticeResult []*model.QueryUserItem
	var operateUserResult []*model.QueryUserItem
	if len(invites) > 0 {
		s.db.Model(&model.TalkRecordsInvite{}).Where("record_id in ?", invites).Scan(&inviteItems)
		for i := range inviteItems {
			hashInvites[inviteItems[i].RecordId] = inviteItems[i]
			if inviteItems[i].Type == 1 || inviteItems[i].Type == 3 {
				notices = append(notices, inviteItems[i].UserIds)
			}

		}
		//邀请通知用户
		if len(notices) > 0 {
			var inviteUserIds []int
			for _, item := range notices {
				tempUserIds := sliceutil.ParseIds(item)
				for i := 0; i < len(tempUserIds); i++ {
					inviteUserIds = append(inviteUserIds, tempUserIds[i])
				}
			}
			s.db.Table("users").Select("id", "nickname", "type as member_type", "member_level", "member_level_title").Where("id in ?", inviteUserIds).Scan(&noticeResult)
		}
		//邀请操作人管理员ID
		if len(hashInvites) > 0 {
			var operateUserIds []int
			for _, item := range hashInvites {
				var isExit bool = false
				for i := 0; i < len(operateUserIds); i++ {
					if item.OperateUserId == operateUserIds[i] {
						isExit = true
					}
				}
				if !isExit {
					operateUserIds = append(operateUserIds, item.OperateUserId)
				}

			}
			s.db.Table("users").Select("id", "nickname").Where("id in ?", operateUserIds).Scan(&operateUserResult)
		}

	}

	hashLocations := make(map[int]*model.TalkRecordsLocation)
	if len(locations) > 0 {
		s.db.Model(&model.TalkRecordsLocation{}).Where("record_id in ?", locations).Scan(&locationItems)
		for i := range locationItems {
			hashLocations[locationItems[i].RecordId] = locationItems[i]
		}
	}

	if len(recordIds) > 0 {
		s.db.Model(&model.RedPackets{}).Where("record_id in ?", recordIds).Scan(&red_packets)
	}

	if len(recordIds) > 0 {
		s.db.Model(&model.RedPacketsRecord{}).Where("point > ?", 0).Where("rp_id in ? and user_id = ?", recordIds, uid).Scan(&red_packets_record)
	}
	if len(usersId) > 0 {
		s.db.Model(&model.Users{}).Where("id in ? ", usersId).Scan(&users)
	}
	senService := s.sensitiveMatchService.GetService()

	newItems := make([]*TalkRecordsItem, 0, len(items))
	for _, item := range items {
		data := &TalkRecordsItem{
			Id:               item.Id,
			TalkType:         item.TalkType,
			MsgType:          item.MsgType,
			UserId:           item.UserId,
			FanLevel:         item.FanLevel,
			FanLabel:         item.FanLabel,
			MemberType:       item.MemberType,
			MemberLevel:      item.MemberLevel,
			MemberLevelTitle: item.MemberLevelTitle,
			ReceiverId:       item.ReceiverId,
			Nickname:         item.Nickname,
			Remarkname:       item.Remarkname,
			Avatar:           item.Avatar,
			IsRevoke:         item.IsRevoke,
			IsMark:           item.IsMark,
			IsRead:           item.IsRead,
			IsLeader:         item.IsLeader,
			IsMute:           item.IsMute,
			Content:          item.Content,
			CreatedAt:        timeutil.FormatDatetime(item.CreatedAt),
			GroupName:        item.GroupName,
			GroupAvatar:      item.GroupAvatar,
			GroupType:        item.GroupType,
		}
		if data.MemberType <= 0 {
			_, content := senService.Match(data.Content, '*')
			if content != "" {
				data.Content = content
			}
		}
		switch item.MsgType {
		case entity.MsgTypeFile:
			if value, ok := hashFiles[item.Id]; ok {
				data.File = value
			} else {
				logger.Warnf("文件消息信息不存在[%d]", item.Id)
			}
		case entity.MsgTypeForward:
			if value, ok := hashForwards[item.Id]; ok {
				list := make([]map[string]interface{}, 0)

				_ = jsonutil.Decode(value.Text, &list)

				data.Forward = map[string]interface{}{
					"num":  len(sliceutil.ParseIds(value.RecordsId)),
					"list": list,
				}
			}
		case entity.MsgTypeCode:
			if value, ok := hashCodes[item.Id]; ok {
				data.CodeBlock = value
			}
		case entity.MsgTypeVote:
			if value, ok := hashVotes[item.Id]; ok {
				options := make(map[string]interface{})
				opts := make([]interface{}, 0)

				if err := jsonutil.Decode(value.AnswerOption, &options); err == nil {
					arr := make([]string, 0, len(options))
					for k := range options {
						arr = append(arr, k)
					}

					sort.Strings(arr)

					for _, v := range arr {
						opts = append(opts, map[string]interface{}{
							"key":   v,
							"value": options[v],
						})
					}
				}

				users := make([]int, 0)
				if uids, err := s.talkRecordsVoteDao.GetVoteAnswerUser(ctx, value.Id); err == nil {
					users = uids
				}

				var statistics interface{}

				if res, err := s.talkRecordsVoteDao.GetVoteStatistics(ctx, value.Id); err != nil {
					statistics = map[string]interface{}{
						"count":   0,
						"options": map[string]int{},
					}
				} else {
					statistics = res
				}

				data.Vote = map[string]interface{}{
					"detail": map[string]interface{}{
						"id":            value.Id,
						"record_id":     value.RecordId,
						"title":         value.Title,
						"answer_mode":   value.AnswerMode,
						"status":        value.Status,
						"answer_option": opts,
						"answer_num":    value.AnswerNum,
						"answered_num":  value.AnsweredNum,
					},
					"statistics": statistics,
					"vote_users": users, // 已投票成员
				}
			}
		case entity.MsgTypeGroupNotice:
		case entity.MsgTypeFriendApply:
		case entity.MsgTypeLogin:
			if value, ok := hashLogins[item.Id]; ok {
				data.Login = map[string]interface{}{
					"address":    value.Address,
					"agent":      value.Agent,
					"created_at": value.CreatedAt.Format(timeutil.DatetimeFormat),
					"ip":         value.Ip,
					"platform":   value.Platform,
					"reason":     value.Reason,
				}
			}
		case entity.MsgTypeGroupInvite:
			if value, ok := hashInvites[item.Id]; ok {
				operateUser := map[string]interface{}{
					"id":       value.OperateUserId,
					"nickname": "",
				}
				for _, user := range operateUserResult {
					if value.OperateUserId == user.Id {
						operateUser["nickname"] = user.Nickname
					}
				}

				m := map[string]interface{}{
					"type":         value.Type,
					"operate_user": operateUser,
					"users":        map[string]interface{}{},
				}

				if value.Type == 1 || value.Type == 3 {
					var results []*model.QueryUserItem
					for _, userId := range sliceutil.ParseIds(value.UserIds) {
						for _, user := range noticeResult {
							if userId == user.Id {
								results = append(results, user)
							}
						}
					}
					m["users"] = results
					//如果是入群通知则再查一次用户数据
					if len(results) > 0 {
						data.MemberLevel = results[0].MemberLevel
						data.MemberLevelTitle = results[0].MemberLevelTitle
						data.MemberType = results[0].MemberType
					}

				} else {
					m["users"] = operateUser
				}

				data.Invite = m
			}
		case entity.MsgTypeLocation:
			if value, ok := hashLocations[item.Id]; ok {
				data.Location = value
			}
		case entity.MsgTypeRedPackets:
			data.IsRedPackets = 1
			record_id := item.Content
			//红包状态(0未领取 1已领取 2已过期 3已领完自己未领取)

			//0未领取
			data.RedPacketsStadus = 0
			for _, rp_item := range red_packets_record {
				if rp_item.RpId == record_id && rp_item.UserId == uid {
					//1已领取
					data.RedPacketsStadus = 1
				}
			}

			for _, r_item := range red_packets {
				if record_id == r_item.RecordId {
					//红包备注
					data.RedPacketsRemarks = r_item.Remark
				}
				//如果未领取  再判断是否过期  或  已经被领完
				if data.RedPacketsStadus == 0 {
					if record_id == r_item.RecordId {
						//当前时间
						cur_time := time.Now().Unix()
						//有效期
						val_time := r_item.ValTime.Unix()

						if val_time < cur_time {
							//过期
							data.RedPacketsStadus = 2
						} else if val_time > cur_time && r_item.Count <= 0 {
							// 3已领完自己未领取
							data.RedPacketsStadus = 3
						}
					}
				}
			}
		case entity.MsgTypeSysRedPackets:
			for _, u_item := range users {
				if u_item.Id == item.ReceiverId {
					data.ReceiverNickname = u_item.Nickname
				}
			}

		}

		newItems = append(newItems, data)
	}

	return newItems, nil
}
