package service

import (
	"context"
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
		errToken := validateMcpTokenForInstance(c, instanceID)
		if errToken != nil {
			logger.Warn("Gateway token validation failed", zap.String("instance_id", instanceID), zap.Error(errToken))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token validation failed"})
			return
		}
	}

	// 5. 将一些信息通过 Header 的方式传递给 Traefik，Traefik 会自动追加到目标后端请求中
	c.Header("X-Mcp-Instance-Id", instanceID)
	
	// 在此处如果未来需要加限流、防重放攻击，也是在 c.JSON / c.Status 返回前添加逻辑即可
	c.Status(http.StatusOK)
}

// RoutesHandler 给 Traefik HTTP Provider 提供 "直连 MCP" (无需 docker) 的动态路由列表
func (s *GatewayService) RoutesHandler(c *gin.Context) {
	// 查询活跃的实例列表
	instances, err := mysql.McpInstanceRepo.FindByStatus(context.Background(), model.InstanceStatusActive)
	if err != nil {
		logger.Error("Failed to fetch instances for traefik routes", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	// 构造 Traefik HTTP Provider 格式
	type Service struct {
		LoadBalancer struct {
			Servers []struct {
				URL string `json:"url"`
			} `json:"servers"`
		} `json:"loadBalancer"`
	}
	type StripPrefix struct {
		Prefixes []string `json:"prefixes"`
	}
	type Middleware struct {
		StripPrefix *StripPrefix `json:"stripPrefix,omitempty"`
	}

	type Router struct {
		Rule        string   `json:"rule"`
		Service     string   `json:"service"`
		Middlewares []string `json:"middlewares"`
	}

	response := struct {
		HTTP struct {
			Routers     map[string]Router     `json:"routers"`
			Services    map[string]Service    `json:"services"`
			Middlewares map[string]Middleware `json:"middlewares"`
		} `json:"http"`
	}{}
	response.HTTP.Routers = make(map[string]Router)
	response.HTTP.Services = make(map[string]Service)
	response.HTTP.Middlewares = make(map[string]Middleware)

	prefix := common.GetGatewayRoutePrefix()
	prefix = strings.Trim(prefix, "/")

	for _, instance := range instances {
		// Hosting 类实例由 Docker Label 或 K8s Ingress 自动发现
		// 此处只为 Direct/Proxy 类型生成 HTTP Provider 路由
		// 如果实例有关联容器（如翻译器 Sidecar），则通过 Docker Label/K8s Ingress 自动发现，此处跳过
		if instance.ContainerName != "" {
			continue
		}
		if instance.ContainerServiceURL == "" {
			continue
		}

		routerName := fmt.Sprintf("mcp-ext-inst-%s", instance.InstanceID)
		serviceName := fmt.Sprintf("mcp-ext-svc-%s", instance.InstanceID)
		stripMidName := fmt.Sprintf("mcp-ext-strip-%s", instance.InstanceID)

		// 路由 (将 /mcp-gateway/xxx/ 前缀捕获过来)
		response.HTTP.Routers[routerName] = Router{
			Rule:        fmt.Sprintf("PathPrefix(`/%s/%s/`)", prefix, instance.InstanceID),
			Service:     serviceName,
			Middlewares: []string{stripMidName, "mcp-auth@file"},
		}

		// 后端服务
		srv := Service{}
		srv.LoadBalancer.Servers = []struct {
			URL string `json:"url"`
		}{{URL: instance.ContainerServiceURL}}
		response.HTTP.Services[serviceName] = srv

		// 剥离前缀的中间件
		mid := Middleware{
			StripPrefix: &StripPrefix{
				Prefixes: []string{fmt.Sprintf("/%s/%s/", prefix, instance.InstanceID)},
			},
		}
		response.HTTP.Middlewares[stripMidName] = mid
	}

	c.JSON(http.StatusOK, response)
}

// 抽取自老网关的 Token 校验逻辑
func validateMcpTokenForInstance(c *gin.Context, instanceID string) error {
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
	
	// 如果是 Bearer xxx 格式，剥离 Bearer
	if strings.HasPrefix(strings.ToLower(token), "bearer ") {
		token = token[7:]
	}

	if strings.TrimSpace(token) == "" {
		return fmt.Errorf("token header missing or empty")
	}

	tokenCache := redis.GetMcpTokenCache()
	cacheKey := tokenCache.GenerateCacheKey(instanceID, token)
	var mcpToken *model.McpToken

	if v := tokenCache.GetRedis(cacheKey); v != nil {
		mcpToken = v
	} else {
		trow, err := mysql.McpTokenRepo.FindByToken(context.Background(), instanceID, token)
		if err != nil || trow == nil || trow.InstanceID != instanceID {
			return fmt.Errorf("not found")
		}
		_ = tokenCache.SetRedis(cacheKey, trow, redis.TokenCacheExpire)
		mcpToken = trow
	}

	if !mcpToken.Enabled {
		return fmt.Errorf("disabled token")
	}

	return nil
}

// 供 strings.Trim 等函数容错使用
func trimStr(s string) string {
	return strings.TrimSpace(s)
}
