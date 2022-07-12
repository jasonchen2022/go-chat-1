package cron

import (
	"context"
	"fmt"
	"go-chat/internal/model"
	"log"
	"strconv"
	"time"

	"gorm.io/gorm"
)

type ClearGroupHandle struct {
	db *gorm.DB
}

func NewClearGroupHandle(db *gorm.DB) *ClearGroupHandle {
	return &ClearGroupHandle{db: db}
}

func (c *ClearGroupHandle) GetServiceName() string {
	return "ClearGroupHandle"
}

// Spec 配置定时任务规则
// 每天凌晨1点执行
func (c *ClearGroupHandle) Spec() string {
	// return "0 1 * * *"
	return "*/10 * * * *" //每十分钟执行
	// return "* * * * *"     //每秒执行
}

func (c *ClearGroupHandle) Handle(ctx context.Context) error {
	c.ClearInvalidGroup()
	return nil
}

//清除一天前的聊天记录
func (c *ClearGroupHandle) ClearInvalidGroup() {
	//查出一天前聊天记录
	ids := make([]string, 0)

	// dateTime := time.Now().AddDate(0, 0, -1) //时间戳：time.Now().AddDate(0, 0, -1).Unix()

	if err := c.db.Model(&model.Group{}).Select("id").Where("type = ?", 3).Where("mark = ?", 1).Scan(&ids).Error; err != nil {
		log.Println("ClearGroupHandle 执行出错")
	}

	log.Println("ClearGroupHandle 开始执行删除条数:" + strconv.Itoa(len(ids)))
	fmt.Println("ClearGroupHandle 开始执行", time.Now().AddDate(0, 0, -1))
	if len(ids) > 0 {

		// c.db.Delete(&model.Group{}, "id in ?", ids)
		c.db.Model(&model.Group{}).Where("id in ?", ids).Update("mark", 0)
	}
}
