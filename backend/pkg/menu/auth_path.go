package menu

import (
	_ "embed"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed auth_path.yaml
var authPathFile string

var serviceAuthPath []*ServiceAuthPath

type AuthPath struct {
	Path       string   `json:"path"`
	Method     string   `json:"method"`
	Permission []string `json:"permission"`
}

type ServiceAuthPath struct {
	Name  string
	Auths []*AuthPath
}

func init() {
	err := yaml.Unmarshal([]byte(authPathFile), &serviceAuthPath)
	if err != nil {
		panic("menu.yaml unmarshal failed: " + err.Error())
	}
}

func GetPathPermission(path string, method string) []string {
	// 遍历serviceAuthPath，找到path匹配的AuthPath
	for _, service := range serviceAuthPath {
		for _, auth := range service.Auths {
			if auth.Path == path && auth.Method == method {
				return auth.Permission
			}
		}
	}
	return nil
}

// MatchPathPermission 当 c.FullPath() 返回空字符串时（Gin 路由冲突导致），
// 使用实际请求路径与已注册的 auth path 模式进行匹配。
// 将 :param 段视为通配符，可匹配任意路径段。
func MatchPathPermission(requestPath string, method string) []string {
	for _, service := range serviceAuthPath {
		for _, auth := range service.Auths {
			if auth.Method != method {
				continue
			}
			if matchPath(auth.Path, requestPath) {
				return auth.Permission
			}
		}
	}
	return nil
}

// matchPath 将注册的路由模式（如 /market/ai/sessions/:id/chat）
// 与实际请求路径（如 /market/ai/sessions/123/chat）进行匹配。
// :param 段匹配任意非空路径段。
func matchPath(pattern, requestPath string) bool {
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")
	requestParts := strings.Split(strings.Trim(requestPath, "/"), "/")

	if len(patternParts) != len(requestParts) {
		return false
	}

	for i, part := range patternParts {
		if strings.HasPrefix(part, ":") {
			// 参数段，匹配任意非空值
			if requestParts[i] == "" {
				return false
			}
			continue
		}
		if part != requestParts[i] {
			return false
		}
	}
	return true
}
