package mysql

import (
	"context"

	"github.com/kymo-mcp/mcpcan/pkg/enterprise.ee/database/model"
	cmysql "github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
	"gorm.io/gorm"
)

var SysDataPermissionRepo *sysDataPermissionRepository

type sysDataPermissionRepository struct {
	db *gorm.DB
}

func NewSysDataPermissionRepository() *sysDataPermissionRepository {
	if SysDataPermissionRepo == nil {
		SysDataPermissionRepo = &sysDataPermissionRepository{
			db: cmysql.GetDB(),
		}
	}
	return SysDataPermissionRepo
}

func (r *sysDataPermissionRepository) InitTable() error {
	return r.db.AutoMigrate(&model.SysDataPermission{})
}

func (r *sysDataPermissionRepository) GetByData(ctx context.Context, dataType string, dataId string) ([]*model.SysDataPermission, error) {
	var list []*model.SysDataPermission
	err := r.db.WithContext(ctx).
		Where("data_type = ? AND data_id = ?", dataType, dataId).
		Find(&list).Error
	return list, err
}

func (r *sysDataPermissionRepository) BatchSave(ctx context.Context, dataType string, dataId string, permissions []*model.SysDataPermission) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Delete old permissions
		err := tx.Where("data_type = ? AND data_id = ?", dataType, dataId).Delete(&model.SysDataPermission{}).Error
		if err != nil {
			return err
		}

		// 2. Insert new permissions
		if len(permissions) > 0 {
			err = tx.Create(&permissions).Error
			if err != nil {
				return err
			}
		}

		return nil
	})
}
