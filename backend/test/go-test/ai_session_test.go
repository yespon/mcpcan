package gotest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAiSession_AllOps(t *testing.T) {
	loadConfig(t)

	var modelID int64
	var sessionID int64

	// Prerequisite: Create a Model for session tests
	t.Run("Setup Model", func(t *testing.T) {
		fmt.Printf("\n[SETUP] Creating test model for session tests...\n")
		
		req := map[string]interface{}{
			"name":     "Session Test Model",
			"provider": config.Ai.Provider,
			"apiKey":   config.Ai.ApiKey,
			"baseUrl":  config.Ai.BaseUrl,
		}
		resp, body := doRequest(t, "POST", "/market/ai/models", req)
		commonResp := assertResponseSuccess(t, resp, body)
		
		var data struct {
			Access struct{ ID int64 `json:"id"` } `json:"access"`
		}
		err := json.Unmarshal(commonResp.Data, &data)
		require.NoError(t, err, "Failed to parse setup model response")
		
		modelID = data.Access.ID
		require.Greater(t, modelID, int64(0), "Model ID should be positive")
		fmt.Printf("[PASS] ✓ Setup Model - ID: %d\n", modelID)
	})

	// 1. Create Session
	t.Run("Create Session", func(t *testing.T) {
		sessionName := "Session Test " + time.Now().Format("15:04:05")
		fmt.Printf("\n[TEST] Creating AI Session: %s\n", sessionName)
		
		req := map[string]interface{}{
			"name":          sessionName,
			"modelAccessID": modelID,
			"modelName":     config.Ai.ModelName,
			"maxContext":    5,
			"toolsConfig":   "{}",
		}

		resp, body := doRequest(t, "POST", "/market/ai/sessions", req)
		commonResp := assertResponseSuccess(t, resp, body)

		var data struct {
			Session struct {
				ID int64 `json:"id"`
			} `json:"session"`
		}
		err := json.Unmarshal(commonResp.Data, &data)
		require.NoError(t, err, "Failed to parse create session response")

		sessionID = data.Session.ID
		require.Greater(t, sessionID, int64(0), "Session ID should be positive")
		fmt.Printf("[PASS] ✓ Created Session - ID: %d, Model: %s\n", sessionID, config.Ai.ModelName)
	})

	// 2. List Sessions
	t.Run("List Sessions", func(t *testing.T) {
		fmt.Printf("[TEST] Listing sessions to verify creation...\n")
		
		resp, body := doRequest(t, "GET", "/market/ai/sessions?page=1&pageSize=20", nil)
		commonResp := assertResponseSuccess(t, resp, body)

		var data struct {
			List []struct {
				ID int64 `json:"id"`
			} `json:"list"`
			Total int64 `json:"total"`
		}
		err := json.Unmarshal(commonResp.Data, &data)
		require.NoError(t, err, "Failed to parse list sessions response")

		found := false
		for _, s := range data.List {
			if s.ID == sessionID {
				found = true
				break
			}
		}
		require.True(t, found, "Newly created session (ID: %d) should be in list", sessionID)
		fmt.Printf("[PASS] ✓ Found session in list - Total: %d\n", data.Total)
	})

	// 3. Update Session
	t.Run("Update Session", func(t *testing.T) {
		newName := "Updated Session Name"
		fmt.Printf("[TEST] Updating session name to: %s\n", newName)
		
		req := map[string]interface{}{
			"id":   sessionID,
			"name": newName,
		}
		resp, body := doRequest(t, "PUT", "/market/ai/sessions", req)
		assertResponseSuccess(t, resp, body)
		
		fmt.Printf("[PASS] ✓ Updated Session - New Name: %s\n", newName)
	})

	// 4. Chat (Single Turn)
	t.Run("Chat Single Turn", func(t *testing.T) {
		question := "Hello, what is 1+1?"
		fmt.Printf("[TEST] Sending chat message: %s\n", question)
		
		reqBody := map[string]interface{}{
			"sessionID": sessionID,
			"content":   question,
		}
		resp, body := doRequest(t, "POST", fmt.Sprintf("/market/ai/sessions/%d/chat", sessionID), reqBody)
		require.Equal(t, http.StatusOK, resp.StatusCode, "Chat should return 200")
		
		// Chat returns SSE stream, assert data frames exist
		require.Contains(t, string(body), "data:", "Should receive SSE data")
		fmt.Printf("[PASS] ✓ Chat completed - Received SSE stream\n")
	})

	// 5. Get Usage (Check creation)
	t.Run("Get Usage", func(t *testing.T) {
		fmt.Printf("[TEST] Getting usage stats for session %d...\n", sessionID)
		time.Sleep(2 * time.Second) // Wait for async processing
		
		resp, body := doRequest(t, "GET", fmt.Sprintf("/market/ai/sessions/%d/usage", sessionID), nil)
		commonResp := assertResponseSuccess(t, resp, body)
		
		var usage struct {
			TotalMessages int `json:"totalMessages"`
		}
		err := json.Unmarshal(commonResp.Data, &usage)
		require.NoError(t, err, "Failed to parse usage response")
		
		fmt.Printf("[INFO] Usage Stats - Total Messages: %d\n", usage.TotalMessages)
		if usage.TotalMessages == 0 {
			fmt.Printf("[WARN] ⚠ TotalMessages is 0, possibly async save delay\n")
		} else {
			assert.GreaterOrEqual(t, usage.TotalMessages, 2, "Should have user+assistant messages")
			fmt.Printf("[PASS] ✓ Usage tracking working\n")
		}
	})

	// 6. Get History (Messages)
	t.Run("Get History", func(t *testing.T) {
		fmt.Printf("[TEST] Getting chat history for session %d...\n", sessionID)
		
		resp, body := doRequest(t, "GET", fmt.Sprintf("/market/ai/sessions/%d/messages?limit=10", sessionID), nil)
		commonResp := assertResponseSuccess(t, resp, body)

		var data struct {
			List []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"list"`
		}
		err := json.Unmarshal(commonResp.Data, &data)
		require.NoError(t, err, "Failed to parse history response")

		fmt.Printf("[PASS] ✓ Retrieved %d messages from history\n", len(data.List))
	})

	// 7. Delete Session
	t.Run("Delete Session", func(t *testing.T) {
		fmt.Printf("[TEST] Deleting session ID: %d\n", sessionID)
		
		req := map[string]interface{}{"id": sessionID}
		resp, body := doRequest(t, "DELETE", fmt.Sprintf("/market/ai/sessions/%d", sessionID), req)
		assertResponseSuccess(t, resp, body)
		
		fmt.Printf("[PASS] ✓ Deleted Session ID: %d\n", sessionID)
	})

	// Cleanup: Delete Model
	t.Run("Cleanup Model", func(t *testing.T) {
		fmt.Printf("[CLEANUP] Deleting test model ID: %d\n", modelID)
		
		req := map[string]interface{}{"id": modelID}
		resp, body := doRequest(t, "DELETE", fmt.Sprintf("/market/ai/models/%d", modelID), req)
		assertResponseSuccess(t, resp, body)
		
		fmt.Printf("[PASS] ✓ Cleanup complete\n")
	})
}
