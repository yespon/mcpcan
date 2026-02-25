package common

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// GetUserIDFromContext 从 gin.Context 中提取用户 ID
// 该函数从认证中间件设置的上下文中获取当前用户的 ID
// 返回值:
//   - int64: 用户 ID
//   - error: 如果用户未认证或 ID 格式错误,返回错误
func GetUserIDFromContext(c *gin.Context) (int64, error) {
	userID, exists := c.Get("userId")
	if !exists {
		return 0, fmt.Errorf("user id not found in context")
	}

	// 类型断言,支持多种数值类型
	switch v := userID.(type) {
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	case uint:
		return int64(v), nil
	case float64:
		return int64(v), nil
	default:
		return 0, fmt.Errorf("invalid user id type: %T", userID)
	}
}

// GetUsernameFromContext 从 gin.Context 中提取用户名
// 该函数从认证中间件设置的上下文中获取当前用户的用户名
// 返回值:
//   - string: 用户名
//   - error: 如果用户未认证或用户名格式错误,返回错误
func GetUsernameFromContext(c *gin.Context) (string, error) {
	username, exists := c.Get("username")
	if !exists {
		return "", fmt.Errorf("username not found in context")
	}

	if str, ok := username.(string); ok {
		return str, nil
	}
	return "", fmt.Errorf("invalid username type: %T", username)
}

// MustGetUserID 从 Context 中提取用户 ID,如果失败则 panic
// 仅在确定用户已认证的场景下使用(如已通过认证中间件的路由)
func MustGetUserID(c *gin.Context) int64 {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		panic(fmt.Sprintf("failed to get user id: %v", err))
	}
	return userID
}
