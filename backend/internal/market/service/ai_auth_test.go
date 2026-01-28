package service

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	pb "github.com/kymo-mcp/mcpcan/api/market/ai_agent"
	"github.com/kymo-mcp/mcpcan/pkg/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestRouter 创建测试用的 Gin 路由
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

// generateTestToken 生成测试用的 JWT Token
func generateTestToken(userID int64, username string, secret string, expiresIn time.Duration) (string, error) {
	jwtManager := jwt.NewManager(&jwt.Config{
		Secret:  secret,
		Expires: expiresIn,
	})
	return jwtManager.GenerateToken(userID, username)
}

// TestAiSessionCreateHandler_Authentication 测试创建会话的认证功能
// 注意: 此测试需要数据库连接,请使用 TestAiIntegration_FullFlow 进行完整集成测试
func TestAiSessionCreateHandler_Authentication(t *testing.T) {
	t.Skip("此测试需要数据库连接,请运行 TestAiIntegration_FullFlow 进行完整集成测试")
	
	router := setupTestRouter()
	service := NewAiSessionService(nil)
	
	// 注册路由
	router.POST("/ai/sessions", service.CreateHandler)

	testSecret := "test-secret-key"

	tests := []struct {
		name           string
		setupAuth      func() string
		expectedStatus int
		expectedError  string
	}{
		{
			name: "无 Token - 应返回 401",
			setupAuth: func() string {
				return "" // 不提供 Token
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "user not authenticated",
		},
		{
			name: "有效 Token - 应成功提取用户 ID",
			setupAuth: func() string {
				token, _ := generateTestToken(123, "testuser", testSecret, time.Hour)
				return "Bearer " + token
			},
			expectedStatus: http.StatusOK,
			expectedError:  "",
		},
		{
			name: "过期 Token - 应返回 401",
			setupAuth: func() string {
				// 生成一个已过期的 Token (过期时间为 -1 小时)
				token, _ := generateTestToken(123, "testuser", testSecret, -time.Hour)
				return "Bearer " + token
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "user not authenticated",
		},
		{
			name: "错误格式的 Token - 应返回 401",
			setupAuth: func() string {
				return "Bearer invalid-token-format"
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "user not authenticated",
		},
		{
			name: "缺少 Bearer 前缀 - 应返回 401",
			setupAuth: func() string {
				token, _ := generateTestToken(123, "testuser", testSecret, time.Hour)
				return token // 不加 Bearer 前缀
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "user not authenticated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 准备请求体
			reqBody := pb.CreateSessionRequest{
				Name:          "Test Session",
				ModelAccessID: 1,
				MaxContext:    10,
				ToolsConfig:   "{}",
			}
			bodyBytes, _ := json.Marshal(reqBody)

			// 创建请求
			req := httptest.NewRequest(http.MethodPost, "/ai/sessions", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			
			// 设置认证头
			authHeader := tt.setupAuth()
			if authHeader != "" {
				req.Header.Set("Authorization", authHeader)
			}

			// 创建响应记录器
			w := httptest.NewRecorder()

			// 手动设置上下文(模拟认证中间件的行为)
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			
			// 如果是有效 Token,模拟中间件设置用户信息
			if tt.name == "有效 Token - 应成功提取用户 ID" {
				c.Set("userId", int64(123))
				c.Set("username", "testuser")
			}

			// 执行处理器
			service.CreateHandler(c)

			// 验证响应状态码
			assert.Equal(t, tt.expectedStatus, w.Code, "HTTP status code mismatch")

			// 如果期望错误,验证错误消息
			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err, "Failed to parse response body")
				
				// 检查是否包含错误信息
				if message, ok := response["message"].(string); ok {
					assert.Contains(t, message, tt.expectedError, "Error message mismatch")
				}
			}
		})
	}
}

// TestAiSessionListHandler_Authentication 测试列表查询的认证功能
func TestAiSessionListHandler_Authentication(t *testing.T) {
	t.Skip("此测试需要数据库连接,请运行 TestAiIntegration_FullFlow 进行完整集成测试")
	
	router := setupTestRouter()
	service := NewAiSessionService(nil)
	
	router.GET("/ai/sessions", service.ListHandler)

	tests := []struct {
		name           string
		setUserContext bool
		userID         int64
		expectedStatus int
	}{
		{
			name:           "无用户上下文 - 应返回 401",
			setUserContext: false,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "有用户上下文 - 应成功",
			setUserContext: true,
			userID:         456,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/ai/sessions?page=1&pageSize=10", nil)
			w := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(w)
			c.Request = req

			if tt.setUserContext {
				c.Set("userId", tt.userID)
				c.Set("username", "testuser")
			}

			service.ListHandler(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestAiModelAccessCreateHandler_Authentication 测试模型配置创建的认证功能
func TestAiModelAccessCreateHandler_Authentication(t *testing.T) {
	t.Skip("此测试需要数据库连接,请运行 TestAiIntegration_FullFlow 进行完整集成测试")
	
	router := setupTestRouter()
	service := NewAiModelAccessService(nil)
	
	router.POST("/ai/models", service.CreateHandler)

	tests := []struct {
		name           string
		setUserContext bool
		userID         int64
		expectedStatus int
	}{
		{
			name:           "无用户上下文 - 应返回 401",
			setUserContext: false,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "有用户上下文 - 应成功",
			setUserContext: true,
			userID:         789,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := pb.CreateModelAccessRequest{
				Name:      "Test Model",
				Provider:  "openai",
				BaseUrl:   "https://api.openai.com/v1",
				ApiKey:    "sk-test-key",
				ModelName: "gpt-4",
			}
			bodyBytes, _ := json.Marshal(reqBody)

			req := httptest.NewRequest(http.MethodPost, "/ai/models", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			c, _ := gin.CreateTestContext(w)
			c.Request = req

			if tt.setUserContext {
				c.Set("userId", tt.userID)
				c.Set("username", "testuser")
			}

			service.CreateHandler(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// TestUserDataIsolation 测试用户数据隔离
func TestUserDataIsolation(t *testing.T) {
	// 这个测试需要实际的数据库连接,这里只做框架演示
	t.Skip("需要数据库连接,跳过集成测试")

	// 测试场景:
	// 1. 用户 A 创建会话
	// 2. 用户 B 尝试访问用户 A 的会话 -> 应该看不到
	// 3. 用户 A 查询会话列表 -> 只能看到自己的会话
}

// BenchmarkGetUserIDFromContext 性能测试
func BenchmarkGetUserIDFromContext(b *testing.B) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("userId", int64(123))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 注意: 实际应该导入 common 包并使用 common.GetUserIDFromContext
		// 这里仅作为性能基准测试框架
		_, _ = c.Get("userId")
	}
}

