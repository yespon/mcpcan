package service

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/stretchr/testify/assert"
)

// TestAuthenticationFlow_UserContextExtraction 测试认证流程中的用户上下文提取
func TestAuthenticationFlow_UserContextExtraction(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupContext   func(*gin.Context)
		expectUserID   int64
		expectError    bool
		expectedErrMsg string
	}{
		{
			name: "成功提取用户 ID - int64 类型",
			setupContext: func(c *gin.Context) {
				c.Set("userId", int64(123))
				c.Set("username", "testuser")
			},
			expectUserID: 123,
			expectError:  false,
		},
		{
			name: "成功提取用户 ID - int 类型",
			setupContext: func(c *gin.Context) {
				c.Set("userId", int(456))
				c.Set("username", "testuser2")
			},
			expectUserID: 456,
			expectError:  false,
		},
		{
			name: "用户未认证 - 缺少 userId",
			setupContext: func(c *gin.Context) {
				// 不设置 userId
			},
			expectError:    true,
			expectedErrMsg: "user id not found in context",
		},
		{
			name: "用户 ID 类型错误",
			setupContext: func(c *gin.Context) {
				c.Set("userId", "invalid_string")
			},
			expectError:    true,
			expectedErrMsg: "invalid user id type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试上下文
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// 设置上下文
			tt.setupContext(c)

			// 提取用户 ID
			userID, err := common.GetUserIDFromContext(c)

			if tt.expectError {
				assert.Error(t, err)
				if tt.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectUserID, userID)
			}
		})
	}
}

// TestAuthenticationFlow_UsernameExtraction 测试用户名提取
func TestAuthenticationFlow_UsernameExtraction(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name             string
		setupContext     func(*gin.Context)
		expectUsername   string
		expectError      bool
		expectedErrMsg   string
	}{
		{
			name: "成功提取用户名",
			setupContext: func(c *gin.Context) {
				c.Set("username", "alice")
			},
			expectUsername: "alice",
			expectError:    false,
		},
		{
			name: "用户名不存在",
			setupContext: func(c *gin.Context) {
				// 不设置 username
			},
			expectError:    true,
			expectedErrMsg: "username not found in context",
		},
		{
			name: "用户名类型错误",
			setupContext: func(c *gin.Context) {
				c.Set("username", 12345)
			},
			expectError:    true,
			expectedErrMsg: "invalid username type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			tt.setupContext(c)

			username, err := common.GetUsernameFromContext(c)

			if tt.expectError {
				assert.Error(t, err)
				if tt.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectUsername, username)
			}
		})
	}
}

// TestAuthenticationFlow_MultipleUsersSeparation 测试多用户数据隔离逻辑
func TestAuthenticationFlow_MultipleUsersSeparation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 模拟两个不同的用户请求
	user1Context := func() *gin.Context {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userId", int64(100))
		c.Set("username", "user1")
		return c
	}

	user2Context := func() *gin.Context {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userId", int64(200))
		c.Set("username", "user2")
		return c
	}

	// 验证用户 1
	c1 := user1Context()
	userID1, err1 := common.GetUserIDFromContext(c1)
	assert.NoError(t, err1)
	assert.Equal(t, int64(100), userID1)

	username1, err1 := common.GetUsernameFromContext(c1)
	assert.NoError(t, err1)
	assert.Equal(t, "user1", username1)

	// 验证用户 2
	c2 := user2Context()
	userID2, err2 := common.GetUserIDFromContext(c2)
	assert.NoError(t, err2)
	assert.Equal(t, int64(200), userID2)

	username2, err2 := common.GetUsernameFromContext(c2)
	assert.NoError(t, err2)
	assert.Equal(t, "user2", username2)

	// 确保两个用户的 ID 不同
	assert.NotEqual(t, userID1, userID2, "不同用户应该有不同的 ID")
}

// TestAuthenticationFlow_MustGetUserID 测试 MustGetUserID 的 panic 行为
func TestAuthenticationFlow_MustGetUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("有效用户 ID - 不应 panic", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userId", int64(999))

		assert.NotPanics(t, func() {
			userID := common.MustGetUserID(c)
			assert.Equal(t, int64(999), userID)
		})
	})

	t.Run("无用户 ID - 应 panic", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		// 不设置 userId

		assert.Panics(t, func() {
			common.MustGetUserID(c)
		})
	})
}
