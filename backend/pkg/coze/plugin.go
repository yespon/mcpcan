package coze

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// 定义请求结构体
type PluginListRequest struct {
	UserFilter          int    `json:"user_filter"`
	ResTypeFilter       []int  `json:"res_type_filter"`
	Name                string `json:"name"`
	PublishStatusFilter int    `json:"publish_status_filter"`
	SpaceID             string `json:"space_id"`
	Size                int    `json:"size"`
}

// 定义响应结构体
type PluginListResponse struct {
	Code         int    `json:"code"`
	Cursor       string `json:"cursor"`
	HasMore      bool   `json:"has_more"`
	Msg          string `json:"msg"`
	Total        int    `json:"total"`
	ResourceList []struct {
		Actions []struct {
			Enable bool `json:"enable"`
			Key    int  `json:"key"`
		} `json:"actions"`
		CreatorAvatar string `json:"creator_avatar"`
		CreatorID     string `json:"creator_id"`
		CreatorName   string `json:"creator_name"`
		Desc          string `json:"desc"`
		DetailDisable bool   `json:"detail_disable"`
		EditTime      int    `json:"edit_time"`
		Icon          string `json:"icon"`
		Name          string `json:"name"`
		PublishStatus int    `json:"publish_status"`
		ResID         string `json:"res_id"`
		ResSubType    int    `json:"res_sub_type"`
		ResType       int    `json:"res_type"`
		SpaceID       string `json:"space_id"`
		UserName      string `json:"user_name"`
	} `json:"resource_list"`
}

// 插件信息结构体（只包含需要的字段）
type PluginInfo struct {
	ResID         string `json:"res_id"`
	Name          string `json:"name"`
	Desc          string `json:"desc"`
	Icon          string `json:"icon"`
	CreatorID     string `json:"creator_id"`
	CreatorName   string `json:"creator_name"`
	UserName      string `json:"user_name"`
	PublishStatus int    `json:"publish_status"`
	ResType       int    `json:"res_type"`
	ResSubType    int    `json:"res_sub_type"`
	EditTime      int    `json:"edit_time"`
	SpaceID       string `json:"space_id"`
}

// Action信息
type ActionInfo struct {
	Enable bool `json:"enable"`
	Key    int  `json:"key"`
}

// 包含Actions的完整插件信息
type PluginInfoWithActions struct {
	PluginInfo
	Actions []ActionInfo `json:"actions"`
}

func GetPluginList(cookie, spaceID string) ([]PluginInfo, error) {
	// 准备请求体
	requestBody := PluginListRequest{
		UserFilter:          0,
		ResTypeFilter:       []int{1, -1}, // 资源类型过滤器
		Name:                "",
		PublishStatusFilter: 0,
		SpaceID:             spaceID,
		Size:                100,
	}

	return fetchPlugins(cookie, requestBody)
}

func fetchPlugins(cookie string, requestBody PluginListRequest) ([]PluginInfo, error) {
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request body failed: %v", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/api/plugin_api/library_resource_list", host),
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("agw-js-conv", "str")
	req.Header.Set("Cookie", cookie)

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request failed: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body failed: %v", err)
	}

	// 解析响应
	var response PluginListResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("marshal JSON response body failed: %v", err)
	}

	// 检查响应状态
	if response.Code != 0 {
		return nil, fmt.Errorf("API result error: %s", response.Msg)
	}

	// 转换为PluginInfo结构体
	var plugins []PluginInfo
	for _, resource := range response.ResourceList {
		plugin := PluginInfo{
			ResID:         resource.ResID,
			Name:          resource.Name,
			Desc:          resource.Desc,
			Icon:          resource.Icon,
			CreatorID:     resource.CreatorID,
			CreatorName:   resource.CreatorName,
			UserName:      resource.UserName,
			PublishStatus: resource.PublishStatus,
			ResType:       resource.ResType,
			ResSubType:    resource.ResSubType,
			EditTime:      resource.EditTime,
			SpaceID:       resource.SpaceID,
		}
		plugins = append(plugins, plugin)
	}

	return plugins, nil
}

// 请求结构体
type GetPluginInfoRequest struct {
	PluginID string `json:"plugin_id"`
}

// 响应结构体
type GetPluginInfoResponse struct {
	Code                int                    `json:"code"`
	CreationMethod      int                    `json:"creation_method"`
	Creator             CreatorInfo            `json:"creator"`
	EditVersion         int                    `json:"edit_version"`
	IdeCodeRuntime      string                 `json:"ide_code_runtime"`
	MetaInfo            MetaInfo               `json:"meta_info"`
	Msg                 string                 `json:"msg"`
	PluginProductStatus int                    `json:"plugin_product_status"`
	PluginType          int                    `json:"plugin_type"`
	PrivacyInfo         string                 `json:"privacy_info"`
	PrivacyStatus       bool                   `json:"privacy_status"`
	Published           bool                   `json:"published"`
	StatisticData       map[string]interface{} `json:"statistic_data"`
	Status              bool                   `json:"status"`
}

type CreatorInfo struct {
	AvatarURL      string                 `json:"avatar_url"`
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Self           bool                   `json:"self"`
	SpaceRolyType  int                    `json:"space_roly_type"`
	UserLabel      map[string]interface{} `json:"user_label"`
	UserUniqueName string                 `json:"user_unique_name"`
}

type MetaInfo struct {
	AuthType            []int                  `json:"auth_type"`
	BenefitReportConfig map[string]interface{} `json:"benefit_report_config"`
	BriefIntro          string                 `json:"brief_intro"`
	CommonParams        map[string][]ParamItem `json:"common_params"`
	Desc                string                 `json:"desc"`
	FixedExportIP       bool                   `json:"fixed_export_ip"`
	Icon                IconInfo               `json:"icon"`
	Name                string                 `json:"name"`
	OauthInfo           string                 `json:"oauth_info"`
	URL                 string                 `json:"url"`
}

type ParamItem struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type IconInfo struct {
	URI string `json:"uri"`
	URL string `json:"url"`
}

// GetPluginInfo 获取插件详情
func GetPluginInfo(cookie string, pluginID string) (*GetPluginInfoResponse, error) {
	// 创建请求数据
	reqData := GetPluginInfoRequest{
		PluginID: pluginID,
	}

	// 序列化请求数据
	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return nil, fmt.Errorf("marshal request data failed: %v", err)
	}

	// 创建 HTTP 请求
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/api/plugin_api/get_plugin_info", host),
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("agw-js-conv", "str")
	req.Header.Set("Cookie", cookie)

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

	// 解析响应
	var result GetPluginInfoResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("unmarshal JSON response body failed: %v", err)
	}

	return &result, nil
}

// 请求结构体
type RegisterPluginRequest struct {
	Name           string                 `json:"name"`
	Desc           string                 `json:"desc"`
	URL            string                 `json:"url"`
	Icon           IconRequest            `json:"icon"`
	AuthType       int                    `json:"auth_type"`
	OauthInfo      string                 `json:"oauth_info"`
	SpaceID        string                 `json:"space_id"`
	CommonParams   map[string][]ParamItem `json:"common_params"`
	IdeCodeRuntime string                 `json:"ide_code_runtime"`
	PluginType     int                    `json:"plugin_type"`
	BriefIntro     string                 `json:"brief_intro"`
}

type IconRequest struct {
	URI string `json:"uri,omitempty"`
	URL string `json:"url,omitempty"`
}

// 响应结构体
type RegisterPluginResponse struct {
	Code     int    `json:"code"`
	Msg      string `json:"msg"`
	PluginID string `json:"plugin_id"`
}

// RegisterPlugin 新增插件
func RegisterPlugin(reqData *RegisterPluginRequest, cookie string) (*RegisterPluginResponse, error) {
	// 序列化请求数据
	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return nil, fmt.Errorf("marshal request data failed: %v", err)
	}

	// 创建 HTTP 请求
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/api/plugin_api/register_plugin_meta", host),
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)

	// 发送请求
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
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

	// 解析响应
	var result RegisterPluginResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("unmarshal JSON response body failed: %v, body: %s", err, string(body))
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("register plugin failed: %v", result.Msg)
	}

	return &result, nil
}

type UpdatePluginRequest struct {
	PluginID     string                 `json:"plugin_id"`
	Name         string                 `json:"name"`
	Desc         string                 `json:"desc"`
	URL          string                 `json:"url"`
	Icon         IconRequest            `json:"icon"`
	AuthType     int                    `json:"auth_type"`
	OAuthInfo    string                 `json:"oauth_info"`
	CommonParams map[string][]ParamItem `json:"common_params"`
	EditVersion  int                    `json:"edit_version"`
	PluginType   int                    `json:"plugin_type"`
	BriefIntro   string                 `json:"brief_intro"`
}

// 定义响应结构
type UpdatePluginResponse struct {
	Code        int    `json:"code"`
	EditVersion int    `json:"edit_version"`
	Msg         string `json:"msg"`
}

func UpdatePlugin(reqData *UpdatePluginRequest, cookie string) (*UpdatePluginResponse, error) {
	// 序列化请求数据
	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return nil, fmt.Errorf("marshal request data failed: %v", err)
	}

	// 创建 HTTP 请求
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/api/plugin_api/update_plugin_meta", host),
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)

	// 发送请求
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
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

	// 解析响应
	var result UpdatePluginResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("unmarshal JSON response body failed: %v, body: %s", err, string(body))
	}
	if result.Code != 0 {
		return nil, fmt.Errorf("update plugin failed: %v", result.Msg)
	}

	return &result, nil
}

// 请求体结构
type RefreshRequest struct {
	PluginID string `json:"plugin_id"`
	SpaceID  string `json:"space_id"`
}

// 响应体结构
type RefreshResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// RefreshToolList 刷新工具列表
func RefreshToolList(cookie string, pluginID string, spaceID string) (*RefreshResponse, error) {
	// 构建请求URL
	url := fmt.Sprintf("%s/api/plugin_api/refresh_mcp_plugin_apis", host)

	// 构建请求体
	requestBody := RefreshRequest{
		PluginID: pluginID,
		SpaceID:  spaceID,
	}

	// 将请求体转换为JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %v", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("agw-js-conv", "str")
	req.Header.Set("Cookie", cookie)

	// 创建HTTP客户端并发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// 解析响应
	var response RefreshResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}
	if response.Code != 0 {
		return nil, fmt.Errorf("refresh plugin failed: %v", response.Msg)
	}

	return &response, nil
}

// 发布请求结构体
type PublishRequest struct {
	PluginID      string `json:"plugin_id"`
	PrivacyStatus bool   `json:"privacy_status"`
	PrivacyInfo   string `json:"privacy_info"`
	VersionName   string `json:"version_name"`
	VersionDesc   string `json:"version_desc"`
}

// 发布响应结构体
type PublishResponse struct {
	Code      int    `json:"code"`
	Msg       string `json:"msg"`
	VersionTS string `json:"version_ts"`
}

// 发布插件
func PublishPlugin(cookie string, req *PublishRequest) (*PublishResponse, error) {
	// 构建完整的URL
	url := fmt.Sprintf("%s/api/plugin_api/publish_plugin", host)

	// 序列化请求体
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Cookie", cookie)

	// 发送请求
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// 解析响应
	var publishResp PublishResponse
	if err := json.Unmarshal(body, &publishResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	if publishResp.Code != 0 {
		return nil, fmt.Errorf("publish plugin failed: %v", publishResp.Msg)
	}

	return &publishResp, nil
}
