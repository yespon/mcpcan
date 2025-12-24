package mcpcansaas

import (
	"context"

	pm "github.com/kymo-mcp/mcpcan/api/market/platform_market"
)

// ListMcpServer calls the remote ListMcpServer API
func (c *Client) ListMcpServer(ctx context.Context, req *pm.ListMcpServerRequest) (*pm.ListMcpServerReply, error) {
	// Create the client
	// Note: Creating a new client for every request is cheap as it reuses the connection
	client := pm.NewMarketClient(c.conn)

	// Add token to context
	ctx = WithToken(ctx)

	// Call the service
	return client.ListMcpServer(ctx, req)
}
