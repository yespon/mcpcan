package model

import (
	"fmt"
)

// SysRolesDepts model for role-department association
type SysRolesDepts struct {
	RoleID uint `gorm:"column:role_id;primaryKey;not null;comment:Role ID" json:"roleId"`
	DeptID uint `gorm:"column:dept_id;primaryKey;not null;comment:Department ID" json:"deptId"`
}

// TableName returns the table name
func (SysRolesDepts) TableName() string {
	return "sys_roles_depts"
}

// ValidateForCreate validates before creation
func (rd *SysRolesDepts) ValidateForCreate() error {
	if rd.RoleID == 0 {
		return fmt.Errorf("role ID cannot be empty")
	}
	if rd.DeptID == 0 {
		return fmt.Errorf("department ID cannot be empty")
	}
	return nil
}

// ValidateForUpdate validates before update
func (rd *SysRolesDepts) ValidateForUpdate() error {
	return rd.ValidateForCreate()
}

// Clone returns a copy of the object
func (rd *SysRolesDepts) Clone() *SysRolesDepts {
	if rd == nil {
		return nil
	}
	return &SysRolesDepts{
		RoleID: rd.RoleID,
		DeptID: rd.DeptID,
	}
}

// GetRoleID returns the role ID
func (rd *SysRolesDepts) GetRoleID() uint {
	return rd.RoleID
}

// GetDeptID returns the department ID
func (rd *SysRolesDepts) GetDeptID() uint {
	return rd.DeptID
}

// SetRoleID sets the role ID
func (rd *SysRolesDepts) SetRoleID(roleID uint) {
	rd.RoleID = roleID
}

// SetDeptID sets the department ID
func (rd *SysRolesDepts) SetDeptID(deptID uint) {
	rd.DeptID = deptID
}

// IsValid returns true if the association is valid
func (rd *SysRolesDepts) IsValid() bool {
	return rd.RoleID > 0 && rd.DeptID > 0
}

// String returns a string representation
func (rd *SysRolesDepts) String() string {
	return fmt.Sprintf("SysRolesDepts{RoleID: %d, DeptID: %d}", rd.RoleID, rd.DeptID)
}
