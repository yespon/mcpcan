package gotest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

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
var configLoaded bool

// loadConfig loads configuration from config.json
func loadConfig(t *testing.T) {
	if configLoaded {
		return
	}
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
	configLoaded = true
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
	// Set X-Consum-User-Id to simulate authenticated user (via AppendUserMiddleware)
	req.Header.Set("X-Consum-User-Id", "1")
	
	if config.Token != "" {
		req.Header.Set("Authorization", config.Token) 
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)

	respBytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	fmt.Printf("[%s] %s -> Status: %d\n", method, path, resp.StatusCode)
	// Pretty print JSON response
	printPrettyJSON(respBytes)

	return resp, respBytes
}

// printPrettyJSON formats and prints JSON response body
func printPrettyJSON(data []byte) {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, data, "", "  "); err != nil {
		// If not valid JSON, print as-is
		fmt.Printf("[RESPONSE] %s\n", string(data))
		return
	}
	fmt.Printf("[RESPONSE]\n%s\n", prettyJSON.String())
}

type CommonResponse struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

// assertResponseSuccess checks struct {code: 0} response
func assertResponseSuccess(t *testing.T, resp *http.Response, body []byte) CommonResponse {
	require.Equal(t, http.StatusOK, resp.StatusCode, "HTTP Status should be 200")

	var commonResp CommonResponse
	err := json.Unmarshal(body, &commonResp)
	require.NoError(t, err, "Should be valid JSON")
	
	// Print error msg if code != 0
	if commonResp.Code != 0 {
		t.Errorf("API Error: Code=%d Msg=%s Data=%s", commonResp.Code, commonResp.Msg, string(commonResp.Data))
	}
	require.Equal(t, 0, commonResp.Code, "Response Code should be 0")
	return commonResp
}
