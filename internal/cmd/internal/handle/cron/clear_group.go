package cron

import (
	"context"
	"time"

	"go-chat/internal/repository/model"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ClearGroup struct {
	db *gorm.DB
}

func NewClearGroup(db *gorm.DB) *ClearGroup {
	return &ClearGroup{db: db}
}

// Spec 配置定时任务规则
// 每天凌晨1点执行
func (c *ClearGroup) Spec() string {
	//return "0 1 * * *"
	return "*/5 * * * *"
}

func (c *ClearGroup) Handle(ctx context.Context) error {

	c.clearGroup()
	c.clearOfficialGroup()
	return nil
}

// 删除聊天室内容
func (c *ClearGroup) clearGroup() {
	lastId := 0
	size := 100

	for {
		items := make([]*model.Group, 0)

		err := c.db.Model(&model.Group{}).Where("id > ? and type = 3 and created_at <= ? and group_name not like '?%'", lastId, time.Now().AddDate(0, 0, -1), "主播").Order("id asc").Limit(size).Scan(&items).Error
		if err != nil {
			break
		}

		for _, item := range items {

			c.db.Delete(&model.Group{}, item.Id)
			c.db.Delete(&model.GroupMember{}, "group_id = ?", item.Id)
			c.db.Delete(&model.TalkSession{}, "talk_type = 2 and receiver_id = ?", item.Id)
			c.db.Delete(&model.TalkRecords{}, "talk_type = 2 and receiver_id = ?", item.Id)

			logrus.Info("删除群：", item.Name, "  群ID：", item.Id)
		}

		if len(items) < size {
			logrus.Info("结束删除群")
			break
		}

		lastId = items[size-1].Id
	}
}

// 删除官方聊天室游客
func (c *ClearGroup) clearOfficialGroup() {
	c.db.Exec("DELETE m FROM group_member as m where m.group_id=34 and EXISTS(SELECT 1 FROM users as u where m.user_id=u.id and u.type=-1)")
}
