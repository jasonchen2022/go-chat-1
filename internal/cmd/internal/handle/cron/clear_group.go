package cron

import (
	"context"
	"time"

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
	return "0 1 * * *"
	//return "*/1 * * * *"
}

func (c *ClearGroup) Handle(ctx context.Context) error {

	c.clearGroup()
	c.clearOfficialGroup()
	c.clearTalkSession()
	return nil
}

// 删除聊天室内容
func (c *ClearGroup) clearGroup() {
	c.db.Exec("DELETE FROM `group` where type = 3 and id != 34 and created_at <= ? and group_name not like '主播%';DELETE gs FROM `group_member` as gs where  not EXISTS(SELECT 1 from `group` as g where gs.group_id=g.id);DELETE gs FROM `talk_session` as gs where gs.talk_type=2 and   not EXISTS(SELECT 1 from `group` as g where gs.receiver_id=g.id);DELETE gs FROM `talk_records` as gs where gs.talk_type=2 and   not EXISTS(SELECT 1 from `group` as g where gs.receiver_id=g.id);", time.Now().AddDate(0, 0, -1))
	logrus.Info("结束删除聊天室内容")
}

// 删除官方聊天室游客
func (c *ClearGroup) clearOfficialGroup() {
	c.db.Exec("DELETE m FROM group_member as m where m.group_id=34 and EXISTS(SELECT 1 FROM users as u where m.user_id=u.id and u.type=-1)")
	logrus.Info("结束删除官方聊天室游客")
}

//删除聊天session
func (c *ClearGroup) clearTalkSession() {
	c.db.Exec("DELETE g FROM group_member as g JOIN users as u on g.group_id=u.id where g.group_id in(SELECT id from `group` where group_name LIKE '%主播%') and u.type=-1; DELETE s FROM talk_session as s where s.talk_type=2 and NOT EXISTS(SELECT 1 FROM  group_member as m where m.user_id=s.user_id and m.group_id=s.receiver_id);")
}
