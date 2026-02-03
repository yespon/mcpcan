package model

// SysUsersRoles user role association table model
type SysUsersRoles struct {
	UserID uint `gorm:"column:user_id;primaryKey;not null;comment:用户ID" json:"userId"`
	RoleID uint `gorm:"column:role_id;primaryKey;not null;comment:角色ID" json:"roleId"`
}

// TableName returns table name
func (SysUsersRoles) TableName() string {
	return "sys_users_roles"
}
