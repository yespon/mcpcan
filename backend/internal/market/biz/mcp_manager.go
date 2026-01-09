package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kymo-mcp/mcpcan/pkg/llm"
	"github.com/kymo-mcp/mcpcan/pkg/utils"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

type McpManager struct {
	clients map[string]client.MCPClient
}

func NewMcpManager() *McpManager {
	return &McpManager{
		clients: make(map[string]client.MCPClient),
	}
}

func (m *McpManager) Close() {
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

	for name, srv := range config.McpServers {
		var mcpClient client.MCPClient
		var err error

		if srv.URL != "" {
			// Using NewSSEMCPClient for remote connections
			mcpClient, err = client.NewSSEMCPClient(
				srv.URL,
				client.WithHeaders(srv.Headers),
			)
		} else if srv.Command != "" {
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
		} else {
			continue
		}

		if err != nil {
			return fmt.Errorf("failed to create client for %s: %v", name, err)
		}

		_, err = mcpClient.Initialize(ctx, mcp.InitializeRequest{
			Params: mcp.InitializeParams{
				ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
				ClientInfo: mcp.Implementation{
					Name:    "mcpcan-backend",
					Version: "1.0.0",
				},
				Capabilities: mcp.ClientCapabilities{},
			},
		})
		if err != nil {
			return fmt.Errorf("failed to initialize client %s: %v", name, err)
		}

		m.clients[name] = mcpClient
	}
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
