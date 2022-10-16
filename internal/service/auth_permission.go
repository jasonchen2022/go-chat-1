package service

import (
	"context"

	"go-chat/internal/entity"
	dao2 "go-chat/internal/repository/dao"
	"go-chat/internal/repository/dao/organize"
	"go-chat/internal/repository/model"
)

type AuthPermissionService struct {
	contactDao     *dao2.ContactDao
	groupMemberDao *dao2.GroupMemberDao
	organizeDao    *organize.OrganizeDao
}

func NewAuthPermissionService(contactDao *dao2.ContactDao, groupMemberDao *dao2.GroupMemberDao, organizeDao *organize.OrganizeDao) *AuthPermissionService {
	return &AuthPermissionService{contactDao: contactDao, groupMemberDao: groupMemberDao, organizeDao: organizeDao}
}

type AuthPermission struct {
	TalkType   int
	UserId     int
	ReceiverId int
}

func (a *AuthPermissionService) IsAuth(ctx context.Context, prem *AuthPermission) bool {
	if prem.TalkType == entity.ChatPrivateMode {
		// 这里需要判断双方是否都是企业成员，如果是则无需添加好友即可聊天
		// if isOk, err := a.organizeDao.IsQiyeMember(prem.UserId, prem.ReceiverId); err != nil {
		// 	logger.Error("[AuthPermission IsAuth] 查询数据异常 err: ", err)
		// 	return false
		// } else if isOk {
		// 	return true
		// }

		//判断当前发起者是否管理员或主播，如果是则无需添加好友即可聊天
		if a.contactDao.IsLeader(ctx, prem.UserId) {
			return true
		}
		//判断当前接收者是不是管理员，否则可以直接发起会话
		if a.contactDao.IsLeader(ctx, prem.ReceiverId) {
			return true
		}

		return a.contactDao.IsFriend(ctx, prem.UserId, prem.ReceiverId, false)
	} else if prem.TalkType == entity.ChatGroupMode {
		// 判断群是否解散
		group := &model.Group{}
		err := a.groupMemberDao.Db().First(group, "id = ?", prem.ReceiverId).Error
		if err != nil {
			return false
		}
		if group.Id == 0 || group.IsDismiss == 1 {
			return false
		}
		//聊天室可以查看
		if group.Type == 3 {
			return true
		}
		return a.groupMemberDao.IsMember(prem.ReceiverId, prem.UserId, true)
	}

	return false
}
