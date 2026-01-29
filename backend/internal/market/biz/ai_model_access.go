package biz

import (
	"context"
	"fmt"
	"time"

	pb "github.com/kymo-mcp/mcpcan/api/market/ai_agent"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
	"github.com/kymo-mcp/mcpcan/pkg/llm"
	_ "github.com/kymo-mcp/mcpcan/pkg/llm/openai"
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

// Temporary structs until buf generate works
type TestConnectionRequest struct {
	ID        int64  `json:"id"`
	Provider  string `json:"provider"`
	BaseUrl   string `json:"baseUrl"`
	ApiKey    string `json:"apiKey"`
	ModelName string `json:"modelName"`
}

type TestConnectionResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	LatencyMs int64  `json:"latencyMs"`
}

func (b *AiModelAccessBiz) Create(ctx context.Context, req *pb.CreateModelAccessRequest, userID int64) (*model.AiModelAccess, error) {
	// Validate provider
	if req.Provider == "" {
		return nil, fmt.Errorf("provider is required")
	}
	if !llm.IsValidProvider(req.Provider) {
		return nil, fmt.Errorf("unsupported provider: %s, supported providers: %v", req.Provider, llm.GetSupportedProviderList())
	}

	modelAccess := &model.AiModelAccess{
		UserID:   userID,
		Name:     req.Name,
		Provider: req.Provider,
		ApiKey:   req.ApiKey,
		BaseUrl:  req.BaseUrl,
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
		modelAccess.ApiKey = req.ApiKey
	}
	if req.BaseUrl != "" {
		modelAccess.BaseUrl = req.BaseUrl
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

func (b *AiModelAccessBiz) TestConnection(ctx context.Context, req *TestConnectionRequest) (*TestConnectionResponse, error) {
	var (
		providerStr string
		baseUrl     string
		apiKey      string
		modelName   string
	)

	// 1. Determine config source
	if req.ID > 0 {
		// Load from DB
		modelAccess, err := mysql.AiModelAccessRepo.FindByID(ctx, req.ID)
		if err != nil {
			return nil, fmt.Errorf("model access not found")
		}
		providerStr = modelAccess.Provider
		baseUrl = modelAccess.BaseUrl
		apiKey = modelAccess.ApiKey
		// ModelName must be provided in request for testing
		modelName = req.ModelName
	} else {
		// Use request params
		providerStr = req.Provider
		baseUrl = req.BaseUrl
		apiKey = req.ApiKey
		modelName = req.ModelName
	}

	// 2. Init Provider
	providerType := llm.ProviderOpenAI
	if providerStr != "" {
		providerType = llm.ProviderType(providerStr)
	}

	config := llm.ProviderConfig{
		BaseURL: baseUrl,
		APIKey:  apiKey,
	}

	provider, err := llm.NewProvider(providerType, config)
	if err != nil {
		return &TestConnectionResponse{
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
		return &TestConnectionResponse{
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
		return &TestConnectionResponse{
			Success:   false,
			Message:   fmt.Sprintf("Connection error during stream: %v", lastErr),
			LatencyMs: latency,
		}, nil
	}

	// Note: Empty content check omitted as some models might filter "Hi" or return empty but valid stream end
	
	return &TestConnectionResponse{
		Success:   true,
		Message:   "Connection successful",
		LatencyMs: latency,
	}, nil
}
