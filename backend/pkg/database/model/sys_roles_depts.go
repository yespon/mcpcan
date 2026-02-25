package model

// SysRolesDepts model for role-department association
type SysRolesDepts struct {
	RoleID uint `gorm:"column:role_id;primaryKey;not null;comment:Role ID" json:"roleId"`
	DeptID uint `gorm:"column:dept_id;primaryKey;not null;comment:Department ID" json:"deptId"`
}

// TableName returns the table name
func (SysRolesDepts) TableName() string {
	return "sys_roles_depts"
}
