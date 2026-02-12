package model

// ImageRequest 图像请求结构
type ImageRequest struct {
	Prompt         string `json:"prompt"`
	Model          string `json:"model"`
	N              int    `json:"n"`
	Size           string `json:"size"`
	Quality        string `json:"quality"`
	ResponseFormat string `json:"response_format"`
	Style          string `json:"style"`
	User           string `json:"user"`
}
