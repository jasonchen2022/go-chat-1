package model

type RedPacketsRecord struct {
	Id     int    `gorm:"column:id;primary_key;AUTO_INCREMENT" json:"id"`   // 主键ID
	RpId   string `gorm:"column:rp_id;" json:"rp_id"`                       // 关联红包记录id
	UserId int    `gorm:"column:user_id;default:0;NOT NULL" json:"user_id"` // 用户id
	Title  string `gorm:"column:title;" json:"title"`                       // 总球币
	Point  int    `gorm:"column:point;default:0;" json:"point"`             // 球币
}

func (r *RedPacketsRecord) TableName() string {
	return "point_detail"
}
