package n8n

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func GetCookieFromLogin(host string, email, password string) (string, error) {
	// 准备请求数据
	data := map[string]string{
		"emailOrLdapLoginId": email,
		"password":           password,
	}

	jsonData, _ := json.Marshal(data)

	// 发送请求
	resp, err := http.Post(
		fmt.Sprintf("%s/rest/login", host),
		"application/json",
		bytes.NewBuffer(jsonData),
	)

	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("login Failed: %s", string(body))
	}

	// 获取Set-Cookie
	cookies := resp.Header.Values("Set-Cookie")
	if len(cookies) == 0 {
		return "", fmt.Errorf("not Found Set-Cookie")
	}

	return strings.Join(cookies, "; "), nil
}

// 定义插件相关的数据结构
type InstalledNode struct {
	Name          string `json:"name"`
	Type          string `json:"type"`
	LatestVersion int    `json:"latestVersion"`
}

type PluginPackage struct {
	UpdatedAt        string          `json:"updatedAt"`
	CreatedAt        string          `json:"createdAt"`
	PackageName      string          `json:"packageName"`
	InstalledVersion string          `json:"installedVersion"`
	AuthorName       string          `json:"authorName"`
	AuthorEmail      string          `json:"authorEmail"`
	InstalledNodes   []InstalledNode `json:"installedNodes"`
}

type PluginResponse struct {
	Data []PluginPackage `json:"data"`
}

// 获取插件列表
func GetPluginList(host string, cookieStr string) (*PluginResponse, error) {
	// 创建请求
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/rest/community-packages", host), nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置Cookie头
	req.Header.Set("Cookie", cookieStr)

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取插件列表失败 (状态码: %d): %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var pluginResp PluginResponse
	err = json.Unmarshal(body, &pluginResp)
	if err != nil {
		return nil, fmt.Errorf("解析JSON失败: %v", err)
	}

	return &pluginResp, nil
}

// InstallPackage 安装 n8n 包的通用函数
func InstallPackage(baseURL, packageName, cookie string) (map[string]interface{}, error) {
	// 构建请求 URL
	url := baseURL + "/rest/community-packages"

	// 准备请求体
	requestBody := map[string]string{
		"name": packageName,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// 创建请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 解析响应
	var result map[string]interface{}
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	return result, nil
}

// Project 项目结构体
type Project struct {
	UpdatedAt   string   `json:"updatedAt"`
	CreatedAt   string   `json:"createdAt"`
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Icon        *string  `json:"icon"`
	Description *string  `json:"description"`
	Role        string   `json:"role"`
	Scopes      []string `json:"scopes"`
}

// ProjectsResponse 项目列表响应结构体
type ProjectsResponse struct {
	Data []Project `json:"data"`
}

// GetProjects 获取 n8n 项目列表
func GetProjects(baseURL, cookie string) (*ProjectsResponse, error) {
	// 构建请求 URL
	url := baseURL + "/rest/projects/my-projects"

	// 创建 HTTP 请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Accept", "application/json")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	// 解析响应
	var response ProjectsResponse
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// CredentialResponse 凭证响应的对象封装
type CredentialResponse struct {
	Data []CredentialItem `json:"data"`
}

// CredentialItem 单个凭证对象
type CredentialItem struct {
	UpdatedAt          string         `json:"updatedAt"`
	CreatedAt          string         `json:"createdAt"`
	ID                 string         `json:"id"`
	Name               string         `json:"name"`
	Data               CredentialData `json:"data"`
	Type               string         `json:"type"`
	IsManaged          bool           `json:"isManaged"`
	IsGlobal           bool           `json:"isGlobal"`
	HomeProject        ProjectInfo    `json:"homeProject"`
	SharedWithProjects []ProjectInfo  `json:"sharedWithProjects"`
	Scopes             []string       `json:"scopes"`
}

// CredentialData 凭证数据（动态类型）
type CredentialData map[string]interface{}

// ProjectInfo 项目信息
type ProjectInfo struct {
	ID   string      `json:"id"`
	Type string      `json:"type"`
	Name string      `json:"name"`
	Icon interface{} `json:"icon"`
}

// GetCredentials 获取凭证列表，返回封装的对象
func GetCredentials(baseURL, cookie, projectID string) (*CredentialResponse, error) {
	// 构建查询参数
	params := url.Values{}

	// 添加 filter 参数
	filter := map[string]string{"projectId": projectID}
	filterJSON, _ := json.Marshal(filter)
	params.Add("filter", string(filterJSON))

	// 添加其他参数
	params.Add("includeScopes", fmt.Sprintf("%v", true))
	params.Add("includeData", fmt.Sprintf("%v", true))
	params.Add("includeGlobal", fmt.Sprintf("%v", true))

	// 构建完整 URL
	apiURL := fmt.Sprintf("%s/rest/credentials?%s", baseURL, params.Encode())

	// 创建 HTTP 请求
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	// 设置请求头
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Accept", "application/json")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()

	// 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	// 解析响应
	var response CredentialResponse
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&response); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	return &response, nil
}

// CreateMCPCredentialRequest 创建 MCP 凭证的请求参数
type CreateMCPCredentialRequest struct {
	Name      string            `json:"name"`
	Type      string            `json:"type"`
	Data      MCPCredentialData `json:"data"`
	ProjectID string            `json:"projectId"`
	IsGlobal  bool              `json:"isGlobal"`
}

// MCPCredentialData MCP 凭证数据
type MCPCredentialData struct {
	HTTPStreamURL string `json:"httpStreamUrl"`
	Headers       string `json:"headers"`
}

// CreateCredentialResponse 创建凭证的响应
type CreateCredentialResponse struct {
	Data Credential `json:"data"`
}

// Credential 凭证对象
type Credential struct {
	UpdatedAt string   `json:"updatedAt"`
	Name      string   `json:"name"`
	Data      string   `json:"data"` // 注意：这里是加密后的字符串
	Type      string   `json:"type"`
	IsManaged bool     `json:"isManaged"`
	IsGlobal  bool     `json:"isGlobal"`
	ID        string   `json:"id"`
	CreatedAt string   `json:"createdAt"`
	Scopes    []string `json:"scopes"`
}

// CreateMCPCredential 创建 MCP 插件凭证
func CreateMCPCredential(baseURL, cookie, name, httpStreamURL, headers, projectID string) (*CreateCredentialResponse, error) {
	// 构建请求 URL
	apiURL := baseURL + "/rest/credentials"

	// 构建请求体
	requestBody := CreateMCPCredentialRequest{
		Name: name,
		Type: "mcpClientHttpApi",
		Data: MCPCredentialData{
			HTTPStreamURL: httpStreamURL,
			Headers:       headers,
		},
		ProjectID: projectID,
		IsGlobal:  true,
	}

	// 将请求体转换为 JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request body failed: %w", err)
	}

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Accept", "application/json")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()

	// 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := json.Marshal(resp.Body)
		return nil, fmt.Errorf("request failed with status %d, response: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var response CreateCredentialResponse
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&response); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	return &response, nil
}

// UpdateCredential 更新凭据信息
func UpdateCredential(baseURL, cookie, id, name, httpStreamURL, headers string) (map[string]interface{}, error) {
	// 构造请求数据
	requestData := map[string]interface{}{
		"id":   id,
		"name": name,
		"type": "mcpClientHttpApi",
		"data": map[string]interface{}{
			"httpStreamUrl": httpStreamURL,
			"headers":       headers,
		},
		"isGlobal": true,
	}

	// 将请求数据转换为 JSON
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("marshal request data failed: %w", err)
	}

	// 构造完整的 URL
	fullURL := fmt.Sprintf("%s/rest/credentials/%s", baseURL, id)

	// 创建 HTTP 请求
	req, err := http.NewRequest("PATCH", fullURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Accept", "application/json")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request failed: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body failed: %v", err)
	}

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("response failed with status %d, response: %s", resp.StatusCode, string(body))
	}

	// 解析响应 JSON
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response JSON failed: %v", err)
	}

	return result, nil
}
