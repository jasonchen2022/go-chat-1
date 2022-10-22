package web

type GetUserInfoResponse struct {
	Id               int    `json:"id"`
	Mobile           string `json:"mobile"`
	Nickname         string `json:"nickname"`
	Avatar           string `json:"avatar"`
	Gender           int    `json:"gender"`
	Motto            string `json:"motto"`
	Email            string `json:"email"`
	MemberId         int    `son:"member_id"`
	MemberLevel      int    `json:"member_level"`
	MemberType       int    `json:"member_type"`
	MemberLevelTitle string `json:"member_level_title"`
}

// ChangeUserDetailRequest ...
type ChangeUserDetailRequest struct {
	Avatar   string `form:"avatar" json:"avatar" binding:"" label:"avatar"`
	Nickname string `form:"nickname" json:"nickname" binding:"required,max=30" label:"nickname"`
	Gender   int    `form:"gender" json:"gender" binding:"oneof=0 1 2" label:"gender"`
	Motto    string `form:"motto" json:"motto" binding:"max=255" label:"motto"`
}

// ChangeUserPasswordRequest ...
type ChangeUserPasswordRequest struct {
	OldPassword string `form:"old_password" json:"old_password" binding:"required" label:"old_password"`
	NewPassword string `form:"new_password" json:"new_password" binding:"required,min=6,max=16" label:"new_password"`
}

// ChangeUserMobileRequest ...
type ChangeUserMobileRequest struct {
	Mobile   string `form:"mobile" json:"mobile" binding:"required,len=11,phone" label:"mobile"`
	Password string `form:"password" json:"password" binding:"required" label:"password"`
	SmsCode  string `form:"sms_code" json:"sms_code" binding:"required,len=6,numeric" label:"sms_code"`
}

// ChangeUserEmailRequest ...
type ChangeUserEmailRequest struct {
	Email    string `form:"email" json:"email" binding:"required" label:"email"`
	Password string `form:"password" json:"password" binding:"required" label:"password"`
	Code     string `form:"code" json:"code" binding:"required,len=6,numeric" label:"code"`
}

//发现好友实体
type RandUserRequest struct {
	UserName string `form:"user_name" json:"user_name" binding:"" label:"user_name"`
	Index    int    `form:"index" json:"index" binding:"required,min=6,max=20" label:"index"`
}
