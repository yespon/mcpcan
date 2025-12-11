package mysql

import (
	"context"
	"fmt"
	"time"

	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"gorm.io/gorm"
)

var McpToIntelligentTaskLogRepo *McpToIntelligentTaskLogRepository

func init() {
	RegisterInit(func() {
		repo := NewMcpToIntelligentTaskLogRepository()
		if err := repo.InitTable(); err != nil {
			panic(fmt.Sprintf("Failed to initialize mcp_to_intelligent_task table: %v", err))
		}
	})
}

// McpToIntelligentTaskLogRepository 封装 mcp_to_intelligent_task_log 表的增删改查操作
type McpToIntelligentTaskLogRepository struct{}

// NewMcpToIntelligentTaskRepository 创建 McpToIntelligentTaskRepository 实例
func NewMcpToIntelligentTaskLogRepository() *McpToIntelligentTaskLogRepository {
	McpToIntelligentTaskLogRepo = &McpToIntelligentTaskLogRepository{}
	return McpToIntelligentTaskLogRepo
}

// getDB 获取数据库连接
func (r *McpToIntelligentTaskLogRepository) getDB() *gorm.DB {
	return GetDB().Model(&model.McpToIntelligentTaskLog{})
}

// Create 创建
func (r *McpToIntelligentTaskLogRepository) Create(ctx context.Context, task *model.McpToIntelligentTaskLog) error {
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()
	return r.getDB().WithContext(ctx).Create(task).Error
}

// BatchCreate 批量创建
func (r *McpToIntelligentTaskLogRepository) BatchCreate(ctx context.Context, logs []*model.McpToIntelligentTaskLog) error {
	for _, log := range logs {
		log.CreatedAt = time.Now()
		log.UpdatedAt = time.Now()
	}
	return r.getDB().WithContext(ctx).Create(logs).Error
}

// FindListByTaskID 根据任务ID查询日志列表
func (r *McpToIntelligentTaskLogRepository) FindListByTaskID(ctx context.Context, taskID int64) ([]*model.McpToIntelligentTaskLog, error) {
	var logs []*model.McpToIntelligentTaskLog
	if err := r.getDB().WithContext(ctx).Where("task_id = ?", taskID).Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

// DeleteByTaskID 根据任务ID删除日志
func (r *McpToIntelligentTaskLogRepository) DeleteByTaskID(ctx context.Context, taskID int64) error {
	return r.getDB().WithContext(ctx).Where("task_id = ?", taskID).Delete(&model.McpToIntelligentTaskLog{}).Error
}

// InitTable 初始化表结构
func (r *McpToIntelligentTaskLogRepository) InitTable() error {
	mod := &model.McpToIntelligentTaskLog{}
	if err := r.getDB().AutoMigrate(mod); err != nil {
		return fmt.Errorf("failed to migrate table %s: %v", mod.TableName(), err)
	}
	return nil
}
