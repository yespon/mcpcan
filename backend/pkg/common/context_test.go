package common

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetUserIDFromContext(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(*gin.Context)
		wantID    int64
		wantErr   bool
	}{
		{
			name: "成功提取 int64 类型的用户 ID",
			setupFunc: func(c *gin.Context) {
				c.Set("userId", int64(123))
			},
			wantID:  123,
			wantErr: false,
		},
		{
			name: "成功提取 int 类型的用户 ID",
			setupFunc: func(c *gin.Context) {
				c.Set("userId", int(456))
			},
			wantID:  456,
			wantErr: false,
		},
		{
			name: "成功提取 uint 类型的用户 ID",
			setupFunc: func(c *gin.Context) {
				c.Set("userId", uint(789))
			},
			wantID:  789,
			wantErr: false,
		},
		{
			name: "成功提取 float64 类型的用户 ID",
			setupFunc: func(c *gin.Context) {
				c.Set("userId", float64(999))
			},
			wantID:  999,
			wantErr: false,
		},
		{
			name: "用户 ID 不存在",
			setupFunc: func(c *gin.Context) {
				// 不设置 userId
			},
			wantID:  0,
			wantErr: true,
		},
		{
			name: "用户 ID 类型错误",
			setupFunc: func(c *gin.Context) {
				c.Set("userId", "invalid_string")
			},
			wantID:  0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(nil)
			tt.setupFunc(c)

			gotID, err := GetUserIDFromContext(c)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantID, gotID)
			}
		})
	}
}

func TestGetUsernameFromContext(t *testing.T) {
	tests := []struct {
		name         string
		setupFunc    func(*gin.Context)
		wantUsername string
		wantErr      bool
	}{
		{
			name: "成功提取用户名",
			setupFunc: func(c *gin.Context) {
				c.Set("username", "testuser")
			},
			wantUsername: "testuser",
			wantErr:      false,
		},
		{
			name: "用户名不存在",
			setupFunc: func(c *gin.Context) {
				// 不设置 username
			},
			wantUsername: "",
			wantErr:      true,
		},
		{
			name: "用户名类型错误",
			setupFunc: func(c *gin.Context) {
				c.Set("username", 12345)
			},
			wantUsername: "",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(nil)
			tt.setupFunc(c)

			gotUsername, err := GetUsernameFromContext(c)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantUsername, gotUsername)
			}
		})
	}
}

func TestMustGetUserID(t *testing.T) {
	t.Run("成功提取用户 ID", func(t *testing.T) {
		c, _ := gin.CreateTestContext(nil)
		c.Set("userId", int64(123))

		userID := MustGetUserID(c)
		assert.Equal(t, int64(123), userID)
	})

	t.Run("用户 ID 不存在时 panic", func(t *testing.T) {
		c, _ := gin.CreateTestContext(nil)

		assert.Panics(t, func() {
			MustGetUserID(c)
		})
	})
}
