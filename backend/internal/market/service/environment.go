package service

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/kymo-mcp/mcpcan/api/market/mcp_environment"
	"github.com/kymo-mcp/mcpcan/internal/market/biz"
	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	i18nresp "github.com/kymo-mcp/mcpcan/pkg/i18n"
)

// EnvironmentService provides environment management functionality
type EnvironmentService struct {
}

// NewEnvironmentService creates a new EnvironmentService instance
func NewEnvironmentService() *EnvironmentService {
	return &EnvironmentService{}
}

// modelToMcpEnvironmentInfo converts model to MCP environment info
func modelToMcpEnvironmentInfo(env *model.McpEnvironment) *mcp_environment.McpEnvironmentInfo {
	configStr := env.Config
	return &mcp_environment.McpEnvironmentInfo{
		Id:          int32(env.ID),
		Name:        env.Name,
		Environment: string(env.Environment),
		Config:      configStr,
		Namespace:   env.Namespace,
		CreatedAt:   env.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   env.UpdatedAt.Format(time.RFC3339),
	}
}

// modelToEnvironmentResponse converts model to environment response
func modelToEnvironmentResponse(env *model.McpEnvironment) *mcp_environment.EnvironmentResponse {
	var envType mcp_environment.McpEnvironmentType
	switch env.Environment {
	case model.McpEnvironmentKubernetes:
		envType = mcp_environment.McpEnvironmentType_Kubernetes
	case model.McpEnvironmentDocker:
		envType = mcp_environment.McpEnvironmentType_Docker
	default:
		envType = mcp_environment.McpEnvironmentType_Kubernetes
	}

	return &mcp_environment.EnvironmentResponse{
		Id:          int32(env.ID),
		Name:        env.Name,
		Environment: envType,
		Config:      env.Config,
		Namespace:   env.Namespace,
		CreatedAt:   env.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   env.UpdatedAt.Format(time.RFC3339),
	}
}

// ListEnvironmentsHandler environment list interface Handler
func (s *EnvironmentService) ListEnvironmentsHandler(c *gin.Context) {
	var req mcp_environment.ListEnvironmentsRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// Set default pagination parameters
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100 // Limit maximum page size
	}

	var environments []*model.McpEnvironment
	var total int64
	var err error

	// Query based on filter conditions
	var envType model.McpEnvironmentType
	switch req.Environment {
	case mcp_environment.McpEnvironmentType_Kubernetes:
		envType = model.McpEnvironmentKubernetes
		environments, total, err = biz.GEnvironmentBiz.ListEnvironmentsByTypeWithPagination(c.Request.Context(), envType, int(req.Page), int(req.PageSize))
	case mcp_environment.McpEnvironmentType_Docker:
		envType = model.McpEnvironmentDocker
		environments, total, err = biz.GEnvironmentBiz.ListEnvironmentsByTypeWithPagination(c.Request.Context(), envType, int(req.Page), int(req.PageSize))
	default:
		// Query all environments
		environments, total, err = biz.GEnvironmentBiz.ListEnvironmentsWithPagination(c.Request.Context(), int(req.Page), int(req.PageSize))
	}

	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to query environment list: %s", err.Error()))
		return
	}

	// Build response list
	var responseList []*mcp_environment.McpEnvironmentInfo
	for _, env := range environments {
		responseList = append(responseList, modelToMcpEnvironmentInfo(env))
	}

	response := &mcp_environment.ListEnvironmentsResponse{
		List:     responseList,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	common.GinSuccess(c, response)
}

// TestConnectivityHandler connectivity test interface Handler
func (s *EnvironmentService) TestConnectivityHandler(c *gin.Context) {
	// Get ID from URL path parameter
	idStr := c.Param("id")
	if idStr == "" {
		common.GinError(c, i18nresp.CodeInternalError, "environment ID cannot be empty")
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		fmt.Printf("Error: Received invalid environment ID string: '%s'\n", idStr)
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("invalid environment ID: %s", idStr))
		return
	}

	// Get environment information
	environment, err := biz.GEnvironmentBiz.GetEnvironment(c.Request.Context(), uint(id))
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to query environment: %s", err.Error()))
		return
	}
	if environment == nil {
		common.GinError(c, i18nresp.CodeInternalError, "environment does not exist")
		return
	}

	// Execute connectivity test
	result, err := biz.GEnvironmentBiz.TestEnvironmentConnectivity(c.Request.Context(), environment)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("connectivity test failed: %s", err.Error()))
		return
	}

	common.GinSuccess(c, result)
}

// ListNamespacesHandler get namespace list Handler
func (s *EnvironmentService) ListNamespacesHandler(c *gin.Context) {
	// Bind request parameters
	var req mcp_environment.ListNamespacesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("parameter binding failed: %s", err.Error()))
		return
	}

	if req.Config == "" {
		common.GinError(c, i18nresp.CodeInternalError, "config parameter cannot be empty")
		return
	}

	// Parse environment type
	var environmentType model.McpEnvironmentType
	switch req.Environment {
	case mcp_environment.McpEnvironmentType_Kubernetes:
		environmentType = model.McpEnvironmentKubernetes
	case mcp_environment.McpEnvironmentType_Docker:
		common.GinError(c, i18nresp.CodeInternalError, "docker environment type is not supported")
		return
	default:
		common.GinError(c, i18nresp.CodeInternalError, "unsupported environment type")
		return
	}

	// Call business logic
	namespaces, err := biz.GEnvironmentBiz.ListNamespaces(c.Request.Context(), req.Config, environmentType)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	// Build response
	response := &mcp_environment.ListNamespacesResponse{
		List: namespaces,
	}

	common.GinSuccess(c, response)
}
