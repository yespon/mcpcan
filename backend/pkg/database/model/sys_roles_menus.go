package model

// SysRolesMenus 角色菜单关联数据库模型（包含关联关系）
type SysRolesMenus struct {
	MenuID int64 `gorm:"primaryKey;column:menu_id;comment:菜单ID" json:"menuId"`
	RoleID int64 `gorm:"primaryKey;column:role_id;comment:角色ID" json:"roleId"`
}

// TableName 指定表名
func (SysRolesMenus) TableName() string {
	return "sys_roles_menus"
}
