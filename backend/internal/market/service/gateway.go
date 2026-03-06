package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
	"github.com/kymo-mcp/mcpcan/pkg/logger"
	"github.com/kymo-mcp/mcpcan/pkg/redis"
	"go.uber.org/zap"
)

type GatewayService struct{}

func NewGatewayService() *GatewayService {
	return &GatewayService{}
}

// AuthHandler 处理 Traefik ForwardAuth 请求
func (s *GatewayService) AuthHandler(c *gin.Context) {
	// 1. 获取 Traefik 透传的原始请求 URI
	forwardedURI := c.GetHeader("X-Forwarded-Uri")
	if forwardedURI == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing X-Forwarded-Uri"})
		return
	}

	// 2. 提取 instanceID
	// 通常 forwardedURI = /mcp-gateway/<instance_id>/...
	prefix := common.GetGatewayRoutePrefix()
	prefix = strings.Trim(prefix, "/")
	if prefix == "" {
		prefix = "mcp-gateway"
	}

	parts := strings.Split(forwardedURI, "/")
	// /mcp-gateway/xxx/ -> parts: ["", "mcp-gateway", "xxx", ...]
	if len(parts) < 3 || parts[1] != prefix {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid path format"})
		return
	}
	instanceID := parts[2]
	if instanceID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "empty instance id"})
		return
	}

	// 3. 读取实例信息
	instanceInfo, err := mysql.McpInstanceRepo.FindByInstanceID(context.Background(), instanceID)
	if err != nil || instanceInfo == nil {
		logger.Warn("Instance missing or error during gateway auth", zap.String("instance_id", instanceID), zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "instance not found"})
		return
	}

	// 4. 校验 Token
	if instanceInfo.EnabledToken {
		mcpToken, errToken := validateMcpTokenForInstance(c, instanceID)
		if errToken != nil {
			logger.Warn("Gateway token validation failed", zap.String("instance_id", instanceID), zap.Error(errToken))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token validation failed"})
			return
		}
		
		// 将 Token 中自定义的外部请求 Header 发载入 Auth response header
		// 这样经过配置了 authResponseHeadersRegex 的 Traefik Middleware 后就会被全数添加回发给上游服务的 Request Headers 中。
		if mcpToken != nil && len(mcpToken.Headers) > 0 {
			var extraHeaders map[string]string
			if err := json.Unmarshal(mcpToken.Headers, &extraHeaders); err == nil {
				for k, v := range extraHeaders {
					c.Header(k, v)
				}
			} else {
				logger.Warn("Failed to unmarshal mcp token extra headers", zap.Error(err))
			}
		}
	}

	// 5. 将一些信息通过 Header 的方式传递给 Traefik，Traefik 会自动追加到目标后端请求中
	c.Header("X-Mcp-Instance-Id", instanceID)
	
	// 在此处如果未来需要加限流、防重放攻击，也是在 c.JSON / c.Status 返回前添加逻辑即可
	c.Status(http.StatusOK)
}

// 抽取自老网关的 Token 校验逻辑
func validateMcpTokenForInstance(c *gin.Context, instanceID string) (*model.McpToken, error) {
	req := c.Request
	token := req.Header.Get("Authorization")
	if token == "" {
		token = req.Header.Get("API-Key")
	}
	if token == "" {
		token = req.Header.Get("X-API-Key")
	}
	if token == "" {
		token = c.Query("token") // 有些 SSE 流喜欢把 token 放在 query 里面
	}
	
	// 预先准备好带有 Bearer 和不带 Bearer 的两种形式，兼容数据库不同的存储历史
	rawToken := token
	strippedToken := token
	if strings.HasPrefix(strings.ToLower(token), "bearer ") {
		strippedToken = strings.TrimSpace(token[7:])
	}
	bearerToken := "Bearer " + strippedToken

	if strings.TrimSpace(strippedToken) == "" {
		return nil, fmt.Errorf("token header missing or empty")
	}

	tokenCache := redis.GetMcpTokenCache()
	var mcpToken *model.McpToken

	// 优先查询原始 token 格式，没找到再查替代格式
	searchTokens := []string{rawToken}
	if rawToken == strippedToken {
		searchTokens = append(searchTokens, bearerToken)
	} else {
		searchTokens = append(searchTokens, strippedToken)
	}

	for _, t := range searchTokens {
		cacheKey := tokenCache.GenerateCacheKey(instanceID, t)
		if v := tokenCache.GetRedis(cacheKey); v != nil {
			mcpToken = v
			break
		}
		
		trow, err := mysql.McpTokenRepo.FindByToken(context.Background(), instanceID, t)
		if err == nil && trow != nil && trow.InstanceID == instanceID {
			_ = tokenCache.SetRedis(cacheKey, trow, redis.TokenCacheExpire)
			mcpToken = trow
			break
		}
	}

	if mcpToken == nil {
		return nil, fmt.Errorf("not found")
	}

	if !mcpToken.Enabled {
		return nil, fmt.Errorf("disabled token")
	}

	return mcpToken, nil
}

// 供 strings.Trim 等函数容错使用
func trimStr(s string) string {
	return strings.TrimSpace(s)
}
