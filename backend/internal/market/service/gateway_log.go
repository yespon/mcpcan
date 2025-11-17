package service

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/fatedier/golib/log"
	"github.com/gin-gonic/gin"
	gatewaylogpb "github.com/kymo-mcp/mcpcan/api/market/gateway_log"
	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
	i18nresp "github.com/kymo-mcp/mcpcan/pkg/i18n"
)

type GatewayLogService struct{}

func NewGatewayLogService() *GatewayLogService { return &GatewayLogService{} }

func (s *GatewayLogService) FindHandler(c *gin.Context) {
	var req gatewaylogpb.FindGatewayLogRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	filters := map[string]interface{}{}
	if strings.TrimSpace(req.InstanceId) != "" {
		filters["instance_id"] = strings.TrimSpace(req.InstanceId)
	}
	if strings.TrimSpace(req.TokenHeader) != "" {
		filters["tokenHeader"] = strings.TrimSpace(req.TokenHeader)
	}
	if strings.TrimSpace(req.Token) != "" {
		filters["token"] = strings.TrimSpace(req.Token)
	}
	if req.Level != gatewaylogpb.Level(0) {
		filters["level"] = protoLevelToInt(req.Level)
	}
	if len(req.Usages) > 0 {
		filters["usages"] = req.Usages
	}
	if req.StartTime > 0 {
		filters["createdAtStart"] = time.Unix(req.StartTime, 0)
	}
	if req.EndTime > 0 {
		filters["createdAtEnd"] = time.Unix(req.EndTime, 0)
	}

	logs, total, err := mysql.GatewayLogRepo.FindWithPagination(context.Background(), page, pageSize, filters, "createdAt", "desc")
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	result := make([]*gatewaylogpb.GatewayLog, 0, len(logs))
	for _, g := range logs {
		var inner *gatewaylogpb.GatewayLog_Log
		if g.Log != nil {
			var ml model.Log
			_ = json.Unmarshal(g.Log, &ml)
			inner = &gatewaylogpb.GatewayLog_Log{
				Event:   string(ml.Event),
				Level:   intToProtoLevel(int(ml.Level)),
				Message: ml.Message,
				Url:     ml.URL,
				Method:  ml.Method,
				Path:    ml.Path,
				Params:  ml.Params,
				IsSSE:   ml.IsSSE,
				Ts:      ml.TS,
			}
		}
		result = append(result, &gatewaylogpb.GatewayLog{
			Id:          uint32(g.ID),
			InstanceId:  g.InstanceID,
			TokenHeader: g.TokenHeader,
			Token:       g.Token,
			Usages:      splitUsages(g.Usages),
			Level:       intToProtoLevel(int(g.Level)),
			Event:       string(g.Event),
			CreatedAt:   g.CreatedAt.Format(time.RFC3339Nano),
			Log:         inner,
		})
	}

	resp := &gatewaylogpb.FindGatewayLogResponse{
		Logs:     result,
		Total:    int32(total),
		PageSize: pageSize,
		Page:     page,
	}
	common.GinSuccess(c, resp)
}

func splitUsages(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return []string{}
	}
	arr := strings.Split(s, ",")
	out := make([]string, 0, len(arr))
	for _, it := range arr {
		if strings.TrimSpace(it) != "" {
			out = append(out, strings.TrimSpace(it))
		}
	}
	return out
}

func protoLevelToInt(l gatewaylogpb.Level) int {
	switch l {
	case gatewaylogpb.Level_TraceLevel:
		return int(log.DebugLevel)
	case gatewaylogpb.Level_DebugLevel:
		return int(log.DebugLevel)
	case gatewaylogpb.Level_InfoLevel:
		return int(log.InfoLevel)
	case gatewaylogpb.Level_WarnLevel:
		return int(log.WarnLevel)
	case gatewaylogpb.Level_ErrorLevel:
		return int(log.ErrorLevel)
	default:
		return int(log.InfoLevel)
	}
}

func intToProtoLevel(v int) gatewaylogpb.Level {
	switch v {
	case int(log.DebugLevel):
		return gatewaylogpb.Level_DebugLevel
	case int(log.InfoLevel):
		return gatewaylogpb.Level_InfoLevel
	case int(log.WarnLevel):
		return gatewaylogpb.Level_WarnLevel
	case int(log.ErrorLevel):
		return gatewaylogpb.Level_ErrorLevel
	default:
		return gatewaylogpb.Level_InfoLevel
	}
}
