package model

import "time"

type RedPackets struct {
	Id         int       `gorm:"column:id;primary_key;AUTO_INCREMENT" json:"id"`   // 主键ID
	RecordId   string    `gorm:"column:record_id;" json:"record_id"`               // 记录id
	UserId     int       `gorm:"column:user_id;default:0;NOT NULL" json:"user_id"` // 用户id
	TotalPoint int       `gorm:"column:total_point;default:0;" json:"total_point"` // 总球币
	Point      int       `gorm:"column:point;default:0;" json:"point"`             // 每个红包分发球币
	Count      int       `gorm:"column:count;default:0;" json:"count"`             // 实时剩余数量
	RealCount  int       `gorm:"column:real_count;default:0;" json:"real_count"`   // 红包个数
	ValTime    time.Time `gorm:"column:val_time;" json:"val_time"`                 // 过期时间
	Remark     string    `gorm:"column:remark;" json:"remark"`                     // 备注
}

func (r *RedPackets) TableName() string {
	return "ff_red_packets"
}
