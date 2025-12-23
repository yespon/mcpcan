package coze

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const host = "https://www.coze.cn"

func GetOrganizationID(cookie string, enterpriseID string) (string, error) {
	// 准备请求
	reqBody := map[string]interface{}{
		"joined":                false,
		"need_people_number":    true,
		"need_workspace_number": true,
		"page":                  1,
		"page_size":             20,
	}
	if enterpriseID != "" {
		reqBody["enterprise_id"] = enterpriseID
	}

	jsonBody, _ := json.Marshal(reqBody)

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/api/permission_api/enterprise/search_organization", host),
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("agw-js-conv", "str")
	req.Header.Set("Cookie", cookie)

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// 使用map解析，更灵活
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	// 检查响应码
	if code, ok := result["code"].(float64); ok && code != 0 {
		return "", fmt.Errorf("API error: %v", result["msg"])
	}

	// 提取organization_id
	data, _ := result["data"].(map[string]interface{})
	orgList, _ := data["organization_list"].([]interface{})

	if len(orgList) > 0 {
		firstOrg, _ := orgList[0].(map[string]interface{})
		orgID, _ := firstOrg["organization_id"].(string)
		return orgID, nil
	}

	return "", fmt.Errorf("not found organization id")
}

type SpaceListRequest struct {
	EnterpriseID   string `json:"enterprise_id,omitempty"`
	OrganizationID string `json:"organization_id,omitempty"`
	Page           int    `json:"page"`
	PageSize       int    `json:"page_size"`
}

type SpaceListResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		BotSpaceList []struct {
			ID            string `json:"id"`
			Name          string `json:"name"`
			OwnerUserID   string `json:"owner_user_id"`
			OwnerUserName string `json:"owner_user_name"`
			RoleType      int    `json:"role_type"`
			SpaceRoleType int    `json:"space_role_type"`
			// 其他字段可以根据需要添加
			//Connectors []struct {
			//	Icon string `json:"icon"`
			//	ID   string `json:"id"`
			//	Name string `json:"name"`
			//} `json:"connectors"`
			//Description    string `json:"description"`
			//IconURL        string `json:"icon_url"`
			//OrganizationID string `json:"organization_id"`
			//EnterpriseID   string `json:"enterprise_id"`
		} `json:"bot_space_list"`
		HasMore          bool `json:"has_more"`
		HasPersonalSpace bool `json:"has_personal_space"`
		MaxTeamSpaceNum  int  `json:"max_team_space_num"`
		TeamSpaceNum     int  `json:"team_space_num"`
		Total            int  `json:"total"`
	} `json:"data"`
}

type SpaceInfo struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	OwnerUserID   string `json:"owner_user_id"`
	OwnerUserName string `json:"owner_user_name"`
	RoleType      int    `json:"role_type"`
	SpaceRoleType int    `json:"space_role_type"`
}

func GetSpaceList(cookie, enterpriseID, organizationID string) ([]SpaceInfo, error) {
	var allSpaceInfos []SpaceInfo
	page := 1
	const pageSize = 20

	for {
		// 准备请求体
		requestBody := SpaceListRequest{
			Page:     page,
			PageSize: pageSize,
		}
		if enterpriseID != "" {
			requestBody.EnterpriseID = enterpriseID
		}
		if organizationID != "" {
			requestBody.OrganizationID = organizationID
		}

		jsonBody, err := json.Marshal(requestBody)
		if err != nil {
			return nil, fmt.Errorf("序列化请求体失败: %v", err)
		}

		// 创建HTTP请求
		req, err := http.NewRequest(
			"POST",
			fmt.Sprintf("%s/api/playground_api/space/list", host),
			bytes.NewBuffer(jsonBody),
		)
		if err != nil {
			return nil, fmt.Errorf("创建请求失败: %v", err)
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
		var response SpaceListResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("unmarshal JSON response body failed: %v", err)
		}

		// 检查响应状态
		if response.Code != 0 {
			return nil, fmt.Errorf("API result error: %s", response.Msg)
		}

		// 提取所需字段
		for _, space := range response.Data.BotSpaceList {
			spaceInfo := SpaceInfo{
				ID:            space.ID,
				Name:          space.Name,
				OwnerUserID:   space.OwnerUserID,
				OwnerUserName: space.OwnerUserName,
				RoleType:      space.RoleType,
				SpaceRoleType: space.SpaceRoleType,
			}
			if spaceInfo.OwnerUserName == "" {
				spaceInfo.OwnerUserName = "myself"
			}
			allSpaceInfos = append(allSpaceInfos, spaceInfo)
		}

		if len(allSpaceInfos) >= response.Data.Total {
			break
		}
		if page*pageSize > response.Data.Total {
			break
		}
		page++
	}

	return allSpaceInfos, nil
}
