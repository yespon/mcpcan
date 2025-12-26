package mcpcansaas

import (
	"net/http"
	"time"
)

const (
	// Token is the authorization token for MCP SaaS Platform
	Token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjQ5MjAxNjM2MjgsImlhdCI6MTc2NjU2MzYyOCwic3ViIjoidGVzdC11c2VyIn0.GeoPlkvHwvcZC1HP3PsrefK8dY5W0eEgRH7wcbv49DEsILWVhaRtwQiTI8nsOSADKZ-dAZgNQIkEvGhkBJ1iVyeA_-96VF0SRqyvp2wLxTM7sLdgzJsaOqpvp1NdBWAETIjqYKuZhmbKu09LMDyDqEeKkRgNvsShJbzVxGNMHAg7-x5DurTjy4F5uW5tfjSgmhQVL90rRWuTnnlXvJDUF3B_glsaOyihPS7TYqg0GL-tdwwL8c1eFv-E9rzu0xb3DgxyN7MhaIE8xO0lnOefP-JUxCE1gsV5A9idWvBZB9Cv76UAWIPI8aCAqdyhUUxzhFbvzkQzlKqCrXtXivOWyQ"

	// SaasURL is the base URL for MCP SaaS Platform
	// SaasURL = "https://www.mcpcan.com"
	SaasURL = "http://localhost:8100"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new MCP SaaS client
func NewClient() (*Client, error) {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: SaasURL,
	}, nil
}

func (c *Client) Close() error {
	return nil
}
