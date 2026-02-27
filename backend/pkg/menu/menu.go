package menu

import (
	_ "embed"
	"strings"

	"github.com/kymo-mcp/mcpcan/pkg/common"
	"gopkg.in/yaml.v3"
)

//go:embed menu.yaml
var menuFile string

var menus []*Menu

func GetMenus(codeModel common.CodeMode) []*Menu {
	var openCodeMenus []*Menu
	if codeModel != common.EnterpriseCodeCodeMode {
		for _, menu := range menus {
			if strings.Contains(menu.Permission, "mcpcan_rbac_manage") || strings.Contains(menu.Permission, "data_permission") {
				continue
			}
			openCodeMenus = append(openCodeMenus, menu)
		}
		return openCodeMenus
	}
	return menus
}

type Menu struct {
	Permission string  `json:"permission"`
	Title      string  `json:"title"`
	Type       int     `json:"type"`
	Sort       int     `json:"sort"`
	Path       string  `json:"path"`
	EngTitle   string  `json:"engTitle" yaml:"engTitle"`
	Children   []*Menu `json:"children"`
}

func init() {
	err := yaml.Unmarshal([]byte(menuFile), &menus)
	if err != nil {
		panic("menu.yaml unmarshal failed: " + err.Error())
	}
}
