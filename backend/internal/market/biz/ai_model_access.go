package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	pb "github.com/kymo-mcp/mcpcan/api/market/ai_agent"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
	llm "github.com/kymo-mcp/mcpcan/pkg/llm_adapter"
)

type AiModelAccessBiz struct {
	ctx context.Context
}

var GAiModelAccessBiz *AiModelAccessBiz

func init() {
	GAiModelAccessBiz = NewAiModelAccessBiz(context.Background())
}

func NewAiModelAccessBiz(ctx context.Context) *AiModelAccessBiz {
	return &AiModelAccessBiz{
		ctx: ctx,
	}
}

// structs removed in favor of pb package

// isMaskedKey 判断 apiKey 是否为脱敏后的占位字符串（含 ****）
// 前端展示用的脱敏 key 不能写回数据库，也不能用于实际请求
func isMaskedKey(key string) bool {
	return len(key) > 0 && len(key) < 64 && (containsStr(key, "****") || containsStr(key, "***"))
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > 0 && func() bool {
		for i := 0; i <= len(s)-len(substr); i++ {
			if s[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	}()))
}

func (b *AiModelAccessBiz) Create(ctx context.Context, req *pb.CreateModelAccessRequest, userID int64) (*model.AiModelAccess, error) {
	// Validate provider
	if req.Provider == "" {
		return nil, fmt.Errorf("provider is required")
	}
	if !llm.IsValidProvider(req.Provider) {
		return nil, fmt.Errorf("unsupported provider: %s, supported providers: %v", req.Provider, llm.GetSupportedProviderList())
	}

	
	allowedModelsJson, _ := json.Marshal(req.AllowedModels)
	if req.AllowedModels == nil {
		allowedModelsJson = []byte("")
	}

	modelAccess := &model.AiModelAccess{
		UserID:        userID,
		Name:          req.Name,
		Provider:      req.Provider,
		ApiKey:        req.ApiKey,
		BaseUrl:       req.BaseUrl,
		AllowedModels: string(allowedModelsJson),
	}

	if err := mysql.AiModelAccessRepo.Create(ctx, modelAccess); err != nil {
		return nil, err
	}
	return modelAccess, nil
}

func (b *AiModelAccessBiz) Update(ctx context.Context, req *pb.UpdateModelAccessRequest) (*model.AiModelAccess, error) {
	modelAccess, err := mysql.AiModelAccessRepo.FindByID(ctx, req.Id)
	if err != nil {
		return nil, fmt.Errorf("model access not found")
	}

	if req.Name != "" {
		modelAccess.Name = req.Name
	}
	if req.Provider != "" {
		// Validate provider before update
		if !llm.IsValidProvider(req.Provider) {
			return nil, fmt.Errorf("unsupported provider: %s, supported providers: %v", req.Provider, llm.GetSupportedProviderList())
		}
		modelAccess.Provider = req.Provider
	}
	if req.ApiKey != "" {
		// 防止前端将脱敏值（含 ****）回写到数据库
		if isMaskedKey(req.ApiKey) {
			// 脱敏占位符，跳过更新（保留数据库中原始 key）
		} else {
			modelAccess.ApiKey = req.ApiKey
		}
	}
	if req.BaseUrl != "" {
		modelAccess.BaseUrl = req.BaseUrl
	}
	// AllowedModels 允许显式清空
	if req.AllowedModels != nil {
		allowedModelsJson, _ := json.Marshal(req.AllowedModels)
		modelAccess.AllowedModels = string(allowedModelsJson)
	}

	if err := mysql.AiModelAccessRepo.Update(ctx, modelAccess); err != nil {
		return nil, err
	}
	return modelAccess, nil
}

func (b *AiModelAccessBiz) Delete(ctx context.Context, id int64) error {
	return mysql.AiModelAccessRepo.Delete(ctx, id)
}

func (b *AiModelAccessBiz) Get(ctx context.Context, id int64) (*model.AiModelAccess, error) {
	return mysql.AiModelAccessRepo.FindByID(ctx, id)
}

func (b *AiModelAccessBiz) List(ctx context.Context, userID int64, page, pageSize int) ([]*model.AiModelAccess, int64, error) {
	accesses, err := mysql.AiModelAccessRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	total := int64(len(accesses))
	if page <= 0 {
		page = 1
	}
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

	return accesses[start:end], total, nil
}

func (b *AiModelAccessBiz) TestConnection(ctx context.Context, req *pb.TestConnectionRequest) (*pb.TestConnectionResponse, error) {
	if req.Id <= 0 {
		return nil, fmt.Errorf("model access id is required")
	}
	if req.ModelName == "" {
		return nil, fmt.Errorf("model name is required")
	}

	// 1. Determine config source: Load from DB purely based on ID
	modelAccess, err := mysql.AiModelAccessRepo.FindByID(ctx, req.Id)
	if err != nil {
		return nil, fmt.Errorf("model access not found")
	}
	
	providerStr := modelAccess.Provider
	baseUrl := modelAccess.BaseUrl
	apiKey := modelAccess.ApiKey
	modelName := req.ModelName

	// 安全检查：如果 DB 中的 apiKey 是脱敏值（历史脏数据或被误写入），直接报错
	if isMaskedKey(apiKey) {
		return &pb.TestConnectionResponse{
			Success: false,
			Message: "API Key 已损坏（存储的是脱敏占位符而非真实 Key）。请重新编辑该配置并重新输入完整的 API Key 后再测试。",
		}, nil
	}

	// 2. Init Provider
	providerType := llm.ProviderOpenAI
	if providerStr != "" {
		providerType = llm.ProviderType(providerStr)
	}

	config := llm.ProviderConfig{
		BaseURL:  baseUrl,
		APIKey:   apiKey,
		ProxyURL: llm.GlobalProxyURL,
	}

	provider, err := llm.NewProvider(providerType, config)
	if err != nil {
		return &pb.TestConnectionResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to init provider: %v", err),
		}, nil // Return success=false as a valid response, not system error
	}

	// 3. Test Connection
	chatReq := llm.ChatRequest{
		Model: modelName,
		Messages: []llm.Message{
			{Role: "user", Content: "Hi"},
		},
		MaxTokens: 5,
		Stream:    true,
	}

	start := time.Now()
	stream, err := provider.StreamChat(ctx, chatReq)
	if err != nil {
		return &pb.TestConnectionResponse{
			Success: false,
			Message: fmt.Sprintf("Connection failed: %v", err),
		}, nil
	}

	// Consume stream to ensure it works
	var lastErr error
	for resp := range stream {
		if resp.Error != nil {
			lastErr = resp.Error
			break
		}
	}
	latency := time.Since(start).Milliseconds()

	if lastErr != nil {
		return &pb.TestConnectionResponse{
			Success:   false,
			Message:   fmt.Sprintf("Connection error during stream: %v. (Hint: If using a custom model, ensure it is supported by the configured Base URL/Provider endpoint)", lastErr),
			LatencyMs: latency,
		}, nil
	}

	// Note: Empty content check omitted as some models might filter "Hi" or return empty but valid stream end
	
	return &pb.TestConnectionResponse{
		Success:   true,
		Message:   "Connection successful",
		LatencyMs: latency,
	}, nil
}
