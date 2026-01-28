package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	pb "github.com/kymo-mcp/mcpcan/api/market/ai_agent"
	"github.com/kymo-mcp/mcpcan/internal/market/biz"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// IntegrationTestConfig 集成测试配置
var integrationConfig = mysql.Config{
	Host:                "localhost",
	Port:                31306,
	Database:            "mcpcan",
	Username:            "mcpcan-user",
	Password:            "a6ApqYJIycJJjl",
	ConnectTimeout:      5 * time.Second,
	MaxIdleConns:        10,
	MaxOpenConns:        100,
	HealthCheckInterval: 30 * time.Second,
	MaxRetries:          3,
	RetryInterval:       2 * time.Second,
}

// setupIntegrationTest 初始化集成测试环境
func setupIntegrationTest(t *testing.T) (*gin.Engine, error) {
	// 1. 初始化数据库
	err := mysql.InitDB(&integrationConfig)
	if err != nil {
		t.Logf("Failed to connect to database: %v. Skipping integration test.", err)
		t.Skip("Database not available")
		return nil, err
	}

	// 初始化 Repositories
	mysql.NewAiSessionRepository()
	mysql.NewAiMessageRepository()
	mysql.NewAiModelAccessRepository()

	// 3. 设置 Gin 路由
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// 添加认证中间件的模拟 (直接在测试中设置 Context)
	router.Use(func(c *gin.Context) {
		// 模拟已认证用户
		c.Set("userId", int64(1))
		c.Set("username", "admin")
		c.Next()
	})

	// 注册服务
	aiSessionService := NewAiSessionService(context.Background())
	aiModelAccessService := NewAiModelAccessService(context.Background())

	routerPrefix := "market/ai"
	router.POST(fmt.Sprintf("/%s/sessions", routerPrefix), aiSessionService.CreateHandler)
	router.GET(fmt.Sprintf("/%s/sessions/:id/usage", routerPrefix), aiSessionService.GetSessionUsageHandler)
	router.POST(fmt.Sprintf("/%s/sessions/:id/chat", routerPrefix), aiSessionService.ChatHandler)
	router.POST(fmt.Sprintf("/%s/models", routerPrefix), aiModelAccessService.CreateHandler)

	return router, nil
}

func TestAiIntegration_FullFlow(t *testing.T) {
	router, err := setupIntegrationTest(t)
	if err != nil {
		return // Skipped
	}

	// Step 1: Create Doubao Model
	t.Log("Step 1: Creating Doubao Model...")
	
	createModelReq := pb.CreateModelAccessRequest{
		Name:      "Doubao 1.5 Pro Integration Test",
		Provider:  "openai", // Doubao 兼容 OpenAI 协议
		BaseUrl:   "https://ark.cn-beijing.volces.com/api/v3",
		ApiKey:    "c8bf6018-561c-4dac-a164-900d0c10396b",
		ModelName: "doubao-1-5-pro-32k-250115",
	}
	body, _ := json.Marshal(createModelReq)
	req := httptest.NewRequest(http.MethodPost, "/market/ai/models", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	
	var modelResp struct {
		Code int `json:"code"`
		Data struct {
			Access struct {
				ID int64 `json:"id"`
			} `json:"access"`
		} `json:"data"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &modelResp)
	require.NoError(t, err)
	require.Equal(t, 0, modelResp.Code)
	modelID := modelResp.Data.Access.ID
	t.Logf("Created Model ID: %d", modelID)

	// Step 2: Create Session
	t.Log("Step 2: Creating AI Session...")
	createSessionReq := pb.CreateSessionRequest{
		Name:          "Integration Method Test Session",
		ModelAccessID: modelID,
		MaxContext:    10,
		ToolsConfig:   "{}",
	}
	body, _ = json.Marshal(createSessionReq)
	req = httptest.NewRequest(http.MethodPost, "/market/ai/sessions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var sessionResp struct {
		Code int `json:"code"`
		Data struct {
			Session struct {
				ID int64 `json:"id"`
			} `json:"session"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &sessionResp)
	sessionID := sessionResp.Data.Session.ID
	t.Logf("Created Session ID: %d", sessionID)

	// Step 3: Get Usage (Should be empty initially)
	t.Log("Step 3: Checking Initial Usage...")
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/market/ai/sessions/%d/usage", sessionID), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	require.Equal(t, http.StatusOK, w.Code)
	var usageResp struct {
		Code int `json:"code"`
		Data biz.SessionUsage `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &usageResp)
	assert.Equal(t, 0, usageResp.Data.TotalTokens)

	// Step 4: Chat (Real Request - Optional/Caution)
	// 注意: 这里会发起真实请求。为了集成测试的完整性，我们执行一次简单的对话。
	t.Log("Step 4: Sending Chat Message...")
	chatReq := pb.ChatRequest{
		SessionID: sessionID, // Optional, path param used
		Content:   "HiDoubao",
	}
	body, _ = json.Marshal(chatReq)
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/market/ai/sessions/%d/chat", sessionID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Chat 接口返回的是 SSE 流，我们需要解析
	// 对于集成测试，我们只检查状态码和是否有数据返回
	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "data:") // 检查响应中是否包含 SSE 数据

	// Step 5: Verify Usage Updated
	t.Log("Step 5: Verifying Token Usage Update...")
	// 稍微等待异步保存完成（如果在 goroutine 中保存）
	// 查看业务逻辑，UserMsg 是同步保存的，AssistantMsg 是在流结束后保存的。
	time.Sleep(3 * time.Second)

	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/market/ai/sessions/%d/usage", sessionID), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	json.Unmarshal(w.Body.Bytes(), &usageResp)
	
	t.Logf("Usage Stats: %+v", usageResp.Data)
	assert.Greater(t, usageResp.Data.TotalTokens, 0, "Total tokens should be greater than 0")
	assert.Greater(t, usageResp.Data.TotalMessages, 0, "Total messages should be greater than 0")
}

// TestAiIntegration_ToolCall 测试 MCP 工具调用流程
// 注意: 此测试需要真实的 MCP Server,默认跳过
func TestAiIntegration_ToolCall(t *testing.T) {
	// 检查是否启用工具调用测试
	if os.Getenv("ENABLE_MCP_TESTS") != "true" {
		t.Skip("MCP 工具调用测试需要设置 ENABLE_MCP_TESTS=true 环境变量")
	}

	router, err := setupIntegrationTest(t)
	if err != nil {
		return
	}

	// Step 1: 创建模型配置
	t.Log("Step 1: Creating Model for Tool Call Test...")
	createModelReq := pb.CreateModelAccessRequest{
		Name:      "Tool Call Test Model",
		Provider:  "openai",
		BaseUrl:   "https://ark.cn-beijing.volces.com/api/v3",
		ApiKey:    "c8bf6018-561c-4dac-a164-900d0c10396b",
		ModelName: "doubao-1-5-pro-32k-250115",
	}
	body, _ := json.Marshal(createModelReq)
	req := httptest.NewRequest(http.MethodPost, "/market/ai/models", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var modelResp struct {
		Code int `json:"code"`
		Data struct {
			Access struct {
				ID int64 `json:"id"`
			} `json:"access"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &modelResp)
	modelID := modelResp.Data.Access.ID
	t.Logf("Created Model ID: %d", modelID)

	// Step 2: 创建带 MCP 配置的会话
	t.Log("Step 2: Creating Session with MCP Tools...")
	mcpConfig := `{
		"mcpServers": {
			"filesystem": {
				"command": "npx",
				"args": ["-y", "@anthropic/mcp-filesystem-server", "/tmp"]
			}
		}
	}`
	createSessionReq := pb.CreateSessionRequest{
		Name:          "Tool Call Test Session",
		ModelAccessID: modelID,
		MaxContext:    10,
		ToolsConfig:   mcpConfig,
	}
	body, _ = json.Marshal(createSessionReq)
	req = httptest.NewRequest(http.MethodPost, "/market/ai/sessions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var sessionResp struct {
		Code int `json:"code"`
		Data struct {
			Session struct {
				ID int64 `json:"id"`
			} `json:"session"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &sessionResp)
	sessionID := sessionResp.Data.Session.ID
	t.Logf("Created Session with MCP Tools, ID: %d", sessionID)

	// Step 3: 发送需要工具调用的消息
	t.Log("Step 3: Sending message that requires tool call...")
	chatReq := pb.ChatRequest{
		SessionID: sessionID,
		Content:   "Please list files in /tmp directory using the filesystem tool",
	}
	body, _ = json.Marshal(chatReq)
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/market/ai/sessions/%d/chat", sessionID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	responseBody := w.Body.String()
	t.Logf("Response contains tool call: %v", strings.Contains(responseBody, "tool"))
	
	// 验证响应
	assert.Contains(t, responseBody, "data:", "Response should contain SSE data")
}

// TestAiIntegration_SessionUsageAPI 单独测试 Session Usage API
func TestAiIntegration_SessionUsageAPI(t *testing.T) {
	router, err := setupIntegrationTest(t)
	if err != nil {
		return
	}

	// 使用之前创建的会话测试 Usage API
	// 测试不存在的会话
	req := httptest.NewRequest(http.MethodGet, "/market/ai/sessions/99999/usage", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// 应该返回 200 但数据为空
	assert.Equal(t, http.StatusOK, w.Code)
	
	var usageResp struct {
		Code int            `json:"code"`
		Data biz.SessionUsage `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &usageResp)
	assert.Equal(t, 0, usageResp.Data.TotalMessages, "Non-existent session should have 0 messages")
}
