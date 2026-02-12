package adaptor

import (
	"context"
	"io"
	"net/http"

	"github.com/kymo-mcp/mcpcan/pkg/llm2/meta"
	"github.com/kymo-mcp/mcpcan/pkg/llm2/model"
)

type GinContext interface{}

type Adaptor interface {
	Init(meta *meta.Meta)
	GetRequestURL(meta *meta.Meta) (string, error)
	SetupRequestHeader(ctx context.Context, req *http.Request, meta *meta.Meta) error
	ConvertRequest(ctx context.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error)
	ConvertImageRequest(request *model.ImageRequest) (any, error)
	DoRequest(ctx context.Context, meta *meta.Meta, requestBody io.Reader) (*http.Response, error)
	DoResponse(ctx context.Context, c GinContext, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode)
	GetModelList() []string
	GetChannelName() string
}
