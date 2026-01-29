package gotest

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAiModel_GetSupportedModels tests GET /ai/models/supported
func TestAiModel_GetSupportedModels(t *testing.T) {
	loadConfig(t)

	fmt.Printf("\n[TEST] Getting supported models and providers...\n")

	resp, body := doRequest(t, "GET", "/market/ai/models/supported", nil)
	commonResp := assertResponseSuccess(t, resp, body)

	var data struct {
		Providers []struct {
			Id          string   `json:"id"`
			Name        string   `json:"name"`
			Models      []string `json:"models"`
			RegisterUrl string   `json:"registerUrl"`
			DocsUrl     string   `json:"docsUrl"`
			BaseUrl     string   `json:"baseUrl"`
		} `json:"providers"`
	}
	err := json.Unmarshal(commonResp.Data, &data)
	require.NoError(t, err, "Failed to parse supported models response")

	require.NotEmpty(t, data.Providers, "Should have at least one provider")

	fmt.Printf("[PASS] ✓ Found %d providers\n", len(data.Providers))
	for _, p := range data.Providers {
		fmt.Printf("  - %s (%s): %d models\n", p.Name, p.Id, len(p.Models))
	}
}

// TestAiModel_GetAvailableModels tests GET /ai/models/available
func TestAiModel_GetAvailableModels(t *testing.T) {
	loadConfig(t)

	fmt.Printf("\n[TEST] Getting available (user-configured) models...\n")

	resp, body := doRequest(t, "GET", "/market/ai/models/available", nil)
	commonResp := assertResponseSuccess(t, resp, body)

	var data struct {
		List []struct {
			Id       int64  `json:"id"`
			Name     string `json:"name"`
			Provider string `json:"provider"`
		} `json:"list"`
		Total int64 `json:"total"`
	}
	err := json.Unmarshal(commonResp.Data, &data)
	require.NoError(t, err, "Failed to parse available models response")

	fmt.Printf("[PASS] ✓ Found %d user-available models\n", len(data.List))
	for _, m := range data.List {
		fmt.Printf("  - [%d] %s (%s)\n", m.Id, m.Name, m.Provider)
	}
}

// TestAiModel_TestConnectionDirect tests POST /ai/models/test with direct credentials
func TestAiModel_TestConnectionDirect(t *testing.T) {
	loadConfig(t)

	fmt.Printf("\n[TEST] Testing connection with direct credentials...\n")

	req := map[string]interface{}{
		"provider":  config.Ai.Provider,
		"apiKey":    config.Ai.ApiKey,
		"baseUrl":   config.Ai.BaseUrl,
		"modelName": config.Ai.ModelName,
	}

	resp, body := doRequest(t, "POST", "/market/ai/models/test", req)
	require.Equal(t, 200, resp.StatusCode, "Should return 200")

	var data struct {
		Success   bool   `json:"success"`
		Message   string `json:"message"`
		LatencyMs int64  `json:"latencyMs"`
	}
	err := json.Unmarshal(body, &data)
	require.NoError(t, err, "Failed to parse test connection response")

	assert.True(t, data.Success, "Direct connection test should succeed: %s", data.Message)
	fmt.Printf("[PASS] ✓ Direct Connection Test - Success: %v, Latency: %dms\n", data.Success, data.LatencyMs)
}
