package gotest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAiModel_AllOps(t *testing.T) {
	loadConfig(t)

	var modelID int64 = 27
	modelName := "Integration Model " + time.Now().Format("150405")

	// 1. Create Model
	t.Run("Create Model", func(t *testing.T) {
		fmt.Printf("\n[TEST] Creating AI Model: %s\n", modelName)
		
		req := map[string]interface{}{
			"name":      modelName,
			"provider":  config.Ai.Provider,
			"apiKey":    config.Ai.ApiKey,
			"baseUrl":   config.Ai.BaseUrl,
		}

		resp, body := doRequest(t, "POST", "/market/ai/models", req)
		commonResp := assertResponseSuccess(t, resp, body)

		var data struct {
			Access struct {
				ID int64 `json:"id"`
			} `json:"access"`
		}
		err := json.Unmarshal(commonResp.Data, &data)
		require.NoError(t, err, "Failed to parse create model response")
		
		modelID = data.Access.ID
		require.Greater(t, modelID, int64(0), "Model ID should be positive")
		fmt.Printf("[PASS] ✓ Created Model - ID: %d, Name: %s\n", modelID, modelName)
	})

	// 2. List Models (Verify creation)
	t.Run("List Models", func(t *testing.T) {
		fmt.Printf("[TEST] Listing models to verify creation...\n")
		resp, body := doRequest(t, "GET", "/market/ai/models?page=1&pageSize=100", nil)
		commonResp := assertResponseSuccess(t, resp, body)

		var data struct {
			List []struct {
				ID   int64  `json:"id"`
				Name string `json:"name"`
			} `json:"list"`
			Total int64 `json:"total"`
		}
		err := json.Unmarshal(commonResp.Data, &data)
		require.NoError(t, err, "Failed to parse list models response")

		found := false
		for _, m := range data.List {
			if m.ID == modelID {
				found = true
				break
			}
		}
		require.True(t, found, "Newly created model (ID: %d) should be in the list", modelID)
		fmt.Printf("[PASS] ✓ Found model in list - Total: %d\n", data.Total)
	})

	// 3. Update Model
	t.Run("Update Model", func(t *testing.T) {
		newName := modelName + "_Updated"
		fmt.Printf("[TEST] Updating model name to: %s\n", newName)
		
		req := map[string]interface{}{
			"id":   modelID,
			"name": newName,
		}

		resp, body := doRequest(t, "PUT", "/market/ai/models", req)
		assertResponseSuccess(t, resp, body)
		
		fmt.Printf("[PASS] ✓ Updated Model - New Name: %s\n", newName)
	})

	// 4. Test Connection
	t.Run("Test Connection", func(t *testing.T) {
		fmt.Printf("[TEST] Testing connection to model: %s\n", config.Ai.ModelName)
		
		req := map[string]interface{}{
			"id":        modelID,
			"modelName": config.Ai.ModelName,
		}

		resp, body := doRequest(t, "POST", "/market/ai/models/test", req)
		require.Equal(t, http.StatusOK, resp.StatusCode, "Connection test should return 200")

		// Note: This API returns raw JSON, not wrapped in code/msg/data
		var data struct {
			Success   bool   `json:"success"`
			Message   string `json:"message"`
			LatencyMs int64  `json:"latencyMs"`
		}
		err := json.Unmarshal(body, &data)
		require.NoError(t, err, "Failed to parse connection test response")
		
		require.True(t, data.Success, "Connection test should succeed: %s", data.Message)
		fmt.Printf("[PASS] ✓ Connection Test - Success: %v, Latency: %dms\n", data.Success, data.LatencyMs)
	})

	// 5. Delete Model
	t.Run("Delete Model", func(t *testing.T) {
		fmt.Printf("[TEST] Deleting model ID: %d\n", modelID)
		
		req := map[string]interface{}{"id": modelID}
		resp, body := doRequest(t, "DELETE", fmt.Sprintf("/market/ai/models/%d", modelID), req)
		assertResponseSuccess(t, resp, body)
		
		fmt.Printf("[PASS] ✓ Deleted Model ID: %d\n", modelID)
	})
	
	// 6. Verify Deletion
	t.Run("Verify Deletion", func(t *testing.T) {
		fmt.Printf("[TEST] Verifying model deletion...\n")
		
		resp, body := doRequest(t, "GET", "/market/ai/models?page=1&pageSize=100", nil)
		commonResp := assertResponseSuccess(t, resp, body)

		var data struct {
			List []struct {
				ID int64 `json:"id"`
			} `json:"list"`
		}
		err := json.Unmarshal(commonResp.Data, &data)
		require.NoError(t, err, "Failed to parse list response")

		for _, m := range data.List {
			if m.ID == modelID {
				t.Fatalf("Model ID %d should have been deleted but still exists", modelID)
			}
		}
		
		fmt.Printf("[PASS] ✓ Verified deletion - Model ID %d not found in list\n", modelID)
	})
}
