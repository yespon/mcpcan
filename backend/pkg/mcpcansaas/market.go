package mcpcansaas

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	pm "github.com/kymo-mcp/mcpcan/api/market/platform_market"
)

// ListMcpServer calls the remote ListMcpServer API
func (c *Client) ListMcpServer(ctx context.Context, req *pm.ListMcpServerRequest) (*pm.ListMcpServerReply, error) {
	// Construct URL with query parameters
	u, err := url.Parse(c.baseURL + "/market/list")
	if err != nil {
		return nil, fmt.Errorf("failed to parse url: %v", err)
	}
	q := u.Query()
	q.Set("page", fmt.Sprintf("%d", req.Page))
	q.Set("pageSize", fmt.Sprintf("%d", req.PageSize))
	if req.Name != "" {
		q.Set("name", req.Name)
	}
	if req.CategoryName != "" {
		q.Set("categoryName", req.CategoryName)
	}
	u.RawQuery = q.Encode()

	httpReq, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+Token)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var reply pm.ListMcpServerReply
	if err := json.Unmarshal(body, &reply); err != nil {
		// Try wrapped response
		var wrapped struct {
			Code    int                    `json:"code"`
			Data    *pm.ListMcpServerReply `json:"data"`
			Message string                 `json:"message"`
		}
		if err2 := json.Unmarshal(body, &wrapped); err2 == nil && wrapped.Data != nil {
			return wrapped.Data, nil
		}
		return nil, fmt.Errorf("failed to unmarshal response: %v, body: %s", err, string(body))
	}

	return &reply, nil
}
