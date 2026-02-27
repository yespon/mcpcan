package model

import "time"

// AiMessage 消息记录与计费
type AiMessage struct {
	ID               int64     `gorm:"primaryKey;autoIncrement;column:id;comment:消息ID"`
	SessionId        int64     `gorm:"column:session_id;type:bigint(20);not null" json:"sessionId"`
	Role             string    `gorm:"type:varchar(50);column:role;comment:角色: system, user, assistant, tool"`
	Content          string    `gorm:"type:text;column:content;comment:消息内容"`
	ToolCalls        string    `gorm:"type:text;column:tool_calls;comment:工具调用参数(JSON)"`
	ToolCallID       string    `gorm:"type:varchar(255);column:tool_call_id;comment:关联的tool_call_id"`
	PromptTokens     int       `gorm:"default:0;column:prompt_tokens;comment:Prompt Token数"`
	CompletionTokens int       `gorm:"default:0;column:completion_tokens;comment:Completion Token数"`
	TotalTokens      int       `gorm:"default:0;column:total_tokens;comment:总Token数"`
	ReasoningContent string    `gorm:"type:text;column:reasoning_content;comment:思考过程"`
	CreateTime       time.Time `gorm:"autoCreateTime;column:create_time;comment:创建时间"`
}

func (m *AiMessage) TableName() string {
	return "mcpcan_ai_message"
}
