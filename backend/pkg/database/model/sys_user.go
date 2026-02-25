package model

import (
	"time"
)

// UserSource 用户来源枚举
type UserSource string

const (
	UserSourcePlatform UserSource = "PLATFORM" // 自建
	UserSourceFeishu   UserSource = "FEISHU"   // 飞书
)

// SysUser 系统用户模型
type SysUser struct {
	UserID       uint       `gorm:"column:user_id;primaryKey;autoIncrement;comment:ID" json:"userId"`
	DeptID       *uint      `gorm:"column:dept_id;comment:部门名称" json:"deptId"`
	Username     *string    `gorm:"column:username;size:180;comment:用户名" json:"username"`
	NickName     *string    `gorm:"column:nick_name;size:255;comment:昵称" json:"nickName"`
	Gender       *string    `gorm:"column:gender;size:2;comment:性别" json:"gender"`
	Phone        *string    `gorm:"column:phone;size:255;comment:手机号码" json:"phone"`
	Email        *string    `gorm:"column:email;size:180;comment:邮箱" json:"email"`
	AvatarName   *string    `gorm:"column:avatar_name;size:255;comment:头像地址" json:"avatarName"`
	AvatarPath   *string    `gorm:"column:avatar_path;size:255;comment:头像真实路径" json:"avatarPath"`
	Password     *string    `gorm:"column:password;size:255;comment:密码" json:"password"`
	Salt         *string    `gorm:"column:salt;size:255;comment:密码盐" json:"salt"`
	IsAdmin      bool       `gorm:"column:is_admin;default:false;comment:是否为admin账号" json:"isAdmin"`
	Enabled      *bool      `gorm:"column:enabled;comment:状态：1启用、0禁用" json:"enabled"`
	CreateBy     *string    `gorm:"column:create_by;size:255;comment:创建者" json:"createBy"`
	UpdateBy     *string    `gorm:"column:update_by;size:255;comment:更新者" json:"updateBy"`
	PwdResetTime *time.Time `gorm:"column:pwd_reset_time;comment:修改密码的时间" json:"pwdResetTime"`
	CreateTime   *time.Time `gorm:"column:create_time;comment:创建日期" json:"createTime"`
	UpdateTime   *time.Time `gorm:"column:update_time;comment:更新时间" json:"updateTime"`
	Source       *string    `gorm:"column:source;size:32;comment:来源 PLATFORM：自建，FEISHU:飞书" json:"source"`
}

func (u *SysUser) GetAvatarPath() string {
	if u.AvatarPath == nil {
		return ""
	}
	return *u.AvatarPath
}

// IsEnabled 是否启用
func (u *SysUser) IsEnabled() bool {
	if u.Enabled == nil {
		return false
	}
	return *u.Enabled
}

func (u *SysUser) GetDeptID() uint {
	if u.DeptID == nil {
		return 0
	}
	return *u.DeptID
}

func (u *SysUser) GetUsername() string {
	if u.Username == nil {
		return ""
	}
	return *u.Username
}

func (u *SysUser) GetNickName() string {
	if u.NickName == nil {
		return ""
	}
	return *u.NickName
}

func (u *SysUser) GetEmail() string {
	if u.Email == nil {
		return ""
	}
	return *u.Email
}

func (u *SysUser) GetPhone() string {
	if u.Phone == nil {
		return ""
	}
	return *u.Phone
}

// TableName 返回表名
func (SysUser) TableName() string {
	return "sys_user"
}
