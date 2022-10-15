package model

import "time"

type GroupApply struct {
	Id        int       `gorm:"column:id;primary_key;AUTO_INCREMENT" json:"id"`     // 自增ID
	GroupId   int       `gorm:"column:group_id;default:0;NOT NULL" json:"group_id"` // 群组ID
	UserId    int       `gorm:"column:user_id;default:0;NOT NULL" json:"user_id"`   // 用户ID
	Remark    string    `gorm:"column:remark;NOT NULL" json:"remark"`               // 备注信息
	CreatedAt time.Time `gorm:"column:created_at;NOT NULL" json:"created_at"`       // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at;NOT NULL" json:"updated_at"`       // 更新时间
}

type GroupApplyList struct {
	Id        int       `gorm:"column:id;primary_key;AUTO_INCREMENT" json:"id"`     // 自增ID
	GroupId   int       `gorm:"column:group_id;default:0;NOT NULL" json:"group_id"` // 群组ID
	UserId    int       `gorm:"column:user_id;default:0;NOT NULL" json:"user_id"`   // 用户ID
	Remark    string    `gorm:"column:remark;NOT NULL" json:"remark"`               // 备注信息
	CreatedAt time.Time `gorm:"column:created_at;NOT NULL" json:"created_at"`       // 创建时间
	Nickname  string    `gorm:"column:nickname;NOT NULL" json:"nickname"`           // 用户昵称
	Avatar    string    `gorm:"column:avatar;NOT NULL" json:"avatar"`               // 用户头像地址
}
