package mysql

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"gorm.io/gorm"
)

var McpToIntelligentTaskRepo *McpToIntelligentTaskRepository

func init() {
	RegisterInit(func() {
		repo := NewMcpToIntelligentTaskRepository()
		if err := repo.InitTable(); err != nil {
			panic(fmt.Sprintf("Failed to initialize mcp_to_intelligent_task table: %v", err))
		}
	})
}

// McpToIntelligentTaskRepository 封装 mcp_to_intelligent_task 表的增删改查操作
type McpToIntelligentTaskRepository struct{}

// NewMcpToIntelligentTaskRepository 创建 McpToIntelligentTaskRepository 实例
func NewMcpToIntelligentTaskRepository() *McpToIntelligentTaskRepository {
	McpToIntelligentTaskRepo = &McpToIntelligentTaskRepository{}
	return McpToIntelligentTaskRepo
}

// getDB 获取数据库连接
func (r *McpToIntelligentTaskRepository) getDB() *gorm.DB {
	return GetDB().Model(&model.McpToIntelligentTask{})
}

// Create 创建
func (r *McpToIntelligentTaskRepository) Create(ctx context.Context, task *model.McpToIntelligentTask) error {
	return r.getDB().WithContext(ctx).Create(task).Error
}

// Update 更新
func (r *McpToIntelligentTaskRepository) Update(ctx context.Context, task *model.McpToIntelligentTask) error {
	return r.getDB().WithContext(ctx).Where("id = ?", task.ID).Updates(task).Error
}

// Update log 更新日志
func (r *McpToIntelligentTaskRepository) UpdateLogs(ctx context.Context, id int64, logs []*model.InstallLog, status string) error {
	logJson, _ := json.Marshal(logs)
	return r.getDB().WithContext(ctx).Where("id = ?", id).Updates(map[string]interface{}{
		"install_logs": logJson,
		"status":       status,
	}).Error
}

// Delete 删除
func (r *McpToIntelligentTaskRepository) Delete(ctx context.Context, id int64) error {
	return r.getDB().WithContext(ctx).Where("id = ?", id).Delete(&model.McpToIntelligentTask{}).Error
}

// FindByID 根据ID查找
func (r *McpToIntelligentTaskRepository) FindByID(ctx context.Context, id int64) (*model.McpToIntelligentTask, error) {
	var task model.McpToIntelligentTask
	err := r.getDB().WithContext(ctx).Where("id = ?", id).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// FindWithPagination 分页查询
func (r *McpToIntelligentTaskRepository) FindWithPagination(ctx context.Context, page, pageSize int, keyword string, status string) ([]*model.McpToIntelligentTask, int64, error) {
	var tasks []*model.McpToIntelligentTask
	var total int64

	query := r.getDB().WithContext(ctx).Omit("install_logs")

	if keyword != "" {
		query = query.Where("`desc` LIKE ?", "%"+keyword+"%")
	}
	if status != "" {
		query = query.Where("`status` = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// InitTable 初始化表结构
func (r *McpToIntelligentTaskRepository) InitTable() error {
	mod := &model.McpToIntelligentTask{}
	if err := r.getDB().AutoMigrate(mod); err != nil {
		return fmt.Errorf("failed to migrate table %s: %v", mod.TableName(), err)
	}
	return nil
}
