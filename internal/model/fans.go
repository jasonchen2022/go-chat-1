package model

type Fans struct {
	Id         int    `gorm:"column:id;primary_key;AUTO_INCREMENT" json:"id"`       // 主键id
	AnchorId   int    `gorm:"column:anchor_id;default:0;NOT NULL" json:"anchor_id"` // 主播ID
	AnchorName string `gorm:"column:anchor_name;" json:"anchor_name"`               // 主播账号
	FansId     int    `gorm:"column:fans_id;" json:"fans_id"`                       // 用户ID
	FansName   string `gorm:"column:fans_name;" json:"fans_name"`                   // 粉丝账号
	FansType   int    `gorm:"column:fans_type;" json:"fans_type"`                   // 粉丝类型（0普通粉丝  1铁粉）
	Mark       int    `gorm:"column:mark;" json:"mark"`                             // 有效标识：1正常 0删除

}

func (m *Fans) TableName() string {
	return "ff_fans"
}
