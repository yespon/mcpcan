package service

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	pb "github.com/kymo-mcp/mcpcan/api/market/ai_agent"
	"github.com/kymo-mcp/mcpcan/internal/market/biz"
	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	i18nresp "github.com/kymo-mcp/mcpcan/pkg/i18n"
	"github.com/kymo-mcp/mcpcan/pkg/llm"
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

	c.JSON(200, resp)
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

	c.JSON(200, resp)
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
	var req pb.UpdateModelAccessRequest
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

	// TODO: Get user id from context
	userID := int64(1)

	accesses, total, err := biz.GAiModelAccessBiz.List(c.Request.Context(), userID, int(req.Page), int(req.PageSize))
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to list ai model accesses: %s", err.Error()))
		return
	}

	var pbAccesses []*pb.AiModelAccess
	for _, access := range accesses {
		pbAccesses = append(pbAccesses, s.convertModelToProto(access))
	}

	resp := &pb.ListModelAccessResponse{
		List:  pbAccesses,
		Total: total,
	}
	i18nresp.SuccessResponse(c, resp)
}

// GetAvailableModelsHandler gets available models for selection (no pagination)
func (s *AiModelAccessService) GetAvailableModelsHandler(c *gin.Context) {
	// TODO: Get user id from context
	userID := int64(1)

	// Fetch all (use a large limit)
	accesses, _, err := biz.GAiModelAccessBiz.List(c.Request.Context(), userID, 1, 1000)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to list ai model accesses: %s", err.Error()))
		return
	}

	var pbAccesses []*pb.AiModelAccess
	for _, access := range accesses {
		pbAccesses = append(pbAccesses, s.convertModelToProto(access))
	}

	resp := &pb.ListModelAccessResponse{
		List:  pbAccesses,
		Total: int64(len(pbAccesses)),
	}
	i18nresp.SuccessResponse(c, resp)
}

// GetSupportedModelsHandler gets supported models
func (s *AiModelAccessService) GetSupportedModelsHandler(c *gin.Context) {
	// 1. OpenAI Models (Generated from go-openai)
	openAIModels := llm.SupportedOpenAIModels

	// 2. DeepSeek Models
	deepSeekModels := llm.DeepSeekModels

	// 3. Aliyun Qwen
	qwenModels := llm.QwenModels

	// 4. Volcengine Doubao
	doubaoModels := llm.DoubaoModels

	// 5. Construct Response
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
			{
				Id:     "qwen",
				Name:   "Aliyun Qwen",
				Models: qwenModels,
			},
			{
				Id:     "doubao",
				Name:   "Volcengine Doubao",
				Models: doubaoModels,
			},
		},
	}

	i18nresp.SuccessResponse(c, resp)
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
		ModelName:  m.ModelName,
		CreateTime: m.CreateTime.Unix(),
		UpdateTime: m.UpdateTime.Unix(),
	}
}
