package utils

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/kymo-mcp/mcpcan/api/authz/user_auth"
	"github.com/kymo-mcp/mcpcan/pkg/common"
)

func GetCurrentUser(c *gin.Context) (*user_auth.UserInfo, error) {
	userInfoAny, exists := c.Get(common.UserInfoContextKey)
	if !exists {
		return nil, errors.New("user not found in context")
	}
	userInfo, ok := userInfoAny.(*user_auth.UserInfo)
	if !ok {
		return nil, errors.New("user info type error")
	}
	return userInfo, nil
}
