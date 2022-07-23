package model

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
