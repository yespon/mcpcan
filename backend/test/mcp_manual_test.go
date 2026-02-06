package test

import (
	"context"
	"testing"
	"time"

	"github.com/kymo-mcp/mcpcan/internal/market/biz"
)

func TestMcpConnection(t *testing.T) {
	configJSON := `{
  "mcpServers": {
    "mcp-157b1866": {
      "url": "https://mcp-dev.itqm.com/mcp-gateway/157b1866-dd5e-4870-8abe-004b74c9ee02/mcp",
      "headers": {
        "Authorization": "Bearer NmRjOTg5M2ItZDEyYy00NjcwLWI5ZTAtOWZkNWU2OTY5NjJmeyJleHBpcmVBdCI6MTc3MDM0NTAzODI1NSwidXNlcklkIjoxLCJ1c2VybmFtZSI6ImFkbWluIn0="
      }
    }
  }
}`

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	manager := biz.NewMcpManager()
	defer manager.Close()

	t.Logf("Initializing MCP Manager...")
	err := manager.Initialize(ctx, configJSON)
	if err != nil {
		t.Fatalf("Failed to initialize MCP Manager: %v", err)
	}

	t.Logf("Getting tools...")
	tools, err := manager.GetTools(ctx)
	if err != nil {
		t.Fatalf("Failed to get tools: %v", err)
	}

	t.Logf("Found %d tools", len(tools))
	for _, tool := range tools {
		t.Logf("- Name: %s", tool.Function.Name)
		t.Logf("  Description: %s", tool.Function.Description)
	}
}
