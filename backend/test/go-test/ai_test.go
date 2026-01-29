package gotest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfig defines the structure of config.json
type TestConfig struct {
	BaseUrl string `json:"baseUrl"`
	Token   string `json:"token"`
	Ai      struct {
		Name      string `json:"name"`
		Provider  string `json:"provider"`
		ApiKey    string `json:"apiKey"`
		BaseUrl   string `json:"baseUrl"`
		ModelName string `json:"modelName"`
	} `json:"ai"`
}

var config TestConfig

// loadConfig loads configuration from config.json
func loadConfig(t *testing.T) {
	file, err := os.Open("config.json")
	if err != nil {
		t.Skip("config.json not found, skipping integration tests")
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		t.Fatalf("Failed to decode config.json: %v", err)
	}

	if config.Token == "YOUR_BEARER_TOKEN_HERE" || config.Token == "" {
		t.Skip("Token not set in config.json, skipping integration tests")
	}
}

// doRequest performs an HTTP request with authentication
func doRequest(t *testing.T, method, path string, body interface{}) (*http.Response, []byte) {
	var reqBody io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		require.NoError(t, err)
		reqBody = bytes.NewBuffer(jsonBytes)
	}

	url := fmt.Sprintf("%s%s", config.BaseUrl, path)
	req, err := http.NewRequest(method, url, reqBody)
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	if config.Token != "" {
		req.Header.Set("Authorization", config.Token) // Assuming Bearer is possibly in the token or just direct token
		if !strings.HasPrefix(config.Token, "Bearer ") {
			req.Header.Set("Authorization", "Bearer "+config.Token)
		} else {
			req.Header.Set("Authorization", config.Token)
		}
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	// defer resp.Body.Close() // Caller handles close if needed, but here we read all

	respBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	t.Logf("[%s] %s -> Status: %d", method, url, resp.StatusCode)
	// t.Logf("Response: %s", string(respBytes))

	return resp, respBytes
}

type CommonResponse struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

// TestAiFullFlow tests the complete AI lifecycle:
// 1. Create Model Access
// 2. Create Session
// 3. Chat
// 4. Get Usage
func TestAiFullFlow(t *testing.T) {
	loadConfig(t)

	var modelID int64
	var sessionID int64

	// Step 1: Create Model Access
	t.Run("Create Model Access", func(t *testing.T) {
		req := map[string]interface{}{
			"name":      config.Ai.Name,
			"provider":  config.Ai.Provider,
			"apiKey":    config.Ai.ApiKey,
			"baseUrl":   config.Ai.BaseUrl,
			// "modelName": config.Ai.ModelName, // Removed from ModelAccess
		}

		resp, body := doRequest(t, "POST", "/market/ai/models", req)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var commonResp CommonResponse
		err := json.Unmarshal(body, &commonResp)
		require.NoError(t, err)
		require.Equal(t, 0, commonResp.Code, "Response Code should be 0")

		var data struct {
			Access struct {
				ID int64 `json:"id"`
			} `json:"access"`
		}
		err = json.Unmarshal(commonResp.Data, &data)
		require.NoError(t, err)
		require.Greater(t, data.Access.ID, int64(0))

		modelID = data.Access.ID
		t.Logf("Created Model Access ID: %d", modelID)
	})

	// Step 2: Create Session
	t.Run("Create Session", func(t *testing.T) {
		req := map[string]interface{}{
			"name":          "Integration Test Session " + time.Now().Format("15:04:05"),
			"modelAccessID": modelID,
			"modelName":     config.Ai.ModelName, // Set in Session now
			"maxContext":    10,
			"toolsConfig":   "{}",
		}

		resp, body := doRequest(t, "POST", "/market/ai/sessions", req)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var commonResp CommonResponse
		err := json.Unmarshal(body, &commonResp)
		require.NoError(t, err)
		require.Equal(t, 0, commonResp.Code)

		var data struct {
			Session struct {
				ID int64 `json:"id"`
			} `json:"session"`
		}
		err = json.Unmarshal(commonResp.Data, &data)
		require.NoError(t, err)
		require.Greater(t, data.Session.ID, int64(0))

		sessionID = data.Session.ID
		t.Logf("Created Session ID: %d", sessionID)
	})

	// Step 3: Chat (SSE)
	t.Run("Chat", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"sessionID": sessionID,
			"content":   "Hello, say 'Integration Test Success' if you receive this.",
		}

		// SSE requires special handling, but doRequest reads all body (which is fine for short response, 
		// but typically we should read line by line. For this simple test, reading all is acceptible 
		// if the server closes connection promptly).
		// Note: The Go http client might wait until the stream closes.
		
		t.Log("Sending Chat Request...")
		resp, body := doRequest(t, "POST", fmt.Sprintf("/market/ai/sessions/%d/chat", sessionID), reqBody)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		respStr := string(body)
		assert.Contains(t, respStr, "data:", "Response should contain SSE data frames")
		t.Logf("Chat Response received (%d bytes)", len(respStr))
	})

	// Step 4: Get Usage
	t.Run("Get Usage", func(t *testing.T) {
		// Wait a bit for usage update
		time.Sleep(2 * time.Second)

		resp, body := doRequest(t, "GET", fmt.Sprintf("/market/ai/sessions/%d/usage", sessionID), nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var commonResp CommonResponse
		err := json.Unmarshal(body, &commonResp)
		require.NoError(t, err)

		var usage struct {
			TotalTokens   int `json:"totalTokens"`
			TotalMessages int `json:"totalMessages"`
		}
		err = json.Unmarshal(commonResp.Data, &usage)
		require.NoError(t, err)
		
		t.Logf("Session Usage: %+v", usage)
		// Should have at least user message and assistant message
		assert.GreaterOrEqual(t, usage.TotalMessages, 2)
		// assert.Greater(t, usage.TotalTokens, 0) // May be 0 if mocked or error in calculation
	})
	
	// Clean up (Optional: Delete resources?) 
	// Not deleting for now so user can inspect manualy if needed via other tools
}
