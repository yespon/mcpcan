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
// afterId > 0 时只取 id > afterId 的消息（记忆重置卡位）
// 返回按时间正序排列的消息 (旧 -> 新)
func (r *AiMessageRepository) GetLastN(ctx context.Context, sessionID int64, n int, afterId int64) ([]*model.AiMessage, error) {
	return r.FindBySessionId(ctx, sessionID, n, afterId)
}

// FindBySessionId 获取会话的消息列表
// afterId > 0 时只取 id > afterId 的消息
// 返回按时间正序排列的消息 (旧 -> 新)
func (r *AiMessageRepository) FindBySessionId(ctx context.Context, sessionID int64, limit int, afterId ...int64) ([]*model.AiMessage, error) {
	var messages []*model.AiMessage
	query := r.getDB().WithContext(ctx).Where("session_id = ?", sessionID)

	// 记忆重置卡位：只取 id > afterId 的消息
	if len(afterId) > 0 && afterId[0] > 0 {
		query = query.Where("id > ?", afterId[0])
	}

	// 先按倒序取最近N条
	err := query.
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

// FindBySessionIdPaged 分页获取会话消息
// 返回按 id desc 排序的消息 (新 -> 旧)，前端可按需反转
func (r *AiMessageRepository) FindBySessionIdPaged(ctx context.Context, sessionID int64, page, pageSize int) ([]*model.AiMessage, int64, error) {
	var messages []*model.AiMessage
	var total int64

	db := r.getDB().WithContext(ctx).Where("session_id = ?", sessionID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := db.Order("id desc").Offset(offset).Limit(pageSize).Find(&messages).Error
	if err != nil {
		return nil, 0, err
	}

	return messages, total, nil
}

// GetLatestMessageId 获取会话的最新消息 ID
func (r *AiMessageRepository) GetLatestMessageId(ctx context.Context, sessionID int64) (int64, error) {
	var msg model.AiMessage
	err := r.getDB().WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("id desc").
		First(&msg).Error
	if err != nil {
		return 0, err
	}
	return msg.ID, nil
}

// InitTable 初始化表结构
func (r *AiMessageRepository) InitTable() error {
	if err := r.getDB().AutoMigrate(&model.AiMessage{}); err != nil {
		return fmt.Errorf("failed to migrate ai_message table: %v", err)
	}
	return nil
}
