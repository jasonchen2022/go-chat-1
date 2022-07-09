package model

type DictData struct {
	Id     int    `gorm:"column:id;primary_key;AUTO_INCREMENT" json:"id"`   // 字典ID
	Name   string `gorm:"column:name;" json:"name"`                         // 字典名称
	Code   string `gorm:"column:code;" json:"code"`                         // 字典值
	DictId int    `gorm:"column:dict_id;default:0;NOT NULL" json:"dict_id"` // 父字典ID
	Status int    `gorm:"column:status;default:0;NOT NULL" json:"status"`   // 状态
}

func (m *DictData) TableName() string {
	return "ff_dict_data"
}
