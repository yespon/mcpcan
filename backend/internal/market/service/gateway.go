package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
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
	// 1. 获取 Traefik 透传的原始请求信息
	forwardedMethod := c.GetHeader("X-Forwarded-Method")
	forwardedURI := c.GetHeader("X-Forwarded-Uri")
	forwardedIP := c.GetHeader("X-Forwarded-For")

	if forwardedURI == "" {
		logger.Warn("Gateway auth failed: missing X-Forwarded-Uri")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing X-Forwarded-Uri"})
		return
	}

	// 2. 提取 instanceID 和 toolName
	prefix := common.GetGatewayRoutePrefix()
	prefix = strings.Trim(prefix, "/")
	if prefix == "" {
		prefix = "mcp-gateway"
	}

	parts := strings.Split(forwardedURI, "/")
	// /mcp-gateway/xxx/ -> parts: ["", "mcp-gateway", "xxx", ...]
	if len(parts) < 3 || parts[1] != prefix {
		logger.Warn("Gateway auth failed: invalid path format",
			zap.String("uri", forwardedURI),
			zap.String("method", forwardedMethod),
			zap.String("ip", forwardedIP),
		)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid path format"})
		return
	}
	instanceID := parts[2]
	if instanceID == "" {
		logger.Warn("Gateway auth failed: empty instance id",
			zap.String("uri", forwardedURI),
			zap.String("method", forwardedMethod),
		)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "empty instance id"})
		return
	}

	// 尝试提取 toolName
	// /mcp-gateway/<instance_id>/tools/<name> -> parts[4]
	// /mcp-gateway/<instance_id>/<action> -> parts[3]
	toolName := ""
	if len(parts) >= 5 && parts[3] == "tools" {
		toolName = parts[4]
	} else if len(parts) >= 4 {
		toolName = parts[3]
	}

	logFields := []zap.Field{
		zap.String("instance_id", instanceID),
		zap.String("tool_name", toolName),
		zap.String("method", forwardedMethod),
		zap.String("uri", forwardedURI),
		zap.String("ip", forwardedIP),
	}

	logger.Info("Gateway auth incoming", logFields...)

	// 3. 读取实例信息
	instanceInfo, err := mysql.McpInstanceRepo.FindByInstanceID(context.Background(), instanceID)
	if err != nil || instanceInfo == nil {
		logger.Warn("Gateway auth failed: instance not found", append(logFields, zap.Error(err))...)
		
		// 记录失败日志到数据库
		go func() {
			logData := map[string]interface{}{
				"method": forwardedMethod,
				"uri":    forwardedURI,
				"ip":     forwardedIP,
				"error":  "instance not found",
			}
			logRaw, _ := json.Marshal(logData)
			_ = mysql.GatewayLogRepo.Create(context.Background(), &model.GatewayLog{
				TraceID:    c.GetString("trace_id"),
				InstanceID: instanceID,
				ToolName:   toolName,
				Level:      5, // Error
				Event:      model.EventAuthFailed,
				Log:        json.RawMessage(logRaw),
			})
		}()

		c.JSON(http.StatusUnauthorized, gin.H{"error": "instance not found"})
		return
	}

	// 4.1 收集各层级 Header，实现优先级逻辑：自定义透传 Header > MCP 配置 Header > 客户端请求 Header
	// 注：客户端请求 Header 由 Traefik 默认透传，如果此处设置同名 Header，则会覆盖客户端的值
	finalHeaders := make(map[string]string)

	// 获取 MCP 配置中的 Header
	var mcpConfig *model.McpConfig
	switch instanceInfo.AccessType {
	case model.AccessTypeProxy:
		_, _, mcpConfig, _ = instanceInfo.GetSourceConfig()
	case model.AccessTypeHosting:
		mcpConfig = &model.McpConfig{
			Type:      instanceInfo.ProxyProtocol.String(),
			Transport: instanceInfo.ProxyProtocol.String(),
			URL:       instanceInfo.ContainerServiceURL,
		}
	}

	// 1. 先加载 MCP 服务器配置中的 Header
	if mcpConfig != nil && len(mcpConfig.Headers) > 0 {
		for k, v := range mcpConfig.Headers {
			finalHeaders[k] = v
		}
	}

	token := ""
	if auth := c.GetHeader("Authorization"); auth != "" {
		token = auth
	} else if key := c.GetHeader("API-Key"); key != "" {
		token = key
	} else if xkey := c.GetHeader("X-API-Key"); xkey != "" {
		token = xkey
	} else {
		token = c.Query("token")
	}

	// 4.2 校验 Token
	if instanceInfo.EnabledToken {
		if _, errToken := validateMcpTokenForInstance(c, instanceID); errToken != nil {
			logger.Warn("Gateway auth failed: token validation failed", append(logFields, zap.Error(errToken))...)
			
			// 记录失败日志到数据库
			go func(t string) {
				logData := map[string]interface{}{
					"method": forwardedMethod,
					"uri":    forwardedURI,
					"ip":     forwardedIP,
					"token":  t,
					"error":  errToken.Error(),
				}
				logRaw, _ := json.Marshal(logData)
				_ = mysql.GatewayLogRepo.Create(context.Background(), &model.GatewayLog{
					TraceID:    c.GetString("trace_id"),
					InstanceID: instanceID,
					ToolName:   toolName,
					Token:      t,
					Level:      5, // Error
					Event:      model.EventAuthFailed,
					Log:        json.RawMessage(logRaw),
				})
			}(token)

			c.JSON(http.StatusUnauthorized, gin.H{"error": "token validation failed"})
			return
		}
	}

	// 4.3 加载实例级自定义透传 Header (优先级最高，覆盖 MCP 配置 Header)
	if len(instanceInfo.Headers) > 0 {
		var extraHeaders map[string]string
		if err := json.Unmarshal(instanceInfo.Headers, &extraHeaders); err == nil {
			for k, v := range extraHeaders {
				finalHeaders[k] = v
			}
		} else {
			logger.Warn("Gateway auth: failed to unmarshal instance extra headers", append(logFields, zap.Error(err))...)
		}
	}
	// 5. 将最终确定的 Header 注入 Auth response，由 Traefik 透传给 Sidecar/Upstream
	for k, v := range finalHeaders {
		c.Header(k, v)
	}

	c.Header("X-Mcp-Instance-Id", instanceID)

	logger.Info("Gateway auth success", logFields...)
	c.Status(http.StatusOK)

	// 记录成功日志到数据库
	go func(t string) {
		logData := map[string]interface{}{
			"method": forwardedMethod,
			"uri":    forwardedURI,
			"ip":     forwardedIP,
		}
		logRaw, _ := json.Marshal(logData)
		_ = mysql.GatewayLogRepo.Create(context.Background(), &model.GatewayLog{
			TraceID:    c.GetString("trace_id"),
			InstanceID: instanceID,
			ToolName:   toolName,
			Token:      t,
			Level:      3, // Info
			Event:      model.EventAuthSuccess,
			Log:        json.RawMessage(logRaw),
		})
	}(token)
}

// 抽取自老网关的 Token 校验逻辑
func validateMcpTokenForInstance(c *gin.Context, instanceID string) (*model.McpToken, error) {
	req := c.Request
	token := req.Header.Get("Authorization")
	if token == "" {
		token = req.Header.Get("X-Mcp-Authorization")
	}
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


// ProxyHandler handles direct MCP gateway requests, performing authentication and reverse proxying to sidecars.
func (s *GatewayService) ProxyHandler(c *gin.Context) {
	instanceID := c.Param("instanceID")
	if instanceID == "" {
		// Try to extract from path if param is not set (generic Any route)
		uri := c.Request.URL.Path
		parts := strings.Split(strings.Trim(uri, "/"), "/")
		if len(parts) >= 2 {
			instanceID = parts[1]
		}
	}

	if instanceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing instance id"})
		return
	}

	// 1. Get Instance Info
	instance, err := mysql.McpInstanceRepo.FindByInstanceID(context.Background(), instanceID)
	if err != nil || instance == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "instance not found"})
		return
	}

	// 2. Perform Internal Authentication
	if instance.EnabledToken {
		if _, err = validateMcpTokenForInstance(c, instanceID); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication failed: " + err.Error()})
			return
		}
	}

	// 3. Prepare Target URL
	targetBase := instance.ContainerServiceURL
	if targetBase == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "instance service URL not configured"})
		return
	}

	targetURL, err := url.Parse(targetBase)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse target URL"})
		return
	}

	// 4. Setup Reverse Proxy
	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = targetURL.Scheme
			req.URL.Host = targetURL.Host

			// 保留完整原始路径（不 strip instanceID 前缀）
			// sidecar 内部配置了 MCP_ROUTE_PREFIX，会自行 strip /mcp-gateway/<instanceID> 前缀
			// 与 Traefik → sidecar 的路由行为完全一致
			req.URL.Path = c.Request.URL.Path
			if c.Request.URL.RawPath != "" {
				req.URL.RawPath = c.Request.URL.RawPath
			}

			req.URL.RawQuery = c.Request.URL.RawQuery
			if targetURL.RawQuery != "" {
				if req.URL.RawQuery == "" {
					req.URL.RawQuery = targetURL.RawQuery
				} else {
					req.URL.RawQuery = req.URL.RawQuery + "&" + targetURL.RawQuery
				}
			}

			// ① 先剥离客户端原始鉴权 headers，防止 MCP Token 透传到 sidecar/upstream
			// 注：实例配置的 headers 会在下方按优先级注入覆盖
			for _, sensitiveKey := range []string{"Authorization", "API-Key", "X-API-Key"} {
				req.Header.Del(sensitiveKey)
			}

			// ② 按优先级注入 headers：MCP config headers（低）→ 实例 headers（高，覆盖）
			injectedKeys := make([]string, 0)

			// 低优先级：MCP config headers（仅 proxy 类型实例有）
			var mcpConfig *model.McpConfig
			if instance.AccessType == model.AccessTypeProxy {
				_, _, mcpConfig, _ = instance.GetSourceConfig()
			}
			if mcpConfig != nil && len(mcpConfig.Headers) > 0 {
				for k, v := range mcpConfig.Headers {
					req.Header.Set(k, v)
					injectedKeys = append(injectedKeys, k)
				}
			}

			// 高优先级：实例 headers（创建/编辑时配置，覆盖 MCP config headers）
			if len(instance.Headers) > 0 {
				var extraHeaders map[string]string
				if err := json.Unmarshal(instance.Headers, &extraHeaders); err == nil {
					for k, v := range extraHeaders {
						req.Header.Set(k, v)
						injectedKeys = append(injectedKeys, k)
					}
				}
			}

			req.Header.Set("X-Mcp-Instance-Id", instanceID)
			req.Header.Set("X-Internal-Request", "true")
			req.Host = targetURL.Host

			logger.Info("[ProxyHandler] forwarding",
				zap.String("instance", instanceID),
				zap.String("target", req.URL.String()),
				zap.Strings("injected_headers", injectedKeys),
			)
		},
		ModifyResponse: func(resp *http.Response) error {
			logger.Info("[ProxyHandler] upstream response",
				zap.String("instance", instanceID),
				zap.String("target", targetBase),
				zap.Int("status", resp.StatusCode),
			)
			return nil
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			logger.Error("[ProxyHandler] upstream error",
				zap.String("instance", instanceID),
				zap.String("target", targetBase),
				zap.Error(err),
			)
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte(`{"error":"upstream unavailable"}`))
		},
		FlushInterval: -1, // SSE streaming 必须
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}

// 供 strings.Trim 等函数容错使用
func trimStr(s string) string {
	return strings.TrimSpace(s)
}
