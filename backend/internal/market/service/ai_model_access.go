package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	pb "github.com/kymo-mcp/mcpcan/api/market/ai_agent"
	"github.com/kymo-mcp/mcpcan/internal/market/biz"
	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	i18nresp "github.com/kymo-mcp/mcpcan/pkg/i18n"
	"github.com/kymo-mcp/mcpcan/pkg/llm/models"
)

type AiModelAccessService struct {
	ctx context.Context
}

func NewAiModelAccessService(ctx context.Context) *AiModelAccessService {
	return &AiModelAccessService{
		ctx: ctx,
	}
}

	// TestConnectionHandler tests connection to the model
func (s *AiModelAccessService) TestConnectionHandler(c *gin.Context) {
	var req biz.TestConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.GinError(c, i18nresp.CodeBadRequest, err.Error())
		return
	}

	resp, err := biz.GAiModelAccessBiz.TestConnection(c.Request.Context(), &req)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	i18nresp.SuccessResponse(c, resp)
}

// TestConnectionWithIdHandler tests connection to an existing model by ID
func (s *AiModelAccessService) TestConnectionWithIdHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		common.GinError(c, i18nresp.CodeBadRequest, "invalid model access id")
		return
	}

	req := biz.TestConnectionRequest{
		ID: id,
	}

	resp, err := biz.GAiModelAccessBiz.TestConnection(c.Request.Context(), &req)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	i18nresp.SuccessResponse(c, resp)
}

// CreateHandler creates a new ai model access
func (s *AiModelAccessService) CreateHandler(c *gin.Context) {
	var req biz.CreateModelAccessRequest
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

	modelAccess, err := biz.GAiModelAccessBiz.Create(c.Request.Context(), &req, userID)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to create ai model access: %s", err.Error()))
		return
	}

	resp := &pb.CreateModelAccessResponse{
		Access: s.convertModelToProto(modelAccess),
	}
	i18nresp.SuccessResponse(c, resp)
}

// UpdateHandler updates ai model access
func (s *AiModelAccessService) UpdateHandler(c *gin.Context) {
	var req biz.UpdateModelAccessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.GinError(c, i18nresp.CodeBadRequest, err.Error())
		return
	}

	modelAccess, err := biz.GAiModelAccessBiz.Update(c.Request.Context(), &req)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to update ai model access: %s", err.Error()))
		return
	}

	resp := &pb.UpdateModelAccessResponse{
		Access: s.convertModelToProto(modelAccess),
	}
	i18nresp.SuccessResponse(c, resp)
}

// DeleteHandler deletes ai model access
func (s *AiModelAccessService) DeleteHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		common.GinError(c, i18nresp.CodeBadRequest, "invalid model access id")
		return
	}

	if err := biz.GAiModelAccessBiz.Delete(c.Request.Context(), id); err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to delete ai model access: %s", err.Error()))
		return
	}

	resp := &pb.DeleteModelAccessResponse{
		Success: true,
	}
	i18nresp.SuccessResponse(c, resp)
}

// GetHandler gets ai model access detail
func (s *AiModelAccessService) GetHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		common.GinError(c, i18nresp.CodeBadRequest, "invalid model access id")
		return
	}

	modelAccess, err := biz.GAiModelAccessBiz.Get(c.Request.Context(), id)
	if err != nil {
		common.GinError(c, i18nresp.CodeNotFound, "model access not found")
		return
	}

	resp := &pb.GetModelAccessResponse{
		Access: s.convertModelToProto(modelAccess),
	}
	i18nresp.SuccessResponse(c, resp)
}

// ListHandler lists ai model accesses
func (s *AiModelAccessService) ListHandler(c *gin.Context) {
	var req pb.ListModelAccessRequest
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

	accesses, total, err := biz.GAiModelAccessBiz.List(c.Request.Context(), userID, int(req.Page), int(req.PageSize))
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to list ai model accesses: %s", err.Error()))
		return
	}

	var respList []*modelAccessResp
	for _, access := range accesses {
		respList = append(respList, s.convertModelToRespFull(access))
	}

	resp := map[string]interface{}{
		"list":  respList,
		"total": total,
	}
	i18nresp.SuccessResponse(c, resp)
}

// GetAvailableModelsHandler gets available models for selection (no pagination)
func (s *AiModelAccessService) GetAvailableModelsHandler(c *gin.Context) {
	// 从认证上下文中获取当前用户 ID
	userID, err := common.GetUserIDFromContext(c)
	if err != nil {
		common.GinError(c, i18nresp.CodeUnauthorized, "user not authenticated")
		return
	}

	// Fetch all (use a large limit)
	accesses, _, err := biz.GAiModelAccessBiz.List(c.Request.Context(), userID, 1, 1000)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to list ai model accesses: %s", err.Error()))
		return
	}

	var respList []*modelAccessResp
	for _, access := range accesses {
		respList = append(respList, s.convertModelToRespFull(access))
	}

	resp := map[string]interface{}{
		"list":  respList,
		"total": int64(len(respList)),
	}
	i18nresp.SuccessResponse(c, resp)
}

// GetSupportedModelsHandler gets supported models with registration URLs
func (s *AiModelAccessService) GetSupportedModelsHandler(c *gin.Context) {
	// 使用 models.AllProviders 动态构建响应
	var providers []*pb.ModelProvider
	for _, p := range models.AllProviders {
		providers = append(providers, &pb.ModelProvider{
			Id:          p.ID,
			Name:        p.Name,
			Models:      p.GetModelIDs(),
			RegisterUrl: p.RegisterURL,
			DocsUrl:     p.DocsURL,
			BaseUrl:     p.BaseURL,
		})
	}

	resp := &pb.GetSupportedModelsResponse{
		Providers: providers,
	}

	i18nresp.SuccessResponse(c, resp)
}


// modelAccessResp 是 AiModelAccess 的扮展响应结构，展展 allowedModels 字段
type modelAccessResp struct {
	Id            int64    `json:"id"`
	Name          string   `json:"name"`
	Provider      string   `json:"provider"`
	ApiKey        string   `json:"apiKey"`
	BaseUrl       string   `json:"baseUrl"`
	CreateTime    int64    `json:"createTime"`
	UpdateTime    int64    `json:"updateTime"`
	AllowedModels []string `json:"allowedModels"` // 允许使用的模型 ID 列表，为空表示不限制
}

func (s *AiModelAccessService) convertModelToProto(m *model.AiModelAccess) *pb.AiModelAccess {
	// Mask API Key
	maskedKey := m.ApiKey
	if len(maskedKey) > 8 {
		maskedKey = maskedKey[:3] + "****" + maskedKey[len(maskedKey)-4:]
	} else if len(maskedKey) > 0 {
		maskedKey = "****"
	}

	return &pb.AiModelAccess{
		Id:         m.ID,
		Name:       m.Name,
		Provider:   m.Provider,
		ApiKey:     maskedKey,
		BaseUrl:    m.BaseUrl,
		// 利用 ModelName 字段传输 allowedModels JSON（临时方案，直到 proto 重新生成）
		ModelName:  m.AllowedModels,
		CreateTime: m.CreateTime.Unix(),
		UpdateTime: m.UpdateTime.Unix(),
	}
}

// convertModelToRespFull 返回包含 allowedModels 的完整响应结构
func (s *AiModelAccessService) convertModelToRespFull(m *model.AiModelAccess) *modelAccessResp {
	// Mask API Key
	maskedKey := m.ApiKey
	if len(maskedKey) > 8 {
		maskedKey = maskedKey[:3] + "****" + maskedKey[len(maskedKey)-4:]
	} else if len(maskedKey) > 0 {
		maskedKey = "****"
	}

	var allowedModels []string
	if m.AllowedModels != "" && m.AllowedModels != "[]" {
		// 尝试解析 JSON 数组
		_ = json.Unmarshal([]byte(m.AllowedModels), &allowedModels)
	}

	return &modelAccessResp{
		Id:            m.ID,
		Name:          m.Name,
		Provider:      m.Provider,
		ApiKey:        maskedKey,
		BaseUrl:       m.BaseUrl,
		CreateTime:    m.CreateTime.Unix(),
		UpdateTime:    m.UpdateTime.Unix(),
		AllowedModels: allowedModels,
	}
}

