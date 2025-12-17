package biz

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/kymo-mcp/mcpcan/pkg/i18n"
	"github.com/kymo-mcp/mcpcan/pkg/logger"
	"go.uber.org/zap"
)

type KymoAuthUser struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	NickName string `json:"nickName"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Avatar   string `json:"avatarPath"`
	Dept     struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	} `json:"dept"`
	Roles []struct {
		ID    int64  `json:"id"`
		Name  string `json:"name"`
		Level int    `json:"level"`
	} `json:"roles"`
}

type KymoAuthResponse struct {
	User  *KymoAuthUser `json:"user"`
	Roles []string      `json:"roles"`
}

type KymoError struct {
	Code    int
	Message string
	Data    interface{}
}

const (
	// dev url := "https://ai-dev.itqm.cn/intelligent-api/auth/info"
	// prod url := "http://intelligent-api-svc/intelligent-api/auth/info"
	KymoAuthURL = "http://intelligent-api-svc/intelligent-api/auth/info"
)

func mapHTTPToI18nCode(status int) int {
	switch status {
	case http.StatusBadRequest:
		return i18n.CodeBadRequest
	case http.StatusUnauthorized:
		return i18n.CodeUnauthorized
	case http.StatusForbidden:
		return i18n.CodeForbidden
	case http.StatusNotFound:
		return i18n.CodeNotFound
	case http.StatusMethodNotAllowed:
		return i18n.CodeMethodNotAllowed
	case http.StatusRequestTimeout:
		return i18n.CodeRequestTimeout
	case http.StatusTooManyRequests:
		return i18n.CodeTooManyRequests
	case http.StatusInternalServerError:
		return i18n.CodeInternalError
	case http.StatusBadGateway, http.StatusServiceUnavailable:
		return i18n.CodeServiceUnavailable
	case http.StatusGatewayTimeout:
		return i18n.CodeGatewayTimeout
	default:
		return i18n.CodeServiceError
	}
}

func (uc *AuthUseBiz) ValidateTokenExternalKymo(ctx context.Context, token string, headers map[string]string, cookies []*http.Cookie) (*ValidateResult, *KymoError, int) {
	req, _ := http.NewRequest("GET", KymoAuthURL, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	for k, v := range headers {
		if v != "" {
			req.Header.Set(k, v)
		}
	}
	if _, ok := req.Header["Accept"]; !ok {
		req.Header.Set("Accept", "application/json")
	}
	for _, ck := range cookies {
		req.AddCookie(ck)
	}

	logger.Info("ValidateTokenExternalKymo Request", zap.String("url", KymoAuthURL), zap.Any("headers", req.Header))

	cli := &http.Client{Timeout: 5 * time.Second}
	resp, err := cli.Do(req)
	if err != nil {
		return nil, &KymoError{Code: i18n.CodeNetworkError, Message: err.Error(), Data: nil}, http.StatusBadGateway
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	logger.Info("ValidateTokenExternalKymo Response", zap.String("url", KymoAuthURL), zap.Any("headers", resp.Header), zap.ByteString("body", body))

	if resp.StatusCode != http.StatusOK {
		var upstream struct {
			Timestamp string `json:"timestamp"`
			Status    int    `json:"status"`
			Error     string `json:"error"`
			Message   string `json:"message"`
			Path      string `json:"path"`
		}
		if json.Unmarshal(body, &upstream) == nil {
			msg := upstream.Message
			if msg == "" {
				msg = upstream.Error
			}
			code := mapHTTPToI18nCode(resp.StatusCode)
			return nil, &KymoError{Code: code, Message: msg, Data: upstream}, resp.StatusCode
		}
		code := mapHTTPToI18nCode(resp.StatusCode)
		return nil, &KymoError{Code: code, Message: "Upstream error", Data: json.RawMessage(body)}, resp.StatusCode
	}

	var parsed KymoAuthResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, &KymoError{Code: i18n.CodeParseError, Message: err.Error(), Data: json.RawMessage(body)}, http.StatusBadGateway
	}
	if parsed.User == nil {
		return nil, &KymoError{Code: i18n.CodeInvalidToken, Message: "user is nil", Data: parsed}, http.StatusUnauthorized
	}

	roleIDs := []uint{}
	roleNames := []string{}
	for _, r := range parsed.User.Roles {
		roleIDs = append(roleIDs, uint(r.ID))
		roleNames = append(roleNames, r.Name)
	}

	userInfo := &UserInfo{
		UserID:    parsed.User.ID,
		Username:  parsed.User.Username,
		Nickname:  parsed.User.NickName,
		Email:     parsed.User.Email,
		Phone:     parsed.User.Phone,
		Avatar:    parsed.User.Avatar,
		DeptID:    parsed.User.Dept.ID,
		DeptName:  parsed.User.Dept.Name,
		RoleIDs:   roleIDs,
		RoleNames: roleNames,
	}

	result := &ValidateResult{
		Valid:     true,
		UserInfo:  userInfo,
		LoginInfo: nil,
	}
	return result, nil, http.StatusOK
}
