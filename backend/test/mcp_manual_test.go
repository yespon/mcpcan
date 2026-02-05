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
    "mcp-9469c672": {
      "url": "http://localhost/mcp-gateway/9469c672-7070-4cc5-b4f8-604f916b4784/mcp",
      "headers": {
        "Authorization": "Bearer ZDYzZDhkOTItMmFhNy00ODNiLWJjZTUtNmQ2ZTU5NWJkNTYzeyJleHBpcmVBdCI6MTc2ODg5MzM0MzYwOSwidXNlcklkIjoxLCJ1c2VybmFtZSI6ImFkbWluIn0=",
        "Accept": "text/event-stream"
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
