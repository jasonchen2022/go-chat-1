package model

import "time"

type Users struct {
	Id               int       `gorm:"column:id;primary_key;AUTO_INCREMENT" json:"id"`                       // 用户ID
	Type             int       `gorm:"column:type;default:0;NOT NULL" json:"type"`                           // 用户类型
	MemberId         int       `gorm:"column:member_id;default:0;NOT NULL" json:"member_id"`                 // 用户唯一ID
	MemberLevel      int       `gorm:"column:member_level;default:0;NOT NULL" json:"member_level"`           // 用户等级
	MemberLevelTitle string    `gorm:"column:member_level_title;" json:"member_level_title"`                 // 用户等级名称
	ExperiencePoints int       `gorm:"column:experience_points;default:0;NOT NULL" json:"experience_points"` // 用户等级
	Mobile           string    `gorm:"column:mobile;NOT NULL" json:"mobile"`                                 // 手机号
	Username         string    `gorm:"column:username;" json:"username"`                                     // 用户昵称
	Nickname         string    `gorm:"column:nickname;NOT NULL" json:"nickname"`                             // 用户昵称
	Avatar           string    `gorm:"column:avatar;NOT NULL" json:"avatar"`                                 // 用户头像地址
	Gender           int       `gorm:"column:gender;default:0;NOT NULL" json:"gender"`                       // 用户性别  0:未知  1:男   2:女
	Password         string    `gorm:"column:password;NOT NULL" json:"-"`                                    // 用户密码
	Motto            string    `gorm:"column:motto;NOT NULL" json:"motto"`                                   // 用户座右铭
	Email            string    `gorm:"column:email;NOT NULL" json:"email"`                                   // 用户邮箱
	IsRobot          int       `gorm:"column:is_robot;default:0;NOT NULL" json:"is_robot"`                   // 是否机器人[0:否;1:是;]
	IsMute           int       `gorm:"column:is_mute;default:0;NOT NULL" json:"is_mute"`                     // 是否禁言 [0:否;1:是;]
	CreatedAt        time.Time `gorm:"column:created_at;NOT NULL" json:"created_at"`                         // 注册时间
	UpdatedAt        time.Time `gorm:"column:updated_at;NOT NULL" json:"updated_at"`                         // 更新时间
}

func (m *Users) TableName() string {
	return "users"
}

type UserTemp struct {
	Id        int    `json:"id"`         // 用户ID
	Type      int    `json:"type"`       // 用户类型
	Mobile    string `json:"mobile"`     // 手机号
	Nickname  string `json:"nickname"`   // 用户昵称
	Avatar    string `json:"avatar"`     // 用户头像地址
	Gender    int    `json:"gender"`     // 用户性别  0:未知  1:男   2:女
	Password  string `json:"-"`          // 用户密码
	Motto     string `json:"motto"`      // 用户座右铭
	Email     string `json:"email"`      // 用户邮箱
	IsRobot   int    `json:"is_robot"`   // 是否机器人[0:否;1:是;]
	IsMute    int    `json:"is_mute"`    // 是否禁言 [0:否;1:是;]
	IsGaunZhu int    `json:"is_guanzhu"` // 是否已关注 [0:否;1:是;]
	FansCount int    `json:"fans_count"` // 粉丝数
}

type QueryUserItem struct {
	Id               int    `json:"id"`
	Nickname         string `json:"nickname"`
	MemberLevel      int    `json:"member_level"`
	MemberType       int    `json:"member_type"`
	MemberLevelTitle string `json:"member_level_title"`
}

type QueryUserTypeItem struct {
	Id     int `json:"id"`
	IsMute int `json:"is_mute"`
	Type   int `json:"type"`
}
