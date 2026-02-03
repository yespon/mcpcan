package menu

import (
	_ "embed"

	"gopkg.in/yaml.v3"
)

//go:embed menu.yaml
var menuFile string

var menus []*Menu

func GetMenus() []*Menu {
	return menus
}

type Menu struct {
	Permission string  `json:"permission"`
	Title      string  `json:"title"`
	Type       int     `json:"type"`
	Sort       int     `json:"sort"`
	Path       string  `json:"path"`
	EngTitle   string  `json:"engTitle"`
	Children   []*Menu `json:"children"`
}

func init() {
	err := yaml.Unmarshal([]byte(menuFile), &menus)
	if err != nil {
		panic("menu.yaml unmarshal failed: " + err.Error())
	}
}
