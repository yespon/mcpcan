package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/llm"
	"github.com/kymo-mcp/mcpcan/pkg/utils"
	"github.com/kymo-mcp/mcpcan/pkg/version"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

type McpManager struct {
	clients      map[string]*client.Client
	configs      map[string]utils.McpServerConfig // 存储配置用于重连
	healthStatus map[string]bool                  // 存储健康状态
	mu           sync.RWMutex                     // 并发安全锁
	ctx          context.Context                  // 管理器生命周期上下文
	cancelFunc   context.CancelFunc               // 取消函数
}

func NewMcpManager() *McpManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &McpManager{
		clients:      make(map[string]*client.Client),
		configs:      make(map[string]utils.McpServerConfig),
		healthStatus: make(map[string]bool),
		ctx:          ctx,
		cancelFunc:   cancel,
	}
}

func (m *McpManager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 取消上下文
	if m.cancelFunc != nil {
		m.cancelFunc()
	}

	// 关闭所有客户端
	for _, c := range m.clients {
		c.Close()
	}
}

func (m *McpManager) Initialize(ctx context.Context, configJSON string) error {
	if configJSON == "" || configJSON == "{}" {
		return nil
	}

	var config utils.McpServersConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return fmt.Errorf("invalid mcp config: %v", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	successCount := 0
	for name, srv := range config.McpServers {
		// 存储配置用于后续重连
		m.configs[name] = srv

		// 初始化客户端
		if err := m.initializeClient(ctx, name, srv); err != nil {
			fmt.Printf("[MCP Manager] failed to init client %s: %v\n", name, err)
			// 初始化失败,标记为不健康,但不返回错误
			m.healthStatus[name] = false
			continue
		}

		// 标记为健康
		m.healthStatus[name] = true
		successCount++
	}

	if len(config.McpServers) > 0 && successCount == 0 {
		return fmt.Errorf("all mcp servers failed to initialize")
	}

	return nil
}

// initializeClient 初始化单个 MCP 客户端 (内部方法,调用者需持有锁)
func (m *McpManager) initializeClient(ctx context.Context, name string, srv utils.McpServerConfig) error {
	var mcpClient *client.Client
	var err error

	// Determine transport type
	transportType := srv.Transport
	if transportType == "" {
		transportType = srv.Type
	}
	
	// Auto-detect if not specified
	if transportType == "" {
		if srv.URL != "" {
			if strings.HasSuffix(srv.URL, "/mcp") {
				transportType = model.McpProtocolStreamableHttp.String()
			} else {
				transportType = model.McpProtocolSSE.String()
			}
		} else if srv.Command != "" {
			transportType = model.McpProtocolStdio.String()
		}
	}

	transportType = strings.ToLower(transportType)
	transportType = strings.ReplaceAll(transportType, "_", "-") // Normalize snake_case to kebab-case if needed

	switch transportType {
	case model.McpProtocolSSE.String(), "http": // Standard MCP over HTTP (SSE)
		if srv.URL == "" {
			return fmt.Errorf("transport is %s but url is empty", transportType)
		}
		// Using NewSSEMCPClient for remote connections
		mcpClient, err = client.NewSSEMCPClient(
			srv.URL,
			client.WithHeaders(srv.Headers),
		)
	case model.McpProtocolStreamableHttp.String(): // Streamable HTTP (NDJSON/etc)
		if srv.URL == "" {
			return fmt.Errorf("transport is %s but url is empty", transportType)
		}
		// Using NewStreamableHttpClient
		mcpClient, err = client.NewStreamableHttpClient(
			srv.URL,
			transport.WithHTTPHeaders(srv.Headers),
		)
	case model.McpProtocolStdio.String():
		if srv.Command == "" {
			return fmt.Errorf("transport is stdio but command is empty")
		}
		// Using NewStdioMCPClient for local command execution
		var env []string
		for k, v := range srv.Env {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
		mcpClient, err = client.NewStdioMCPClient(
			srv.Command,
			env,
			srv.Args...,
		)
	default:
		return fmt.Errorf("unsupported transport: %s", transportType)
	}

	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}

	// Start client
	if err := mcpClient.Start(ctx); err != nil {
		return fmt.Errorf("failed to start client: %v", err)
	}

	_, err = mcpClient.Initialize(ctx, mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
			ClientInfo: mcp.Implementation{
				Name:    "mcpcan-backend",
				Version: version.Version,
			},
			Capabilities: mcp.ClientCapabilities{},
		},
	})
	if err != nil {
		mcpClient.Close()
		return fmt.Errorf("failed to initialize client: %v", err)
	}

	m.clients[name] = mcpClient
	return nil
}

func (m *McpManager) GetTools(ctx context.Context) ([]llm.Tool, error) {
	var allTools []llm.Tool
	for name, c := range m.clients {
		resp, err := c.ListTools(ctx, mcp.ListToolsRequest{})
		if err != nil {
			return nil, fmt.Errorf("failed to list tools for %s: %v", name, err)
		}

		// Check ListToolsResult structure. Assuming it has Tools field.
		// The doc says NewListToolsResult(tools []Tool, ...). So Result struct has Tools.
		// resp is *mcp.ListToolsResult
		for _, t := range resp.Tools {
			toolName := fmt.Sprintf("%s__%s", name, t.Name)
			// t.InputSchema is ToolInputSchema which is ToolArgumentsSchema
			// We need to marshal it to json.RawMessage if it isn't already, or pass as is.
			// llm.Tool.Function.Parameters is json.RawMessage (from previous context).
			// mcp.Tool.InputSchema is ToolInputSchema.
			// Let's assume we can marshal it or it's compatible.
			// ToolInputSchema is likely a struct.
			schemaBytes, _ := json.Marshal(t.InputSchema)
			
			llmTool := llm.Tool{
				Type: "function",
				Function: llm.Function{
					Name:        toolName,
					Description: t.Description,
					Parameters:  json.RawMessage(schemaBytes),
				},
			}
			allTools = append(allTools, llmTool)
		}
	}
	return allTools, nil
}

func (m *McpManager) CallTool(ctx context.Context, name string, args string) (*mcp.CallToolResult, error) {
	parts := strings.SplitN(name, "__", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid tool name format: %s", name)
	}
	serverName := parts[0]
	toolName := parts[1]

	c, ok := m.clients[serverName]
	if !ok {
		return nil, fmt.Errorf("server not found: %s", serverName)
	}

	var argsMap map[string]interface{}
	if err := json.Unmarshal([]byte(args), &argsMap); err != nil {
		return nil, fmt.Errorf("invalid arguments json: %v", err)
	}

	return c.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      toolName,
			Arguments: argsMap,
		},
	})
}
