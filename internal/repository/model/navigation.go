package model

type Navigation struct {
	Id      int    `gorm:"column:id;primary_key;AUTO_INCREMENT" json:"id"`     // 用户ID
	Title   string `gorm:"column:title;" json:"title"`                         // 标题
	SportId int    `gorm:"column:sport_id;default:1;NOT NULL" json:"sport_id"` // 球类ID  1:足球   2:篮球
	Logo    string `gorm:"column:logo;" json:"logo"`                           // logo
	Url     string `gorm:"column:url;" json:"url"`                             // 跳转链接
	Sort    int    `gorm:"column:sort;" json:"sort"`                           // 排序

}

func (m *Navigation) TableName() string {
	return "ff_site_navigation"
}
