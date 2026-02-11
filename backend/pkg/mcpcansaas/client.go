package mcpcansaas

import (
	"net/http"
	"time"
)

const (
	// Token is the authorization token for MCP SaaS Platform
	Token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjQ5MjQzODAwNDYsImlhdCI6MTc3MDc4MDA0Niwic3ViIjoidGVzdC11c2VyIn0.XLzdmy_2aG9RHW1DLABESnXNhAnWhIt60PREpuNus8Ak5cs4aST6ZGT4fMHboLaux5oZ_1N2h51CIBjd8mfkkeEMj60P-xJGY3ht2HBMlfulB8ilXf_DNy2caxUvlr-r3VcMtDn-RfpTjM9W1Ur0_ope2RzR3GNFijkd2UIKpd1fRLgPASLVnSw0UGgMt6cE7oFoDQYWrTVPE2l1seFOt8S5CDji1DR9AutpEIX6p8gPJh1gonqBaWgzsyq0OuZn7eT1TrMSjo80j7-LtHMj1jkMv2IZAQdhgTUiVZm4B3onEwGSEtg1huBswTMMvrBNVAKr6_QeXZJuIVqMBueYjQ"

	// SaasURL is the base URL for MCP SaaS Platform
	SaasURL = "https://www.mcpcan.com"
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
