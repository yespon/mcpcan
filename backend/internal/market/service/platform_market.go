package service

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	pm "github.com/kymo-mcp/mcpcan/api/market/platform_market"
	"github.com/kymo-mcp/mcpcan/pkg/common"
	i18nresp "github.com/kymo-mcp/mcpcan/pkg/i18n"
	"github.com/kymo-mcp/mcpcan/pkg/mcpcansaas"
)

// PlatformMarketService provides platform market service functionality
type PlatformMarketService struct {
	client *mcpcansaas.Client
}

// NewPlatformMarketService creates a new PlatformMarketService instance
func NewPlatformMarketService() *PlatformMarketService {
	client, err := mcpcansaas.NewClient()
	if err != nil {
		fmt.Printf("Failed to connect to platform market service: %v\n", err)
		return nil
	}

	return &PlatformMarketService{
		client: client,
	}
}

// Close closes the gRPC connection
func (s *PlatformMarketService) Close() {
	if s.client != nil {
		s.client.Close()
	}
}

// ListMcpServer retrieves a list of MCP servers from the platform market
func (s *PlatformMarketService) ListMcpServer(c *gin.Context) {
	// Bind request parameters
	var req pm.ListMcpServerRequest
	if err := common.BindAndValidateUniversal(c, &req); err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("parameter validation failed: %v", err))
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// page default value
	if req.Page <= 0 {
		req.Page = 1
	}

	// page_size default value
	if req.PageSize <= 0 {
		req.PageSize = 100
	}

	// Call remote gRPC service via mcpcansaas client
	rpcReq := &pm.ListMcpServerRequest{
		Page:         req.Page,
		PageSize:     req.PageSize,
		Name:         req.Name,
		CategoryName: req.CategoryName,
	}

	resp, err := s.client.ListMcpServer(ctx, rpcReq)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to call platform market API: %v", err))
		return
	}

	// Convert response
	list := make([]*pm.McpServer, 0, len(resp.List))
	for _, item := range resp.List {
		categories := make([]*pm.McpServer_Category, 0, len(item.CategoryIds))
		for _, cat := range item.CategoryIds {
			categories = append(categories, &pm.McpServer_Category{
				Id:   cat.Id,
				Name: cat.Name,
				Code: cat.Code,
			})
		}

		list = append(list, &pm.McpServer{
			Id:                          item.Id,
			Name:                        item.Name,
			NameEn:                      item.NameEn,
			CategoryIds:                 categories,
			Description:                 item.Description,
			DescriptionEn:               item.DescriptionEn,
			Status:                      item.Status,
			PublishTime:                 item.PublishTime,
			CreateTime:                  item.CreateTime,
			UpdateTime:                  item.UpdateTime,
			ConfigTemplate:              item.ConfigTemplate,
			DeployMode:                  item.DeployMode,
			McpProtocol:                 item.McpProtocol,
			ImageUrl:                    item.ImageUrl,
			InitScript:                  item.InitScript,
			Tags:                        item.Tags,
			GithubStargazersCount:       item.GithubStargazersCount,
			GithubWatchersCount:         item.GithubWatchersCount,
			GithubForksCount:            item.GithubForksCount,
			GithubLicenseName:           item.GithubLicenseName,
			GithubReadme:                item.GithubReadme,
			GithubDefaultBranch:         item.GithubDefaultBranch,
			GithubDefaultBranchLastTime: item.GithubDefaultBranchLastTime,
			GithubOwner:                 item.GithubOwner,
			GithubOwnerAvatarUrl:        item.GithubOwnerAvatarUrl,
			GithubRepoUrl:               item.GithubRepoUrl,
		})
	}

	reply := &pm.ListMcpServerReply{
		Total: resp.Total,
		List:  list,
	}

	common.GinSuccess(c, reply)
}
