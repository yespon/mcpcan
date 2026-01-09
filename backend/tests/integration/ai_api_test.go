package integration

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	pb "github.com/kymo-mcp/mcpcan/api/market/ai_agent"
)

// Package integration provides integration tests for the AI API endpoints.
//
// Usage:
// 1. Ensure the backend server is running on localhost:8080.
// 2. Run tests using `go test -v ./backend/tests/integration/ai_api_test.go`
//
// Configuration:
// The tests load configuration from a .env file (e.g., backend/tests/integration/.env).
// Sensitive data like API Keys should not be hardcoded.

// Configuration (Global Variables loaded in init)
var (
	baseURL       string
	TestApiKey    string
	TestModelName string
	TestBaseURL   string
	TestProvider  string
)

func init() {
	loadEnv()
}

func loadEnv() {
	// Try multiple paths to find .env
	envPaths := []string{
		".env",
		"backend/tests/integration/.env",
		"tests/integration/.env",
		"../.env", // If running from integration dir
		"/Users/nolan/go/src/mcpcan/backend/tests/integration/.env", // Absolute path fallback
	}

	var envFile string
	for _, path := range envPaths {
		if _, err := os.Stat(path); err == nil {
			envFile = path
			break
		}
	}

	if envFile != "" {
		fmt.Printf("Loading config from %s\n", envFile)
		file, err := os.Open(envFile)
		if err != nil {
			fmt.Printf("Warning: Failed to open .env file: %v\n", err)
		} else {
			defer file.Close()
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])
					// Only set if not already set in environment
					if _, exists := os.LookupEnv(key); !exists {
						os.Setenv(key, value)
					}
				}
			}
		}
	} else {
		fmt.Println("Warning: .env file not found, relying on system environment variables")
	}

	// Load into variables
	baseURL = getEnv("TEST_SERVER_URL", "http://localhost:8080")
	TestApiKey = getEnv("TEST_API_KEY", "")
	TestModelName = getEnv("TEST_MODEL_NAME", "")
	TestBaseURL = getEnv("TEST_BASE_URL", "")
	TestProvider = getEnv("TEST_PROVIDER", "")

	// Validate critical config
	if TestApiKey == "" {
		fmt.Println("Warning: TEST_API_KEY is not set")
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// --- Struct Definitions ---

// Common API Response Wrapper
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// Note: Other structs are now imported from "github.com/kymo-mcp/mcpcan/api/market/ai_agent" (as pb)

// --- Helper Functions ---

func makeRequest(t *testing.T, method, path string, body interface{}, response interface{}) {
	url := baseURL + path
	var reqBody io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
		reqBody = bytes.NewBuffer(jsonBytes)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second} // Increased timeout for real API calls
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if resp.StatusCode >= 400 {
		t.Fatalf("API returned error status %d: %s", resp.StatusCode, string(respBody))
	}

	if response != nil {
		// Check if it's wrapped in standard API response (code/message/data)
		// We try to unmarshal into APIResponse first if the target isn't explicitly handling it
		var apiResp APIResponse
		// We create a temporary map to check structure
		var tempMap map[string]interface{}
		json.Unmarshal(respBody, &tempMap)

		if _, ok := tempMap["code"]; ok {
			// It is a wrapped response
			apiResp.Data = response // Pointer to the target struct
			if err := json.Unmarshal(respBody, &apiResp); err != nil {
				t.Fatalf("Failed to unmarshal API response wrapper: %v, body: %s", err, string(respBody))
			}
			if apiResp.Code > 0 {
				t.Fatalf("API Error Code %d: %s", apiResp.Code, apiResp.Message)
			}
		} else {
			// Direct response
			if err := json.Unmarshal(respBody, response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v, body: %s", err, string(respBody))
			}
		}
	}
}

// Helper to create a model access for other tests
func createTestModelAccess(t *testing.T) int64 {
	req := pb.CreateModelAccessRequest{
		Name:      "Test Helper Doubao",
		Provider:  TestProvider,
		ApiKey:    TestApiKey,
		BaseUrl:   TestBaseURL,
		ModelName: TestModelName,
	}
	var resp pb.CreateModelAccessResponse
	makeRequest(t, "POST", "/ai/models", req, &resp)
	return resp.Access.Id
}

// Helper to delete model access
func deleteTestModelAccess(t *testing.T, id int64) {
	if id == 0 {
		return
	}
	path := fmt.Sprintf("/ai/models/%d", id)
	url := baseURL + path
	req, _ := http.NewRequest("DELETE", url, nil)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err == nil {
		resp.Body.Close()
	} else {
		t.Logf("Warning: Failed to delete model access %d: %v", id, err)
	}
}

// Helper to create a session
func createTestSession(t *testing.T, modelID int64) int64 {
	req := pb.CreateSessionRequest{
		Name:          "Test Helper Session",
		ModelAccessID: modelID,
		MaxContext:    10,
	}
	var resp pb.CreateSessionResponse
	makeRequest(t, "POST", "/ai/sessions", req, &resp)
	return resp.Session.Id
}

// Helper to delete session
func deleteTestSession(t *testing.T, id int64) {
	if id == 0 {
		return
	}
	path := fmt.Sprintf("/ai/sessions/%d", id)
	url := baseURL + path
	req, _ := http.NewRequest("DELETE", url, nil)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err == nil {
		resp.Body.Close()
	} else {
		t.Logf("Warning: Failed to delete session %d: %v", id, err)
	}
}

// --- Individual Tests ---

// 1. GET /ai/models/supported
// Description: Retrieves the list of supported AI providers and models.
func TestGetSupportedModels(t *testing.T) {
	var resp pb.GetSupportedModelsResponse
	makeRequest(t, "GET", "/ai/models/supported", nil, &resp)

	if len(resp.Providers) == 0 {
		t.Error("Expected supported providers, got empty list")
	}
	found := false
	for _, p := range resp.Providers {
		if p.Id == TestProvider {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected provider %s to be supported", TestProvider)
	}
	t.Logf("Found %d providers", len(resp.Providers))
	// fmt.Printf("Providers: %+v\n", resp.Providers)
}

// 2. POST /ai/models
// Description: Creates a new AI model access configuration.
func TestCreateModelAccess(t *testing.T) {
	req := pb.CreateModelAccessRequest{
		Name:      "Integration Test Create",
		Provider:  TestProvider,
		ApiKey:    TestApiKey,
		BaseUrl:   TestBaseURL,
		ModelName: TestModelName,
	}
	var resp pb.CreateModelAccessResponse
	makeRequest(t, "POST", "/ai/models", req, &resp)

	if resp.Access.Id == 0 {
		t.Error("Created model access ID is 0")
	}
	t.Logf("Created Model Access: %+v", resp.Access)

	// Cleanup
	deleteTestModelAccess(t, resp.Access.Id)
}

// 3. GET /ai/models
// Description: Lists all AI model access configurations (with pagination).
func TestListModelAccess(t *testing.T) {
	// Ensure at least one exists
	id := createTestModelAccess(t)
	defer deleteTestModelAccess(t, id)

	var resp pb.ListModelAccessResponse
	makeRequest(t, "GET", "/ai/models?page=1&pageSize=10", nil, &resp)

	t.Logf("Total Models: %d", resp.Total)
	if len(resp.List) == 0 {
		t.Error("Expected at least one model access")
	}
}

// 4. GET /ai/models/:id
// Description: Retrieves details of a specific AI model access configuration.
func TestGetModelAccess(t *testing.T) {
	id := createTestModelAccess(t)
	defer deleteTestModelAccess(t, id)

	path := fmt.Sprintf("/ai/models/%d", id)
	var resp pb.GetModelAccessResponse
	makeRequest(t, "GET", path, nil, &resp)

	if resp.Access.Id != id {
		t.Errorf("Expected ID %d, got %d", id, resp.Access.Id)
	}
	t.Logf("Retrieved Model Access: %s", resp.Access.Name)
}

// 5. PUT /ai/models
// Description: Updates an existing AI model access configuration.
func TestUpdateModelAccess(t *testing.T) {
	id := createTestModelAccess(t)
	defer deleteTestModelAccess(t, id)

	req := pb.UpdateModelAccessRequest{
		Id:   id,
		Name: "Updated Name Integration Test",
	}
	var resp pb.UpdateModelAccessResponse
	makeRequest(t, "PUT", "/ai/models", req, &resp)

	if resp.Access.Name != "Updated Name Integration Test" {
		t.Errorf("Expected name update, got %s", resp.Access.Name)
	}
}

// 6. DELETE /ai/models/:id
// Description: Deletes an AI model access configuration.
func TestDeleteModelAccess(t *testing.T) {
	// Create one specifically to delete
	id := createTestModelAccess(t)

	path := fmt.Sprintf("/ai/models/%d", id)
	makeRequest(t, "DELETE", path, nil, nil)

	// Verify it's gone (should return 404 or error code)
	url := baseURL + path
	req, _ := http.NewRequest("GET", url, nil)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Logf("Delete verification request failed: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		// If backend returns 200 with error code in body
		body, _ := io.ReadAll(resp.Body)
		var apiResp APIResponse
		json.Unmarshal(body, &apiResp)
		if apiResp.Code == 200 {
			t.Error("Model access should have been deleted, but GET returned 200 OK")
		} else {
			t.Logf("Confirmed deletion, GET returned application error: %d", apiResp.Code)
		}
	} else {
		t.Logf("Confirmed deletion, GET returned status: %d", resp.StatusCode)
	}
}

// 7. GET /ai/models/available
// Description: Lists available models for selection (simplified list).
func TestGetAvailableModels(t *testing.T) {
	id := createTestModelAccess(t)
	defer deleteTestModelAccess(t, id)

	var resp pb.GetAvailableModelsResponse
	makeRequest(t, "GET", "/ai/models/available", nil, &resp)

	if len(resp.List) == 0 {
		t.Error("Expected at least one available model")
	}
	t.Logf("Available Models Count: %d", len(resp.List))
}

// 8. POST /ai/models/:id/test
// Description: Tests connection using a saved model access configuration ID.
func TestTestConnectionByID(t *testing.T) {
	id := createTestModelAccess(t)
	defer deleteTestModelAccess(t, id)

	var resp pb.TestConnectionResponse
	path := fmt.Sprintf("/ai/models/%d/test", id)
	makeRequest(t, "POST", path, nil, &resp)

	t.Logf("Connection Test (By ID): Success=%v, Message=%s, Latency=%dms", resp.Success, resp.Message, resp.LatencyMs)
	if !resp.Success {
		t.Errorf("Connection test failed: %s", resp.Message)
	}
}

// 9. POST /ai/models/test
// Description: Tests connection using provided credentials directly (without saving).
func TestTestConnectionDirect(t *testing.T) {
	req := pb.TestConnectionRequest{
		Provider:  TestProvider,
		ApiKey:    TestApiKey,
		BaseUrl:   TestBaseURL,
		ModelName: TestModelName,
	}
	var resp pb.TestConnectionResponse
	makeRequest(t, "POST", "/ai/models/test", req, &resp)

	t.Logf("Connection Test (Direct): Success=%v, Message=%s, Latency=%dms", resp.Success, resp.Message, resp.LatencyMs)
	if !resp.Success {
		t.Errorf("Connection test failed: %s", resp.Message)
	}
}

// 10. POST /ai/sessions
// Description: Creates a new chat session.
func TestCreateSession(t *testing.T) {
	modelID := createTestModelAccess(t)
	defer deleteTestModelAccess(t, modelID)

	req := pb.CreateSessionRequest{
		Name:          "Integration Test Session",
		ModelAccessID: modelID,
		MaxContext:    10,
	}
	var resp pb.CreateSessionResponse
	makeRequest(t, "POST", "/ai/sessions", req, &resp)

	if resp.Session.Id == 0 {
		t.Error("Created session ID is 0")
	}
	t.Logf("Created Session: %+v", resp.Session)

	// Cleanup
	deleteTestSession(t, resp.Session.Id)
}

// 11. GET /ai/sessions
// Description: Lists all chat sessions.
func TestListSessions(t *testing.T) {
	modelID := createTestModelAccess(t)
	defer deleteTestModelAccess(t, modelID)
	sessionID := createTestSession(t, modelID)
	defer deleteTestSession(t, sessionID)

	var resp pb.ListSessionsResponse
	makeRequest(t, "GET", "/ai/sessions?page=1&pageSize=10", nil, &resp)

	if len(resp.List) == 0 {
		t.Error("Expected at least one session")
	}
	t.Logf("Total Sessions: %d", resp.Total)
}

// 12. GET /ai/sessions/:id
// Description: Retrieves details of a specific chat session.
func TestGetSession(t *testing.T) {
	modelID := createTestModelAccess(t)
	defer deleteTestModelAccess(t, modelID)
	sessionID := createTestSession(t, modelID)
	defer deleteTestSession(t, sessionID)

	path := fmt.Sprintf("/ai/sessions/%d", sessionID)
	var resp pb.GetSessionResponse
	makeRequest(t, "GET", path, nil, &resp)

	if resp.Session.Id != sessionID {
		t.Errorf("Expected Session ID %d, got %d", sessionID, resp.Session.Id)
	}
}

// 13. PUT /ai/sessions
// Description: Updates an existing chat session.
func TestUpdateSession(t *testing.T) {
	modelID := createTestModelAccess(t)
	defer deleteTestModelAccess(t, modelID)
	sessionID := createTestSession(t, modelID)
	defer deleteTestSession(t, sessionID)

	req := pb.UpdateSessionRequest{
		Id:   sessionID,
		Name: "Updated Session Name",
	}
	var resp pb.UpdateSessionResponse
	makeRequest(t, "PUT", "/ai/sessions", req, &resp)

	if resp.Session.Name != "Updated Session Name" {
		t.Errorf("Expected name update, got %s", resp.Session.Name)
	}
}

// 14. DELETE /ai/sessions/:id
// Description: Deletes a chat session.
func TestDeleteSession(t *testing.T) {
	modelID := createTestModelAccess(t)
	defer deleteTestModelAccess(t, modelID)
	sessionID := createTestSession(t, modelID)

	path := fmt.Sprintf("/ai/sessions/%d", sessionID)
	makeRequest(t, "DELETE", path, nil, nil)

	// Verify
	url := baseURL + path
	req, _ := http.NewRequest("GET", url, nil)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Logf("Delete verification request failed: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		var apiResp APIResponse
		json.Unmarshal(body, &apiResp)
		if apiResp.Code == 200 {
			t.Error("Session should have been deleted")
		}
	} else {
		t.Logf("Confirmed session deletion, status: %d", resp.StatusCode)
	}
}

// 15. POST /ai/sessions/:id/chat
// Description: Initiates a chat message in the session (expects SSE stream).
func TestChatSSEInitiation(t *testing.T) {
	modelID := createTestModelAccess(t)
	defer deleteTestModelAccess(t, modelID)
	sessionID := createTestSession(t, modelID)
	defer deleteTestSession(t, sessionID)

	path := fmt.Sprintf("/ai/sessions/%d/chat", sessionID)
	url := baseURL + path
	// Use Content instead of Message, and SessionID
	reqBody, _ := json.Marshal(pb.ChatRequest{
		SessionID: sessionID,
		Content:   "Hello, this is an integration test.",
	})

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	client := &http.Client{Timeout: 15 * time.Second} // Allow time for LLM response
	resp, err := client.Do(req)
	if err != nil {
		t.Logf("Chat request error: %v", err)
	} else {
		defer resp.Body.Close()
		t.Logf("Chat SSE Status: %d", resp.StatusCode)

		// Basic check if we got a stream
		contentType := resp.Header.Get("Content-Type")
		t.Logf("Content-Type: %s", contentType)

		// Read first chunk to verify connectivity
		buf := make([]byte, 1024)
		n, _ := resp.Body.Read(buf)
		if n > 0 {
			t.Logf("Received initial stream data: %s", string(buf[:n]))
		}
	}
}

// 16. GET /ai/sessions/:id/messages
// Description: Retrieves chat history/messages for a session.
func TestGetSessionMessages(t *testing.T) {
	modelID := createTestModelAccess(t)
	defer deleteTestModelAccess(t, modelID)
	sessionID := createTestSession(t, modelID)
	defer deleteTestSession(t, sessionID)

	// Note: Without sending a chat first, this might be empty, but should still succeed
	path := fmt.Sprintf("/ai/sessions/%d/messages", sessionID)
	var resp pb.GetSessionMessagesResponse
	makeRequest(t, "GET", path, nil, &resp)

	t.Logf("Retrieved %d messages", len(resp.List))
}
