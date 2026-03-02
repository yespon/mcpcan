package mysql

import (
	"context"
	"fmt"
	"time"

	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"gorm.io/gorm"
)

var AiSessionRepo *AiSessionRepository

// AiSessionRepository 封装 ai_session 表的操作
type AiSessionRepository struct{}

// NewAiSessionRepository 创建 AiSessionRepository 实例
func NewAiSessionRepository() *AiSessionRepository {
	AiSessionRepo = &AiSessionRepository{}
	return AiSessionRepo
}

func (r *AiSessionRepository) getDB() *gorm.DB {
	return GetDB().Model(&model.AiSession{})
}

// Create 创建会话
func (r *AiSessionRepository) Create(ctx context.Context, session *model.AiSession) error {
	session.CreateTime = time.Now()
	session.UpdateTime = time.Now()
	return r.getDB().WithContext(ctx).Create(session).Error
}

// Update 更新会话
func (r *AiSessionRepository) Update(ctx context.Context, session *model.AiSession) error {
	session.UpdateTime = time.Now()
	return r.getDB().WithContext(ctx).Where("id = ?", session.ID).Updates(session).Error
}

// Delete 删除会话
func (r *AiSessionRepository) Delete(ctx context.Context, id int64) error {
	return r.getDB().WithContext(ctx).Where("id = ?", id).Delete(&model.AiSession{}).Error
}

// FindByID 根据ID查找
func (r *AiSessionRepository) FindByID(ctx context.Context, id int64) (*model.AiSession, error) {
	var session model.AiSession
	err := r.getDB().WithContext(ctx).First(&session, id).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// FindByUserID 查找用户的会话列表
func (r *AiSessionRepository) FindByUserID(ctx context.Context, userID int64) ([]*model.AiSession, error) {
	var sessions []*model.AiSession
	err := r.getDB().WithContext(ctx).Where("user_id = ?", userID).Order("update_time desc").Find(&sessions).Error
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

// ResetMemory 重置会话记忆：更新卡位消息 ID
func (r *AiSessionRepository) ResetMemory(ctx context.Context, sessionID int64, messageID int64) error {
	return GetDB().WithContext(ctx).
		Model(&model.AiSession{}).
		Where("id = ?", sessionID).
		Update("memory_reset_message_id", messageID).Error
}

// InitTable 初始化表结构
func (r *AiSessionRepository) InitTable() error {
	if err := r.getDB().AutoMigrate(&model.AiSession{}); err != nil {
		return fmt.Errorf("failed to migrate ai_session table: %v", err)
	}
	return nil
}
