package mysql

import (
	"context"
	"fmt"
	"time"

	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"gorm.io/gorm"
)

var AiMessageRepo *AiMessageRepository

// AiMessageRepository 封装 ai_message 表的操作
type AiMessageRepository struct{}

// NewAiMessageRepository 创建 AiMessageRepository 实例
func NewAiMessageRepository() *AiMessageRepository {
	AiMessageRepo = &AiMessageRepository{}
	return AiMessageRepo
}

func (r *AiMessageRepository) getDB() *gorm.DB {
	return GetDB().Model(&model.AiMessage{})
}

// Create 创建消息
func (r *AiMessageRepository) Create(ctx context.Context, message *model.AiMessage) error {
	message.CreateTime = time.Now()
	return r.getDB().WithContext(ctx).Create(message).Error
}

// GetLastN 获取最近 N 条消息 (用于构建 Context)
// 返回按时间正序排列的消息 (旧 -> 新)
func (r *AiMessageRepository) GetLastN(ctx context.Context, sessionID int64, n int) ([]*model.AiMessage, error) {
	return r.FindBySessionID(ctx, sessionID, n)
}

// FindBySessionID 获取会话的消息列表
// 返回按时间正序排列的消息 (旧 -> 新)
func (r *AiMessageRepository) FindBySessionID(ctx context.Context, sessionID int64, limit int) ([]*model.AiMessage, error) {
	var messages []*model.AiMessage
	// 先按倒序取最近N条
	err := r.getDB().WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("id desc").
		Limit(limit).
		Find(&messages).Error

	if err != nil {
		return nil, err
	}

	// 反转切片，使其变为正序 (旧 -> 新)
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// InitTable 初始化表结构
func (r *AiMessageRepository) InitTable() error {
	if err := r.getDB().AutoMigrate(&model.AiMessage{}); err != nil {
		return fmt.Errorf("failed to migrate ai_message table: %v", err)
	}
	return nil
}
