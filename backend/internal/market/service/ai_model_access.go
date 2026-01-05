package service

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	pb "github.com/kymo-mcp/mcpcan/api/market/ai_agent"
	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
	i18nresp "github.com/kymo-mcp/mcpcan/pkg/i18n"
	"github.com/sashabaranov/go-openai"
)

type AiModelAccessService struct {
	ctx context.Context
}

func NewAiModelAccessService(ctx context.Context) *AiModelAccessService {
	return &AiModelAccessService{
		ctx: ctx,
	}
}

// CreateHandler creates a new ai model access
func (s *AiModelAccessService) CreateHandler(c *gin.Context) {
	var req pb.CreateModelAccessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.GinError(c, i18nresp.CodeBadRequest, err.Error())
		return
	}

	// TODO: Get user id from context
	userID := int64(1)

	modelAccess := &model.AiModelAccess{
		UserID:    userID,
		Name:      req.Name,
		Provider:  req.Provider,
		ApiKey:    req.ApiKey,
		BaseUrl:   req.BaseUrl,
		ModelName: req.ModelName,
	}

	if err := mysql.AiModelAccessRepo.Create(s.ctx, modelAccess); err != nil {
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
	var req pb.UpdateModelAccessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.GinError(c, i18nresp.CodeBadRequest, err.Error())
		return
	}

	modelAccess, err := mysql.AiModelAccessRepo.FindByID(s.ctx, req.Id)
	if err != nil {
		common.GinError(c, i18nresp.CodeNotFound, "model access not found")
		return
	}

	if req.Name != "" {
		modelAccess.Name = req.Name
	}
	if req.Provider != "" {
		modelAccess.Provider = req.Provider
	}
	if req.ApiKey != "" {
		modelAccess.ApiKey = req.ApiKey
	}
	if req.BaseUrl != "" {
		modelAccess.BaseUrl = req.BaseUrl
	}
	if req.ModelName != "" {
		modelAccess.ModelName = req.ModelName
	}

	if err := mysql.AiModelAccessRepo.Update(s.ctx, modelAccess); err != nil {
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

	if err := mysql.AiModelAccessRepo.Delete(s.ctx, id); err != nil {
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

	modelAccess, err := mysql.AiModelAccessRepo.FindByID(s.ctx, id)
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

	// TODO: Get user id from context
	userID := int64(1)

	accesses, err := mysql.AiModelAccessRepo.FindByUserID(s.ctx, userID)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to list ai model accesses: %s", err.Error()))
		return
	}

	// Manual pagination
	total := int64(len(accesses))
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

	var pbAccesses []*pb.AiModelAccess
	for _, access := range accesses[start:end] {
		pbAccesses = append(pbAccesses, s.convertModelToProto(access))
	}

	resp := &pb.ListModelAccessResponse{
		List:  pbAccesses,
		Total: total,
	}
	i18nresp.SuccessResponse(c, resp)
}

// GetSupportedModelsHandler gets supported models
func (s *AiModelAccessService) GetSupportedModelsHandler(c *gin.Context) {
	// 1. OpenAI Models (Referencing go-openai constants)
	openAIModels := []string{
		openai.GPT4o,
		openai.GPT4oMini,
		openai.GPT4Turbo,
		openai.GPT4,
		openai.GPT3Dot5Turbo,
		openai.GPT3Dot5Turbo16K,
	}

	// 2. DeepSeek Models (Manual)
	deepSeekModels := []string{
		"deepseek-chat",
		"deepseek-coder",
	}

	// 3. Construct Response
	resp := &pb.GetSupportedModelsResponse{
		Providers: []*pb.ModelProvider{
			{
				Id:     "openai",
				Name:   "OpenAI",
				Models: openAIModels,
			},
			{
				Id:     "deepseek",
				Name:   "DeepSeek",
				Models: deepSeekModels,
			},
			// Add more providers here (e.g. Moonshot, Qwen)
		},
	}

	i18nresp.SuccessResponse(c, resp)
}

func (s *AiModelAccessService) convertModelToProto(m *model.AiModelAccess) *pb.AiModelAccess {
	return &pb.AiModelAccess{
		Id:         m.ID,
		Name:       m.Name,
		Provider:   m.Provider,
		ApiKey:     m.ApiKey,
		BaseUrl:    m.BaseUrl,
		ModelName:  m.ModelName,
		CreateTime: m.CreateTime.Unix(),
		UpdateTime: m.UpdateTime.Unix(),
	}
}
