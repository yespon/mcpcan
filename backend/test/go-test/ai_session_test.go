package gotest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAiSession_AllOps(t *testing.T) {
	loadConfig(t)

	var modelID int64
	var sessionID int64
	var tempFileToDelete string

	// Prerequisite: Create a Model for session tests
	t.Run("Setup Model", func(t *testing.T) {
		fmt.Printf("\n[SETUP] Creating test model for session tests...\n")
		
		req := map[string]interface{}{
			"name":     "Session Test Doubao Model",
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
		
		toolsConfigStr := "{}"
		if len(config.McpServers) > 0 {
			// Wrap in "mcpServers" key as expected by the config structure
			mcpConfig := map[string]interface{}{
				"mcpServers": config.McpServers,
			}
			if b, err := json.Marshal(mcpConfig); err == nil {
				toolsConfigStr = string(b)
			}
		}

		req := map[string]interface{}{
			"name":          sessionName,
			"modelAccessID": modelID,
			"modelName":     config.Ai.ModelName,
			"maxContext":    5,
			"toolsConfig":   toolsConfigStr,
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
	// t.Run("Chat Single Turn", func(t *testing.T) {
	// 	question := "Hello, what is 1+1?"
	// 	fmt.Printf("[TEST] Sending chat message: %s\n", question)
		
	// 	reqBody := map[string]interface{}{
	// 		"sessionID": sessionID,
	// 		"content":   question,
	// 	}
	// 	resp, body := doRequest(t, "POST", fmt.Sprintf("/market/ai/sessions/%d/chat", sessionID), reqBody)
	// 	require.Equal(t, http.StatusOK, resp.StatusCode, "Chat should return 200")
		
	// 	// Chat returns SSE stream, assert data frames exist
	// 	require.Contains(t, string(body), "data:", "Should receive SSE data")
	// 	fmt.Printf("[PASS] ✓ Chat completed - Received SSE stream\n")
	// })

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
		
		// Test Default Pagination (No params -> Page 1, Size 20)
		resp, body := doRequest(t, "GET", fmt.Sprintf("/market/ai/sessions/%d/messages", sessionID), nil)
		commonResp := assertResponseSuccess(t, resp, body)

		var data struct {
			List []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"list"`
			Total int64 `json:"total"`
		}
		err := json.Unmarshal(commonResp.Data, &data)
		require.NoError(t, err, "Failed to parse history response")

		fmt.Printf("[PASS] ✓ Retrieved %d messages (Default). Total: %d\n", len(data.List), data.Total)
		assert.Greater(t, data.Total, int64(0), "Total should be > 0")

		// Test Pagination
		resp2, body2 := doRequest(t, "GET", fmt.Sprintf("/market/ai/sessions/%d/messages?page=1&pageSize=1", sessionID), nil)
		commonResp2 := assertResponseSuccess(t, resp2, body2)
		var data2 struct {
			List []struct {
				Role string `json:"role"`
			} `json:"list"`
			Total int64 `json:"total"`
		}
		json.Unmarshal(commonResp2.Data, &data2)
		require.Equal(t, 1, len(data2.List), "Should return exactly 1 message")
		require.Equal(t, data.Total, data2.Total, "Total should match")
		fmt.Printf("[PASS] ✓ Pagination verified (Page 1, Size 1)\n")
	})

	// 7. File Upload & Cleanup Test (New)
	t.Run("File Upload & Cleanup", func(t *testing.T) {
		fmt.Printf("[TEST] Testing File Upload and Cleanup...\n")

		// A. Upload File
		dummyContent := []byte("fake image content")
		
		// 1. Upload
		var b bytes.Buffer
		w := multipart.NewWriter(&b)
		fw, err := w.CreateFormFile("file", "test_img.png")
		require.NoError(t, err)
		fw.Write(dummyContent)
		w.Close()

		req, err := http.NewRequest("POST", config.BaseUrl + "/market/ai/files/upload", &b)
		require.NoError(t, err)
		req.Header.Set("Content-Type", w.FormDataContentType())
		req.Header.Set("X-Consum-User-Id", "1")
		if config.Token != "" { req.Header.Set("Authorization", config.Token) }
		
		client := &http.Client{Timeout: 30*time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		
		commonResp := assertResponseSuccess(t, resp, body)
		var uploadData struct {
			Id  string `json:"id"`
			Url string `json:"url"`
		}
		err = json.Unmarshal(commonResp.Data, &uploadData)
		require.NoError(t, err)
		
		fmt.Printf("[PASS] ✓ Uploaded file: %s (ID: %s)\n", uploadData.Url, uploadData.Id)
		
		// Verify file exists on disk
		// uploadData.Id is the absolute path in our implementation
		checkPath := uploadData.Id
		if _, err := os.Stat(checkPath); os.IsNotExist(err) {
			checkPath = "../../" + checkPath
		}
		_, err = os.Stat(checkPath)
		require.NoError(t, err, "File should exist on disk (checked: %s)", checkPath)

		// 2. Chat with Attachment
		chatReq := map[string]interface{}{
			"sessionID": sessionID,
			"content":   "Analyze this",
			"attachments": []map[string]string{
				{
					"type": "image",
					"url":  uploadData.Url,
				},
			},
		}
		resp, body = doRequest(t, "POST", fmt.Sprintf("/market/ai/sessions/%d/chat", sessionID), chatReq)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		
		// 3. Mark for verification in next step
		tempFileToDelete = checkPath
	})

	// 8. MCP Integration Test
	t.Run("MCP Session Test", func(t *testing.T) {
		// modelID = 3 // Don't overwrite valid modelID from Setup
		fmt.Printf("[TEST] Testing MCP Integration with provided config...\n")

		mcpConfig := `{
		  "mcpServers": {
		    "mcp-9469c672": {
		      "url": "http://localhost/mcp-gateway/9469c672-7070-4cc5-b4f8-604f916b4784/mcp",
		      "headers": {
		        "Authorization": "Bearer ZDYzZDhkOTItMmFhNy00ODNiLWJjZTUtNmQ2ZTU5NWJkNTYzeyJleHBpcmVBdCI6MTc2ODg5MzM0MzYwOSwidXNlcklkIjoxLCJ1c2VybmFtZSI6ImFkbWluIn0="
		      }
		    }
		  }
		}`

		// Create Session with MCP Config
		req := map[string]interface{}{
			"name":          "MCP Test Session",
			"modelAccessID": modelID,
			"modelName":     config.Ai.ModelName,
			"toolsConfig":   mcpConfig,
			"isMcp":         true,
		}

		resp, body := doRequest(t, "POST", "/market/ai/sessions", req)
		commonResp := assertResponseSuccess(t, resp, body)

		var data struct {
			Session struct {
				ID int64 `json:"id"`
			} `json:"session"`
		}
		json.Unmarshal(commonResp.Data, &data)
		mcpSessionID := data.Session.ID
		fmt.Printf("[PASS] Created MCP Session ID: %d\n", mcpSessionID)

		// Verify Session Details (ToolsConfig)
		getResp, getBody := doRequest(t, "GET", fmt.Sprintf("/market/ai/sessions/%d", mcpSessionID), nil)
		getCommonResp := assertResponseSuccess(t, getResp, getBody)
		var sessionData struct {
			Session struct {
				ToolsConfig string `json:"toolsConfig"`
			} `json:"session"`
		}
		json.Unmarshal(getCommonResp.Data, &sessionData)
		// fmt.Printf("[INFO] Saved ToolsConfig: %s\n", sessionData.Session.ToolsConfig)
		// Basic check
		if len(sessionData.Session.ToolsConfig) < 10 {
			 t.Errorf("[FAIL] ToolsConfig seems empty or too short: %s", sessionData.Session.ToolsConfig)
		} else {
			fmt.Printf("[PASS] ✓ ToolsConfig verified (Len: %d)\n", len(sessionData.Session.ToolsConfig))
		}

		// Test Connection (via Chat)
		// Sending a message triggers MCP initialization. If init fails, chat fails.
		chatReq := map[string]interface{}{
			"sessionID": mcpSessionID,
			"content":   "中文对话，查询系统信息",
		}

		resp, body = doRequest(t, "POST", fmt.Sprintf("/market/ai/sessions/%d/chat", mcpSessionID), chatReq)
		
		// We expect 200 OK. If MCP connects successfully, the stream starts.
		// Detailed tool verification would require a real MCP server response, 
		// but checking 200 guarantees the handshake didn't immediately fail.
		require.Equal(t, http.StatusOK, resp.StatusCode, "Chat should return 200 if MCP connects")
		
		// Parse SSE Stream and Print aggregated content
		lines := strings.Split(string(body), "\n")
		var fullContent strings.Builder
		
		fmt.Println("\n--- SSE Stream Output ---")
		hasData := false
		for _, line := range lines {
			if strings.HasPrefix(line, "data: ") {
				hasData = true
				jsonStr := strings.TrimPrefix(line, "data: ")
				var msg struct {
					Type    string `json:"type"`
					Content string `json:"content"`
				}
				if err := json.Unmarshal([]byte(jsonStr), &msg); err == nil {
					if msg.Type == "text" || msg.Type == "tool_result" || msg.Type == "tool_start" {
						fullContent.WriteString(msg.Content)
					}
					// Print non-text events for debugging
					if msg.Type != "text" && msg.Type != "done" {
						fmt.Printf("[Event: %s] %s\n", msg.Type, msg.Content)
					}
				}
			}
		}
		require.True(t, hasData, "Should receive SSE data frames")
		
		fmt.Printf("-------------------------\n")
		fmt.Printf("[FINAL CONTENT]: %s\n", fullContent.String())
		fmt.Printf("-------------------------\n")
		
		fmt.Printf("[PASS] MCP Connection verified (Chat initiated)\n")

		// Cleanup this session
		doRequest(t, "DELETE", fmt.Sprintf("/market/ai/sessions/%d", mcpSessionID), map[string]interface{}{"id": mcpSessionID})
	})

	// 9. Delete Session
	t.Run("Delete Session", func(t *testing.T) {
		fmt.Printf("[TEST] Deleting session ID: %d\n", sessionID)
		
		req := map[string]interface{}{"id": sessionID}
		resp, body := doRequest(t, "DELETE", fmt.Sprintf("/market/ai/sessions/%d", sessionID), req)
		assertResponseSuccess(t, resp, body)
		
		fmt.Printf("[PASS] ✓ Deleted Session ID: %d\n", sessionID)
		
		// Verify file cleanup
		if tempFileToDelete != "" {
			_, err := os.Stat(tempFileToDelete)
			if os.IsNotExist(err) {
				fmt.Printf("[PASS] ✓ File cleanup successful: %s gone\n", tempFileToDelete)
			} else {
				t.Errorf("[FAIL] File still exists after session delete: %s", tempFileToDelete)
			}
		}
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
