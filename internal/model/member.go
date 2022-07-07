package model

type Member struct {
	Id       int    `gorm:"column:id;primary_key;AUTO_INCREMENT" json:"id"` // 用户ID
	Type     int    `gorm:"column:type;default:0;NOT NULL" json:"type"`     // 用户类型
	Mobile   string `gorm:"column:mobile;" json:"mobile"`                   // 手机号
	UserName string `gorm:"column:username;" json:"username"`               // 用户账号
	Nickname string `gorm:"column:nickname;" json:"nickname"`               // 用户昵称
	Avatar   string `gorm:"column:avatar;" json:"avatar"`                   // 用户头像地址
	Gender   int    `gorm:"column:gender;default:0;NOT NULL" json:"gender"` // 用户性别  0:未知  1:男   2:女
	Password string `gorm:"column:password;" json:"-"`                      // 用户密码
	Motto    string `gorm:"column:motto;" json:"motto"`                     // 用户座右铭
}

func (m *Member) TableName() string {
	return "ff_member"
}
