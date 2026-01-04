package mcpcansaas_test

import (
	"context"
	"testing"
	"time"

	pm "github.com/kymo-mcp/mcpcan/api/market/platform_market"
	"github.com/kymo-mcp/mcpcan/pkg/mcpcansaas"
)

func TestListMcpServer(t *testing.T) {
	// Create a new client
	client, err := mcpcansaas.NewClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	t.Log("Connected to MCP SaaS Market")

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Prepare the request
	req := &pm.ListMcpServerRequest{
		Page:     1,
		PageSize: 10,
		Name:     "MarkMap",
	}

	t.Logf("Sending ListMcpServer request: page=%d, pageSize=%d\n", req.Page, req.PageSize)

	// Call the API
	reply, err := client.ListMcpServer(ctx, req)
	if err != nil {
		t.Fatalf("Failed to call ListMcpServer: %v", err)
	}

	t.Logf("Received reply: Total=%v\n", reply.Total)

	// Pretty print the list
	if len(reply.List) > 0 {
		t.Log("MCP Servers (showing first 1):")
		for i, server := range reply.List {
			if i >= 1 {
				break
			}
			t.Logf("- [%v] %s\n", server.Id, server.Name)
		}
	} else {
		t.Log("No servers found.")
	}
}
