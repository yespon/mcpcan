package service

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	pb "github.com/kymo-mcp/mcpcan/api/market/ai_agent"
	"github.com/kymo-mcp/mcpcan/internal/market/biz"
	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
	i18nresp "github.com/kymo-mcp/mcpcan/pkg/i18n"
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

	// 从认证上下文中获取当前用户 ID
	userID, err := common.GetUserIDFromContext(c)
	if err != nil {
		common.GinError(c, i18nresp.CodeUnauthorized, "user not authenticated")
		return
	}

	session, err := biz.GAiSessionBiz.Create(c.Request.Context(), &req, userID)
	if err != nil {
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

	session, err := biz.GAiSessionBiz.Update(c.Request.Context(), &req)
	if err != nil {
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

	if err := biz.GAiSessionBiz.Delete(c.Request.Context(), id); err != nil {
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

	session, err := biz.GAiSessionBiz.Get(c.Request.Context(), id)
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

	// 从认证上下文中获取当前用户 ID
	userID, err := common.GetUserIDFromContext(c)
	if err != nil {
		common.GinError(c, i18nresp.CodeUnauthorized, "user not authenticated")
		return
	}

	sessions, total, err := biz.GAiSessionBiz.List(c.Request.Context(), userID, int(req.Page), int(req.PageSize))
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to list ai sessions: %s", err.Error()))
		return
	}

	var pbSessions []*pb.AiSession
	for _, session := range sessions {
		pbSessions = append(pbSessions, s.convertModelToProto(session))
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
			req.SessionId = id
		}
	}

	if req.SessionId == 0 {
		common.GinError(c, i18nresp.CodeBadRequest, "session id is required")
		return
	}

	// Default PageSize if not strict
	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 20
	}
	page := int(req.Page)
	if page <= 0 {
		page = 1
	}

	messages, total, err := biz.GAiSessionBiz.GetMessages(c.Request.Context(), req.SessionId, page, pageSize)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to get session messages: %s", err.Error()))
		return
	}

	var pbMessages []*pb.AiMessage
	for _, msg := range messages {
		pbMessages = append(pbMessages, &pb.AiMessage{
			Id:               msg.ID,
			SessionId:        msg.SessionId,
			Role:             msg.Role,
			Content:          msg.Content,
			ToolCalls:        msg.ToolCalls,
			ToolCallID:       msg.ToolCallID,
			CreateTime:       msg.CreateTime.Unix(),
			PromptTokens:     int32(msg.PromptTokens),
			CompletionTokens: int32(msg.CompletionTokens),
			TotalTokens:      int32(msg.TotalTokens),
		})
	}

	resp := &pb.GetSessionMessagesResponse{
		List:  pbMessages,
		Total: total,
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
	req.SessionId = sessionID

	if req.Content == "" {
		common.GinError(c, i18nresp.CodeBadRequest, "content is required")
		return
	}

	// Set headers for SSE
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	// Call Biz
	stream, err := biz.GAiSessionBiz.Chat(c.Request.Context(), &req)
	if err != nil {
		s.sendSSE(c, "error", fmt.Sprintf("failed to start chat: %v", err))
		return
	}

	// Stream Loop
	for resp := range stream {
		if resp.Error != nil {
			s.sendSSE(c, "error", resp.Error.Error())
			break
		}

		// Handle Usage Event
		if resp.Usage != nil {
			usageBytes, _ := json.Marshal(resp.Usage)
			s.sendSSE(c, "usage", string(usageBytes))
		}

		if resp.Content != "" {
			s.sendSSE(c, "text", resp.Content)
		}

		if len(resp.ToolOutputs) > 0 {
			if b, err := json.Marshal(resp.ToolOutputs); err == nil {
				s.sendSSE(c, "tool_result", string(b))
			}
		}

		if len(resp.ToolCalls) > 0 {
			if b, err := json.Marshal(resp.ToolCalls); err == nil {
				s.sendSSE(c, "tool_calls", string(b))
			}
		}
	}
	s.sendSSE(c, "done", "")
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
		ModelName:     m.ModelName,
		Temperature:   float32(m.Temperature),
		SystemPrompt:  m.SystemPrompt,
		ToolsConfig:   string(m.ToolsConfig),
		MaxContext:    int32(m.MaxContext),
		CreateTime:    m.CreateTime.Unix(),
		UpdateTime:    m.UpdateTime.Unix(),
	}
}

// GetSessionUsageHandler 获取会话的 Token 使用统计
func (s *AiSessionService) GetSessionUsageHandler(c *gin.Context) {
	idStr := c.Param("id")
	sessionID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		common.GinError(c, i18nresp.CodeBadRequest, "invalid session id")
		return
	}

	usage, err := biz.GAiSessionBiz.GetSessionUsage(c.Request.Context(), sessionID)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to get session usage: %s", err.Error()))
		return
	}

	i18nresp.SuccessResponse(c, usage)
}

// UploadFileHandler handles file upload for chat
// 表单字段: file (multipart), sessionId (string)
func (s *AiSessionService) UploadFileHandler(c *gin.Context) {
	// 解析 multipart form（最大 100MB）
	if err := c.Request.ParseMultipartForm(100 << 20); err != nil {
		common.GinError(c, i18nresp.CodeBadRequest, fmt.Sprintf("failed to parse multipart form: %v", err))
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		common.GinError(c, i18nresp.CodeBadRequest, fmt.Sprintf("failed to get file: %v", err))
		return
	}
	defer file.Close()

	// 获取并校验会话
	var sessionId int64
	if sidStr := c.Request.FormValue("sessionId"); sidStr != "" {
		fmt.Sscanf(sidStr, "%d", &sessionId)
	}

	session, err := biz.GAiSessionBiz.Get(c.Request.Context(), sessionId)
	if err != nil {
		common.GinError(c, i18nresp.CodeNotFound, "session not found")
		return
	}

	// 加载模型限制
	modelAccess, err := mysql.AiModelAccessRepo.FindByID(c.Request.Context(), session.ModelAccessID)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, "model access config not found")
		return
	}

	if !modelAccess.VisionEnabled {
		common.GinError(c, i18nresp.CodeBadRequest, "The active model does not support vision/image upload")
		return
	}

	// 文件大小校验
	maxBytes := int64(modelAccess.MaxImageSize * 1024 * 1024)
	if header.Size > maxBytes {
		common.GinError(c, i18nresp.CodeBadRequest, fmt.Sprintf("Image size exceeds the model's limit of %d MB", modelAccess.MaxImageSize))
		return
	}

	// 格式后缀校验
	if modelAccess.AllowedImageTypes != "" && modelAccess.AllowedImageTypes != "[]" {
		var allowedTypes []string
		if err := json.Unmarshal([]byte(modelAccess.AllowedImageTypes), &allowedTypes); err == nil && len(allowedTypes) > 0 {
			ext := strings.ToLower(filepath.Ext(header.Filename))
			ext = strings.TrimPrefix(ext, ".")
			validExt := false
			for _, allowed := range allowedTypes {
				if strings.ToLower(allowed) == ext {
					validExt = true
					break
				}
			}
			if !validExt {
				common.GinError(c, i18nresp.CodeBadRequest, fmt.Sprintf("Unsupported image format. Allowed formats: %v", allowedTypes))
				return
			}
		}
	}

	// 调用 Biz 上传 (纯相对路径存储，不传 host)
	resp, err := biz.GAiFileManager.UploadChatImage(c.Request.Context(), sessionId, file, header)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to upload file: %v", err))
		return
	}

	i18nresp.SuccessResponse(c, resp)
}

