package menu

import (
	_ "embed"

	"gopkg.in/yaml.v3"
)

//go:embed auth_path.yaml
var authPathFile string

var serviceAuthPath []*ServiceAuthPath

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
