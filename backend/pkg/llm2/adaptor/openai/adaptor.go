package openai

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/kymo-mcp/mcpcan/pkg/llm2/adaptor"
	"github.com/kymo-mcp/mcpcan/pkg/llm2/channeltype"
	"github.com/kymo-mcp/mcpcan/pkg/llm2/meta"
	"github.com/kymo-mcp/mcpcan/pkg/llm2/model"
	"github.com/kymo-mcp/mcpcan/pkg/llm2/relaymode"
)

type Adaptor struct {
	ChannelType int
}

func (a *Adaptor) Init(meta *meta.Meta) {
	a.ChannelType = meta.ChannelType
}

func (a *Adaptor) GetRequestURL(meta *meta.Meta) (string, error) {
	// 根据不同的渠道类型返回不同的 URL
	switch meta.ChannelType {
	case channeltype.Azure:
		// Azure OpenAI 需要特定的 URL 格式
		requestURL := strings.Split(meta.RequestURLPath, "?")[0]
		modelName := meta.ActualModelName
		modelName = strings.Replace(modelName, ".", "", -1)
		// {your endpoint}/openai/deployments/{your azure_model}/chat/completions?api-version={api_version}
		requestURL = fmt.Sprintf("/openai/deployments/%s/%s", modelName, strings.TrimPrefix(requestURL, "/v1/"))
		return fmt.Sprintf("%s%s", meta.BaseURL, requestURL), nil
	default:
		// 默认 OpenAI 格式
		return fmt.Sprintf("%s%s", meta.BaseURL, meta.RequestURLPath), nil
	}
}

func (a *Adaptor) SetupRequestHeader(ctx context.Context, req *http.Request, meta *meta.Meta) error {
	req.Header.Set("Content-Type", "application/json")

	if meta.ChannelType == channeltype.Azure {
		req.Header.Set("api-key", meta.APIKey)
		return nil
	}

	req.Header.Set("Authorization", "Bearer "+meta.APIKey)
	return nil
}

func (a *Adaptor) ConvertRequest(ctx context.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, fmt.Errorf("request is nil")
	}

	// 对于流式请求，始终返回用量
	if request.Stream {
		if request.StreamOptions == nil {
			request.StreamOptions = &model.StreamOptions{}
		}
		request.StreamOptions.IncludeUsage = true
	}

	return request, nil
}

func (a *Adaptor) ConvertImageRequest(request *model.ImageRequest) (any, error) {
	if request == nil {
		return nil, fmt.Errorf("request is nil")
	}
	return request, nil
}

func (a *Adaptor) DoRequest(ctx context.Context, meta *meta.Meta, requestBody io.Reader) (*http.Response, error) {
	// 创建 HTTP 请求
	requestURL, err := a.GetRequestURL(meta)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, requestBody)
	if err != nil {
		return nil, err
	}

	// 设置请求头
	err = a.SetupRequestHeader(ctx, httpReq, meta)
	if err != nil {
		return nil, err
	}

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (a *Adaptor) DoResponse(ctx context.Context, c adaptor.GinContext, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	if meta.IsStream {
		// 流式响应处理
		usage, err = a.handleStreamResponse(c, resp, meta.Mode)
	} else {
		// 非流式响应处理
		switch meta.Mode {
		case relaymode.ImagesGenerations:
			// 图片生成响应处理
			err = a.handleImageResponse(c, resp)
		default:
			// 一般响应处理
			usage, err = a.handleTextResponse(c, resp, meta.PromptTokens, meta.ActualModelName)
		}
	}
	return
}

func (a *Adaptor) handleStreamResponse(c adaptor.GinContext, resp *http.Response, relayMode int) (*model.Usage, *model.ErrorWithStatusCode) {
	// 这里应该实现流式响应处理逻辑
	// 由于我们暂时不直接依赖 gin，我们返回基本结构
	return nil, nil
}

func (a *Adaptor) handleImageResponse(c adaptor.GinContext, resp *http.Response) *model.ErrorWithStatusCode {
	// 图片响应处理逻辑
	return nil
}

func (a *Adaptor) handleTextResponse(c adaptor.GinContext, resp *http.Response, promptTokens int, modelName string) (*model.Usage, *model.ErrorWithStatusCode) {
	// 文本响应处理逻辑
	return nil, nil
}

func (a *Adaptor) GetModelList() []string {
	// 返回支持的模型列表
	return []string{
		"gpt-4",
		"gpt-4o",
		"gpt-4o-mini",
		"gpt-3.5-turbo",
		"text-embedding-ada-002",
		"davinci-002",
		"babbage-002",
	}
}

func (a *Adaptor) GetChannelName() string {
	switch a.ChannelType {
	case channeltype.Azure:
		return "azure"
	default:
		return "openai"
	}
}
