package mcpcansaas

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

const (
	// Token is the authorization token for MCP SaaS Platform
	Token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjQ5MjAxNjM2MjgsImlhdCI6MTc2NjU2MzYyOCwic3ViIjoidGVzdC11c2VyIn0.GeoPlkvHwvcZC1HP3PsrefK8dY5W0eEgRH7wcbv49DEsILWVhaRtwQiTI8nsOSADKZ-dAZgNQIkEvGhkBJ1iVyeA_-96VF0SRqyvp2wLxTM7sLdgzJsaOqpvp1NdBWAETIjqYKuZhmbKu09LMDyDqEeKkRgNvsShJbzVxGNMHAg7-x5DurTjy4F5uW5tfjSgmhQVL90rRWuTnnlXvJDUF3B_glsaOyihPS7TYqg0GL-tdwwL8c1eFv-E9rzu0xb3DgxyN7MhaIE8xO0lnOefP-JUxCE1gsV5A9idWvBZB9Cv76UAWIPI8aCAqdyhUUxzhFbvzkQzlKqCrXtXivOWyQ"

	// debug localhost:9100
	// saas www.mcpcan.com:443
	SaasHost = "www.mcpcan.com:443"
)

type Client struct {
	conn *grpc.ClientConn
}

// NewClient creates a new MCP SaaS client
func NewClient() (*Client, error) {
	var opts []grpc.DialOption
	target := SaasHost

	if strings.Contains(target, "localhost") || strings.Contains(target, "127.0.0.1") {
		// Use insecure for local development
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		// Use TLS for production
		creds := credentials.NewTLS(&tls.Config{})
		opts = append(opts, grpc.WithTransportCredentials(creds))
	}

	conn, err := grpc.NewClient(target, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mcp saas: %v", err)
	}

	return &Client{conn: conn}, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) GetConn() *grpc.ClientConn {
	return c.conn
}

// WithToken adds the authorization token to the context
func WithToken(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+Token)
}
