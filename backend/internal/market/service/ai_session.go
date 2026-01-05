package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	pb "github.com/kymo-mcp/mcpcan/api/market/ai_agent"
	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
	i18nresp "github.com/kymo-mcp/mcpcan/pkg/i18n"
	"github.com/sashabaranov/go-openai"
)

// AiSessionService struct for ai session service
type AiSessionService struct {
	ctx context.Context
}

// NewAiSessionService creates a new ai session service
func NewAiSessionService(ctx context.Context) *AiSessionService {
	return &AiSessionService{
		ctx: ctx,
	}
}

// CreateHandler creates ai session
func (s *AiSessionService) CreateHandler(c *gin.Context) {
	var req pb.CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.GinError(c, i18nresp.CodeBadRequest, err.Error())
		return
	}

	if req.Name == "" {
		common.GinError(c, i18nresp.CodeBadRequest, "session name is required")
		return
	}
	if req.ModelAccessID == 0 {
		common.GinError(c, i18nresp.CodeBadRequest, "model access id is required")
		return
	}

	// Validate ModelAccessID exists
	if _, err := mysql.AiModelAccessRepo.FindByID(s.ctx, req.ModelAccessID); err != nil {
		common.GinError(c, i18nresp.CodeBadRequest, "invalid model access id")
		return
	}

	// TODO: Get current user ID from context
	// userID := c.GetInt64("user_id")
	userID := int64(1) // Mock user ID for now

	session := &model.AiSession{
		UserID:        userID,
		Name:          req.Name,
		ModelAccessID: req.ModelAccessID,
		MaxContext:    int(req.MaxContext),
		ToolsConfig:   json.RawMessage(req.ToolsConfig),
	}

	if err := mysql.AiSessionRepo.Create(s.ctx, session); err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to create ai session: %s", err.Error()))
		return
	}

	resp := &pb.CreateSessionResponse{
		Session: s.convertModelToProto(session),
	}
	i18nresp.SuccessResponse(c, resp)
}

// UpdateHandler updates ai session
func (s *AiSessionService) UpdateHandler(c *gin.Context) {
	var req pb.UpdateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.GinError(c, i18nresp.CodeBadRequest, err.Error())
		return
	}

	session, err := mysql.AiSessionRepo.FindByID(s.ctx, req.Id)
	if err != nil {
		common.GinError(c, i18nresp.CodeNotFound, "session not found")
		return
	}

	// Update fields if provided
	if req.Name != "" {
		session.Name = req.Name
	}
	if req.ModelAccessID != 0 {
		// Validate ModelAccessID exists
		if _, err := mysql.AiModelAccessRepo.FindByID(s.ctx, req.ModelAccessID); err != nil {
			common.GinError(c, i18nresp.CodeBadRequest, "invalid model access id")
			return
		}
		session.ModelAccessID = req.ModelAccessID
	}
	if req.ToolsConfig != "" {
		session.ToolsConfig = json.RawMessage(req.ToolsConfig)
	}
	if req.MaxContext != 0 {
		session.MaxContext = int(req.MaxContext)
	}

	if err := mysql.AiSessionRepo.Update(s.ctx, session); err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to update ai session: %s", err.Error()))
		return
	}

	resp := &pb.UpdateSessionResponse{
		Session: s.convertModelToProto(session),
	}
	i18nresp.SuccessResponse(c, resp)
}

// DeleteHandler deletes ai session
func (s *AiSessionService) DeleteHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		common.GinError(c, i18nresp.CodeBadRequest, "invalid session id")
		return
	}

	if err := mysql.AiSessionRepo.Delete(s.ctx, id); err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to delete ai session: %s", err.Error()))
		return
	}

	resp := &pb.DeleteSessionResponse{
		Success: true,
	}
	i18nresp.SuccessResponse(c, resp)
}

// GetHandler gets ai session detail
func (s *AiSessionService) GetHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		common.GinError(c, i18nresp.CodeBadRequest, "invalid session id")
		return
	}

	session, err := mysql.AiSessionRepo.FindByID(s.ctx, id)
	if err != nil {
		common.GinError(c, i18nresp.CodeNotFound, "session not found")
		return
	}

	resp := &pb.GetSessionResponse{
		Session: s.convertModelToProto(session),
	}
	i18nresp.SuccessResponse(c, resp)
}

// ListHandler lists ai sessions
func (s *AiSessionService) ListHandler(c *gin.Context) {
	var req pb.ListSessionsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.GinError(c, i18nresp.CodeBadRequest, err.Error())
		return
	}

	// TODO: Get UserID
	userID := int64(1)

	sessions, err := mysql.AiSessionRepo.FindByUserID(s.ctx, userID)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to list ai sessions: %s", err.Error()))
		return
	}

	// Manual pagination
	total := int64(len(sessions))
	page := int(req.Page)
	if page <= 0 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 10
	}

	start := (page - 1) * pageSize
	end := start + pageSize
	if start > int(total) {
		start = int(total)
	}
	if end > int(total) {
		end = int(total)
	}

	var pbSessions []*pb.AiSession
	if start < int(total) {
		slicedSessions := sessions[start:end]
		for _, session := range slicedSessions {
			pbSessions = append(pbSessions, s.convertModelToProto(session))
		}
	}

	resp := &pb.ListSessionsResponse{
		List:  pbSessions,
		Total: total,
	}
	i18nresp.SuccessResponse(c, resp)
}

// GetSessionMessagesHandler gets session messages
func (s *AiSessionService) GetSessionMessagesHandler(c *gin.Context) {
	var req pb.GetSessionMessagesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.GinError(c, i18nresp.CodeBadRequest, err.Error())
		return
	}

	idStr := c.Param("id")
	if idStr != "" {
		if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
			req.SessionID = id
		}
	}

	if req.SessionID == 0 {
		common.GinError(c, i18nresp.CodeBadRequest, "session id is required")
		return
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}

	messages, err := mysql.AiMessageRepo.FindBySessionID(s.ctx, req.SessionID, int(req.Limit))
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to get session messages: %s", err.Error()))
		return
	}

	var pbMessages []*pb.AiMessage
	for _, msg := range messages {
		pbMessages = append(pbMessages, &pb.AiMessage{
			Id:               msg.ID,
			SessionID:        msg.SessionID,
			Role:             msg.Role,
			Content:          msg.Content,
			ToolCalls:        msg.ToolCalls,
			ToolCallID:       msg.ToolCallID,
			PromptTokens:     int32(msg.PromptTokens),
			CompletionTokens: int32(msg.CompletionTokens),
			TotalTokens:      int32(msg.TotalTokens),
			CreateTime:       msg.CreateTime.Unix(),
		})
	}

	resp := &pb.GetSessionMessagesResponse{
		List: pbMessages,
	}
	i18nresp.SuccessResponse(c, resp)
}

// ChatHandler handles chat interaction
func (s *AiSessionService) ChatHandler(c *gin.Context) {
	idStr := c.Param("id")
	sessionID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		common.GinError(c, i18nresp.CodeBadRequest, "invalid session id")
		return
	}

	var req pb.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.GinError(c, i18nresp.CodeBadRequest, err.Error())
		return
	}
	// Override sessionID if path param is present
	req.SessionID = sessionID

	if req.Content == "" {
		common.GinError(c, i18nresp.CodeBadRequest, "content is required")
		return
	}

	// 1. Load Session
	session, err := mysql.AiSessionRepo.FindByID(s.ctx, req.SessionID)
	if err != nil {
		common.GinError(c, i18nresp.CodeNotFound, "session not found")
		return
	}

	// 2. Load Model Access
	modelAccess, err := mysql.AiModelAccessRepo.FindByID(s.ctx, session.ModelAccessID)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, "model access config not found")
		return
	}

	// 3. Load History
	limit := session.MaxContext
	if limit <= 0 {
		limit = 20
	}
	historyMessages, err := mysql.AiMessageRepo.GetLastN(s.ctx, req.SessionID, limit)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, "failed to load history")
		return
	}

	// 4. Init OpenAI Client
	config := openai.DefaultConfig(modelAccess.ApiKey)
	if modelAccess.BaseUrl != "" {
		config.BaseURL = modelAccess.BaseUrl
	}
	client := openai.NewClientWithConfig(config)

	// 5. Construct Messages
	var messages []openai.ChatCompletionMessage
	// Add System Prompt if needed (optional, maybe from session config?)

	// Add History
	for _, msg := range historyMessages {
		role := openai.ChatMessageRoleUser
		if msg.Role == "assistant" {
			role = openai.ChatMessageRoleAssistant
		} else if msg.Role == "system" {
			role = openai.ChatMessageRoleSystem
		} else if msg.Role == "tool" {
			role = openai.ChatMessageRoleTool
		}

		m := openai.ChatCompletionMessage{
			Role:    role,
			Content: msg.Content,
		}
		// Handle ToolCalls if any
		if msg.ToolCalls != "" && msg.ToolCalls != "[]" && msg.ToolCalls != "null" {
			var toolCalls []openai.ToolCall
			if err := json.Unmarshal([]byte(msg.ToolCalls), &toolCalls); err == nil {
				m.ToolCalls = toolCalls
			}
		}
		if msg.ToolCallID != "" {
			m.ToolCallID = msg.ToolCallID
		}
		messages = append(messages, m)
	}

	// Add Current User Message
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: req.Content,
	})

	// 6. Create Stream
	// Set headers for SSE
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	stream, err := client.CreateChatCompletionStream(
		s.ctx,
		openai.ChatCompletionRequest{
			Model:    modelAccess.ModelName,
			Messages: messages,
			Stream:   true,
			StreamOptions: &openai.StreamOptions{
				IncludeUsage: true,
			},
			// Tools: ... (Phase 4)
		},
	)
	if err != nil {
		s.sendSSE(c, "error", fmt.Sprintf("failed to create stream: %v", err))
		return
	}
	defer stream.Close()

	// Save User Message
	userMsg := &model.AiMessage{
		SessionID:  req.SessionID,
		Role:       "user",
		Content:    req.Content,
		CreateTime: time.Now(),
	}
	if err := mysql.AiMessageRepo.Create(s.ctx, userMsg); err != nil {
		// Log error but continue
		fmt.Printf("failed to save user message: %v\n", err)
	}

	var fullContent string
	var usage *openai.Usage

	// 7. Stream Loop
	for {
		response, err := stream.Recv()
		if err == io.EOF {
			s.sendSSE(c, "done", "")
			break
		}
		if err != nil {
			s.sendSSE(c, "error", err.Error())
			break
		}

		if response.Usage != nil {
			usage = response.Usage
		}

		if len(response.Choices) > 0 {
			content := response.Choices[0].Delta.Content
			if content != "" {
				fullContent += content
				s.sendSSE(c, "text", content)
			}
		}
	}

	// Save Assistant Message
	if fullContent != "" {
		assistantMsg := &model.AiMessage{
			SessionID:  req.SessionID,
			Role:       "assistant",
			Content:    fullContent,
			CreateTime: time.Now(),
		}
		if usage != nil {
			assistantMsg.PromptTokens = usage.PromptTokens
			assistantMsg.CompletionTokens = usage.CompletionTokens
			assistantMsg.TotalTokens = usage.TotalTokens
			
			// Also update user message with prompt tokens?
			// Usually Usage is total for the request.
			// We can attribute PromptTokens to userMsg and CompletionTokens to assistantMsg.
			// But userMsg is already saved. We could update it.
			// For simplicity, store all in assistant message or split.
			// The model has these fields on every message.
			// Let's put everything on assistant message for now as it captures the "turn".
		}
		if err := mysql.AiMessageRepo.Create(s.ctx, assistantMsg); err != nil {
			fmt.Printf("failed to save assistant message: %v\n", err)
		}
	}
}

func (s *AiSessionService) sendSSE(c *gin.Context, msgType, content string) {
	resp := &pb.ChatResponse{
		Type:    msgType,
		Content: content,
	}
	data, _ := json.Marshal(resp)
	c.Writer.Write([]byte("data: " + string(data) + "\n\n"))
	c.Writer.Flush()
}

func (s *AiSessionService) convertModelToProto(m *model.AiSession) *pb.AiSession {
	return &pb.AiSession{
		Id:            m.ID,
		Name:          m.Name,
		ModelAccessID: m.ModelAccessID,
		ToolsConfig:   string(m.ToolsConfig),
		MaxContext:    int32(m.MaxContext),
		CreateTime:    m.CreateTime.Unix(),
		UpdateTime:    m.UpdateTime.Unix(),
	}
}
