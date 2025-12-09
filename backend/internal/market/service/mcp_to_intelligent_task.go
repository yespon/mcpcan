package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	iapb "github.com/kymo-mcp/mcpcan/api/market/intelligent_access"
	pb "github.com/kymo-mcp/mcpcan/api/market/mcp_to_intelligent_task"
	"github.com/kymo-mcp/mcpcan/internal/market/repository"
	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/postgres"
	"github.com/kymo-mcp/mcpcan/pkg/dify"
	i18nresp "github.com/kymo-mcp/mcpcan/pkg/i18n"
	"github.com/kymo-mcp/mcpcan/pkg/logger"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
	"go.uber.org/zap"
)

// McpToIntelligentTaskService struct for mcp to intelligent task service
type McpToIntelligentTaskService struct {
	ctx context.Context
}

// NewMcpToIntelligentTaskService creates a new mcp to intelligent task service
func NewMcpToIntelligentTaskService(ctx context.Context) *McpToIntelligentTaskService {
	go func() {
		// 程序启动的时候加载运行中的任务，启动执行
		tasks, _, err := mysql.McpToIntelligentTaskRepo.FindWithPagination(context.Background(), 1, 1, "", pb.McpToIntelligentTaskStatus_Running.String())
		if err != nil {
			logger.Error(fmt.Sprintf("failed to find mcp to intelligent tasks: %s", err.Error()))
			return
		}
		if len(tasks) > 0 {
			logger.Info(fmt.Sprintf("found mcp to intelligent tasks: %v", tasks))
			ProcessMcpToIntelligentTask(tasks[0].ID)
		}
	}()

	return &McpToIntelligentTaskService{
		ctx: ctx,
	}
}

// CreateHandler creates mcp to intelligent task HTTP handler function
func (s *McpToIntelligentTaskService) CreateHandler(c *gin.Context) {
	var req pb.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.GinError(c, i18nresp.CodeBadRequest, err.Error())
		return
	}
	// validate request
	if req.Desc == "" || req.IntelligentAccessID == 0 || len(req.InsertIntelligentInfos) == 0 || len(req.McpInstanceIDs) == 0 {
		common.GinError(c, i18nresp.CodeBadRequest, "invalid request")
		return
	}

	// 转换 InsertIntelligentInfos validate request
	var insertInfos model.InsertIntelligentInfos
	for _, info := range req.InsertIntelligentInfos {
		if info.DifySpaceID == "" || info.DifyUserID == "" || info.DifySpaceName == "" || info.DifyUserName == "" {
			common.GinError(c, i18nresp.CodeBadRequest, "invalid request")
			return
		}

		setInfo := &model.InsertIntelligentInfo{
			DifySpaceID:   info.DifySpaceID,
			DifyUserID:    info.DifyUserID,
			DifySpaceName: info.DifySpaceName,
			DifyUserName:  info.DifyUserName,
			Headers:       map[string]*model.HeaderInfo{},
		}
		for key, val := range info.Headers {
			setInfo.Headers[key] = &model.HeaderInfo{
				Token:   val.Token,
				Headers: val.Headers,
			}
		}
		insertInfos = append(insertInfos, setInfo)
	}

	_, total, err := mysql.McpToIntelligentTaskRepo.FindWithPagination(s.ctx, 1, 1, "", pb.McpToIntelligentTaskStatus_Running.String())
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to find mcp to intelligent tasks: %s", err.Error()))
		return
	}
	if total > 0 {
		common.GinError(c, i18nresp.CodeBadRequest, "running mcp sync task already exists")
		return
	}

	instances, err := mysql.McpInstanceRepo.FindByInstanceIDs(s.ctx, req.McpInstanceIDs)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to find mcp instances: %s", err.Error()))
		return
	}
	if len(instances) != len(req.McpInstanceIDs) {
		common.GinError(c, i18nresp.CodeBadRequest, "invalid request")
		return
	}

	// 获取智能体平台名称
	intelligentAccess, err := repository.IntelligentAccessRepo.FindByID(s.ctx, req.IntelligentAccessID)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to find intelligent access: %s", err.Error()))
		return
	}

	task := &model.McpToIntelligentTask{
		Desc:                   req.Desc,
		IntelligentAccessID:    req.IntelligentAccessID,
		IntelligentAccessName:  intelligentAccess.AccessName,
		InsertIntelligentInfos: insertInfos,
		McpInstanceIDs:         req.McpInstanceIDs,
		Status:                 pb.McpToIntelligentTaskStatus_Running.String(), // 默认状态为运行中
		InstallLogs:            model.InstallLogs{},
		Domain:                 req.Domain,
	}

	if err := mysql.McpToIntelligentTaskRepo.Create(s.ctx, task); err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to create mcp to intelligent task: %s", err.Error()))
		return
	}

	go func() {
		ProcessMcpToIntelligentTask(task.ID)
	}()
	common.GinSuccess(c, &pb.CreateResponse{
		Task: s.convertToPbTask(task),
	})
}

// CancelHandler cancels mcp to intelligent task HTTP handler function
func (s *McpToIntelligentTaskService) CancelHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		common.GinError(c, i18nresp.CodeBadRequest, fmt.Sprintf("invalid id: %s", err.Error()))
		return
	}

	// 先查找任务
	task, err := mysql.McpToIntelligentTaskRepo.FindByID(s.ctx, id)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to find mcp to intelligent task: %s", err.Error()))
		return
	}

	// 更新状态为取消
	task.Status = pb.McpToIntelligentTaskStatus_Cancel.String()

	if err := mysql.McpToIntelligentTaskRepo.Update(s.ctx, task); err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to cancel mcp to intelligent task: %s", err.Error()))
		return
	}

	common.GinSuccess(c, &pb.CancelResponse{})
}

// GetHandler finds mcp to intelligent task by ID HTTP handler function
func (s *McpToIntelligentTaskService) GetHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		common.GinError(c, i18nresp.CodeBadRequest, fmt.Sprintf("invalid id: %s", err.Error()))
		return
	}

	task, err := mysql.McpToIntelligentTaskRepo.FindByID(s.ctx, id)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to find mcp to intelligent task: %s", err.Error()))
		return
	}

	common.GinSuccess(c, &pb.GetResponse{
		Task: s.convertToPbTask(task),
	})
}

// ListHandler finds all mcp to intelligent tasks with pagination HTTP handler function
func (s *McpToIntelligentTaskService) ListHandler(c *gin.Context) {
	var req pb.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.GinError(c, i18nresp.CodeBadRequest, err.Error())
		return
	}

	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 {
		req.Size = 10
	}

	tasks, total, err := mysql.McpToIntelligentTaskRepo.FindWithPagination(s.ctx, int(req.Page), int(req.Size), req.Keyword, "")
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to find mcp to intelligent tasks: %s", err.Error()))
		return
	}

	var pbTasks []*pb.McpToIntelligentTask
	for _, task := range tasks {
		pbTasks = append(pbTasks, s.convertToPbTask(task))
	}

	common.GinSuccess(c, &pb.ListResponse{
		List:  pbTasks,
		Total: total,
	})
}

// convertToPbTask converts model task to protobuf task
func (s *McpToIntelligentTaskService) convertToPbTask(task *model.McpToIntelligentTask) *pb.McpToIntelligentTask {
	// 转换 InsertIntelligentInfos
	var pbInsertInfos []*pb.InsertIntelligentInfo
	for _, info := range task.InsertIntelligentInfos {
		pbInsertInfos = append(pbInsertInfos, &pb.InsertIntelligentInfo{
			DifySpaceID:   info.DifySpaceID,
			DifyUserID:    info.DifyUserID,
			DifySpaceName: info.DifySpaceName,
			DifyUserName:  info.DifyUserName,
		})
	}

	// 转换 InstallLogs
	var pbInstallLogs []*pb.InstallLog
	for _, log := range task.InstallLogs {
		var pbInsertIntelligentLogs []*pb.InsertIntelligentLog
		for _, insertLog := range log.InsertIntelligentLogs {
			pbInsertIntelligentLogs = append(pbInsertIntelligentLogs, &pb.InsertIntelligentLog{
				InsertIntelligentInfo: &pb.InsertIntelligentInfo{
					DifySpaceID:   insertLog.InsertIntelligentInfo.DifySpaceID,
					DifyUserID:    insertLog.InsertIntelligentInfo.DifyUserID,
					DifySpaceName: insertLog.InsertIntelligentInfo.DifySpaceName,
					DifyUserName:  insertLog.InsertIntelligentInfo.DifyUserName,
				},
				ErrorLog: insertLog.ErrorLog,
				Status:   insertLog.Status,
			})
		}

		pbInstallLogs = append(pbInstallLogs, &pb.InstallLog{
			McpInstanceID:         log.McpInstanceID,
			McpInstanceName:       log.McpInstanceName,
			InsertIntelligentLogs: pbInsertIntelligentLogs,
			ErrorLog:              log.ErrorLog,
			Status:                log.Status,
		})
	}

	// 转换状态
	var status pb.McpToIntelligentTaskStatus
	switch task.Status {
	case pb.McpToIntelligentTaskStatus_Running.String():
		status = pb.McpToIntelligentTaskStatus_Running
	case pb.McpToIntelligentTaskStatus_Success.String():
		status = pb.McpToIntelligentTaskStatus_Success
	case pb.McpToIntelligentTaskStatus_Failed.String():
		status = pb.McpToIntelligentTaskStatus_Failed
	case pb.McpToIntelligentTaskStatus_Cancel.String():
		status = pb.McpToIntelligentTaskStatus_Cancel
	default:
		status = pb.McpToIntelligentTaskStatus_Unknown
	}

	return &pb.McpToIntelligentTask{
		Id:                     task.ID,
		Desc:                   task.Desc,
		IntelligentAccessID:    task.IntelligentAccessID,
		IntelligentAccessName:  task.IntelligentAccessName,
		InsertIntelligentInfos: pbInsertInfos,
		McpInstanceIDs:         task.McpInstanceIDs,
		Status:                 status,
		InstallLogs:            pbInstallLogs,
		CreatedAt:              task.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:              task.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func ProcessMcpToIntelligentTask(id int64) {
	logger.Info("process mcp to intelligent task service", zap.Int64("taskId", id))
	defer logger.Info("process mcp to intelligent task service: finished", zap.Int64("taskId", id))

	// 从数据库中获取任务
	task, err := mysql.McpToIntelligentTaskRepo.FindByID(context.Background(), id)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to find mcp to intelligent task: %s", err.Error()), zap.Int64("taskId", id))
		return
	}
	// 从数据库中获取智能体访问信息
	intelligentAccess, err := repository.IntelligentAccessRepo.FindByID(context.Background(), task.IntelligentAccessID)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to find intelligent access: %s", err.Error()), zap.Int64("IntelligentAccessID", task.IntelligentAccessID))
		return
	}
	conn, err := BuildTemporaryPostgresConnection(intelligentAccess.DbHost, intelligentAccess.DbPort, intelligentAccess.DbUser, intelligentAccess.DbPassword, intelligentAccess.DbName, false)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to build temporary postgres connection: %s", err.Error()), zap.Int64("IntelligentAccessID", task.IntelligentAccessID))
		return
	}
	defer conn.Close()
	userSpaces, err := GetDifyUserSpace(conn)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to get dify user space: %s", err.Error()), zap.Int64("IntelligentAccessID", task.IntelligentAccessID))
		return
	}

	var logs = []*model.InstallLog{}
	var lastStatus = pb.McpToIntelligentTaskStatus_Success.String()

	var wg sync.WaitGroup
	var mu sync.Mutex
	concurrency := 10
	sem := make(chan struct{}, concurrency)

	for _, instanceID := range task.McpInstanceIDs {
		// 获取实例信息
		mcpInstance, err := mysql.McpInstanceRepo.FindByInstanceID(context.Background(), instanceID)
		if err != nil {
			logs = append(logs, &model.InstallLog{
				McpInstanceID:   instanceID,
				McpInstanceName: "",
				Status:          false,
				ErrorLog:        fmt.Sprintf("failed to find mcpInstance: %s", err.Error()),
			})
			lastStatus = pb.McpToIntelligentTaskStatus_Failed.String()
			if err = mysql.McpToIntelligentTaskRepo.UpdateLogs(context.Background(), id, logs, pb.McpToIntelligentTaskStatus_Running.String()); err != nil {
				logger.Error(fmt.Sprintf("failed to update mcp to intelligent task logs: %s", err.Error()), zap.Int64("taskId", id))
			}
			continue
		}

		// 获取实例令牌
		tokens, err := mysql.McpTokenRepo.ListByInstanceID(context.Background(), mcpInstance.InstanceID)
		if err != nil {
			logs = append(logs, &model.InstallLog{
				McpInstanceID:   instanceID,
				McpInstanceName: "",
				Status:          false,
				ErrorLog:        fmt.Sprintf("failed to find mcpInstance: %s", err.Error()),
			})
			lastStatus = pb.McpToIntelligentTaskStatus_Failed.String()
			if err = mysql.McpToIntelligentTaskRepo.UpdateLogs(context.Background(), id, logs, pb.McpToIntelligentTaskStatus_Running.String()); err != nil {
				logger.Error(fmt.Sprintf("failed to update mcp to intelligent task logs: %s", err.Error()), zap.Int64("taskId", id))
			}
			continue
		}

		log := &model.InstallLog{
			McpInstanceID:         instanceID,
			McpInstanceName:       mcpInstance.InstanceName,
			Status:                true,
			ErrorLog:              "",
			InsertIntelligentLogs: nil,
		}
		logs = append(logs, log)
		// 记录每条 space 的插入日志
		var insertIntelligentLogs []*model.InsertIntelligentLog
		// 遍历插入信息列表
		for _, insertInfo := range task.InsertIntelligentInfos {
			wg.Add(1)
			sem <- struct{}{}
			go func(insertInfo *model.InsertIntelligentInfo) {
				defer func() {
					<-sem
					wg.Done()
				}()

				searchTask, err := mysql.McpToIntelligentTaskRepo.FindByID(context.Background(), id)
				if err != nil {
					logger.Error(fmt.Sprintf("failed to find mcp to intelligent task: %s", err.Error()), zap.Int64("taskId", id))
				} else {
					// 监控任务是否取消，如果取消则直接跳出循环
					if searchTask.Status == pb.McpToIntelligentTaskStatus_Cancel.String() {
						return
					}
				}

				// 执行创建 dify tools
				err = createDifyTools(instanceID, task.Domain, insertInfo, mcpInstance, userSpaces, conn)
				if err != nil {
					mu.Lock()
					insertIntelligentLogs = append(insertIntelligentLogs, &model.InsertIntelligentLog{
						InsertIntelligentInfo: insertInfo,
						Status:                false,
						ErrorLog:              err.Error(),
					})

					log.Status = false
					log.ErrorLog = err.Error()
					log.InsertIntelligentLogs = insertIntelligentLogs

					lastStatus = pb.McpToIntelligentTaskStatus_Failed.String()
					if err = mysql.McpToIntelligentTaskRepo.UpdateLogs(context.Background(), id, logs, pb.McpToIntelligentTaskStatus_Running.String()); err != nil {
						logger.Error(fmt.Sprintf("failed to update mcp to intelligent task logs: %s", err.Error()), zap.Int64("taskId", id))
					}
					mu.Unlock()
					return
				}

				if insertInfo.Headers[instanceID] != nil {
					err = createOrUpdateInstanceToken(instanceID, insertInfo, task, intelligentAccess, tokens)
					if err != nil {
						mu.Lock()
						insertIntelligentLogs = append(insertIntelligentLogs, &model.InsertIntelligentLog{
							InsertIntelligentInfo: insertInfo,
							Status:                false,
							ErrorLog:              err.Error(),
						})
						log.Status = false
						log.ErrorLog = err.Error()
						log.InsertIntelligentLogs = insertIntelligentLogs

						lastStatus = pb.McpToIntelligentTaskStatus_Failed.String()
						if err = mysql.McpToIntelligentTaskRepo.UpdateLogs(context.Background(), id, logs, pb.McpToIntelligentTaskStatus_Running.String()); err != nil {
							logger.Error(fmt.Sprintf("failed to update mcp to intelligent task logs: %s", err.Error()), zap.Int64("taskId", id))
						}
						mu.Unlock()
						return
					}
				}

				mu.Lock()
				insertIntelligentLogs = append(insertIntelligentLogs, &model.InsertIntelligentLog{
					InsertIntelligentInfo: insertInfo,
					Status:                true,
					ErrorLog:              "",
				})
				log.InsertIntelligentLogs = insertIntelligentLogs
				if err = mysql.McpToIntelligentTaskRepo.UpdateLogs(context.Background(), id, logs, pb.McpToIntelligentTaskStatus_Running.String()); err != nil {
					logger.Error(fmt.Sprintf("failed to update mcp to intelligent task logs: %s", err.Error()), zap.Int64("taskId", id))
				}
				mu.Unlock()
			}(insertInfo)
		}
		wg.Wait()
	}
	// 最后更新整个任务的状态
	if err = mysql.McpToIntelligentTaskRepo.UpdateLogs(context.Background(), id, logs, lastStatus); err != nil {
		logger.Error(fmt.Sprintf("failed to update mcp to intelligent task logs: %s", err.Error()), zap.Int64("taskId", id))
	}
}

func createOrUpdateInstanceToken(instanceID string, insertInfo *model.InsertIntelligentInfo, task *model.McpToIntelligentTask, intelligentAccess *model.IntelligentAccess, tokens []*model.McpToken) error {
	headerJson, _ := json.Marshal(insertInfo.Headers[instanceID].Headers)
	// 给 instance 创建或者更新对应的 token
	usageToken := findTokenByUsage(tokens, insertInfo, task.IntelligentAccessID)
	if usageToken == nil {
		labels := []string{
			fmt.Sprintf("dify_user_id=%v", insertInfo.DifyUserID),
			fmt.Sprintf("dify_user_name=%v", insertInfo.DifyUserName),
			fmt.Sprintf("dify_space_id=%v", insertInfo.DifySpaceID),
			fmt.Sprintf("dify_space_name=%v", insertInfo.DifySpaceName),
			fmt.Sprintf("intelligent_access_id=%v", task.IntelligentAccessID),
			fmt.Sprintf("intelligent_access_name=%v", task.IntelligentAccessName),
			fmt.Sprintf("intelligent_access_type=%v", intelligentAccess.AccessType),
		}
		labelsJson, _ := json.Marshal(labels)
		usageToken = &model.McpToken{
			InstanceID: instanceID,
			Token:      insertInfo.Headers[instanceID].Token,
			Enabled:    true,
			Headers:    headerJson,
			Usages:     labelsJson,
			ExpireAt:   0,
			PublishAt:  time.Now().UnixMilli(),
		}
		err := mysql.McpTokenRepo.Create(context.Background(), usageToken)
		if err != nil {
			return fmt.Errorf("failed to create token: %s", err.Error())
		}
	} else {
		usageToken.Headers = headerJson
		err := mysql.McpTokenRepo.Update(context.Background(), usageToken)
		if err != nil {
			return fmt.Errorf("failed to update token: %s", err.Error())
		}
	}
	return nil
}

func createDifyTools(instanceID string, domain string, insertInfo *model.InsertIntelligentInfo, mcpInstance *model.McpInstance, userSpaces []*iapb.DifyUserSpace, difyConn *sql.DB) error {
	var findUserSpace *iapb.DifyUserSpace
	for _, userSpace := range userSpaces {
		if userSpace.UserID == insertInfo.DifyUserID && userSpace.TenantID == insertInfo.DifySpaceID {
			findUserSpace = userSpace
			break
		}
	}
	if findUserSpace == nil {
		return fmt.Errorf("failed to find dify user space")
	}

	// 获取该插入信息的对应的实例要传递的 header
	listToolsHeaders := map[string]string{}
	gatewayHeader := map[string]string{}
	insertHeader := insertInfo.Headers[instanceID]
	if insertHeader != nil {
		listToolsHeaders = insertHeader.Headers
		token, err := dify.EncryptToken(findUserSpace.EncryptPublicKey, insertHeader.Token)
		if err != nil {
			return fmt.Errorf("failed to encrypt dify token: %s", err.Error())
		}
		gatewayHeader["Authorization"] = token
	}

	// 给该 mcp 实例创建对应的 http client
	mcpClient, err := client.NewStreamableHttpClient(
		mcpInstance.ContainerServiceURL,
		transport.WithHTTPHeaders(listToolsHeaders),
	)
	if err != nil {
		return fmt.Errorf("failed to call mcp failed: %s", err.Error())
	}
	_, err = mcpClient.Initialize(context.Background(), mcp.InitializeRequest{})
	if err != nil {
		return fmt.Errorf("failed to call mcp failed, init mcp failed: %s", err.Error())
	}
	// 调用 mcp 服务的 list tools 接口
	tools, err := mcpClient.ListTools(context.Background(), mcp.ListToolsRequest{})
	if err != nil {
		return fmt.Errorf("failed to call mcp, list tools failed: %s", err.Error())
	}
	toolsJson, err := json.Marshal(tools.Tools)
	if err != nil {
		return fmt.Errorf("failed to marshal tools failed: %s", err.Error())
	}

	mcpServerUrl := fmt.Sprintf("%s%s", domain, mcpInstance.PublicProxyPath)
	serverURL, err := dify.EncryptToken(findUserSpace.EncryptPublicKey, mcpServerUrl)
	if err != nil {
		return fmt.Errorf("encrypt token failed: %s", err.Error())
	}
	mcpServerUrlHash := computeSHA256Hash(mcpServerUrl)
	provider, err := postgres.GetToolMcpProvider(difyConn, mcpServerUrlHash, insertInfo.DifySpaceID)
	if err != nil {
		return fmt.Errorf("get dify provider failed: %s", err.Error())
	}
	// 创建
	if provider == nil {
		gatewayHeaderJson, _ := json.Marshal(gatewayHeader)
		provider = &postgres.ToolMcpProvider{
			Name:                 mcpInstance.InstanceName,
			ServerIdentifier:     mcpInstance.InstanceID[:23],
			ServerURL:            serverURL,
			ServerURLHash:        mcpServerUrlHash,
			Icon:                 "",
			TenantId:             insertInfo.DifySpaceID,
			UserId:               insertInfo.DifyUserID,
			EncryptedCredentials: "{}",
			Authed:               true,
			Tools:                string(toolsJson),
			Timeout:              30,
			SseReadTimeout:       300,
			CreatedAt:            time.Now(),
			UpdatedAt:            time.Now(),
			EncryptedHeaders:     string(gatewayHeaderJson),
		}
		err = postgres.CreateToolMcpProvider(difyConn, provider)
		if err != nil {
			return fmt.Errorf("failed to create provider failed: %s", err.Error())
		}
	} else {
		provider.Name = mcpInstance.InstanceName
		provider.Tools = string(toolsJson)
		provider.UpdatedAt = time.Now()
		err := postgres.UpdateToolMcpProvider(difyConn, provider)
		if err != nil {
			return fmt.Errorf("failed to update provider failed: %s", err.Error())
		}
	}
	return nil
}

func computeSHA256Hash(input string) string {
	// 创建 SHA256 哈希对象
	hasher := sha256.New()

	// 写入数据
	hasher.Write([]byte(input))

	// 计算哈希并转换为十六进制字符串
	return hex.EncodeToString(hasher.Sum(nil))
}

func findTokenByUsage(tokens []*model.McpToken, info *model.InsertIntelligentInfo, intelligentAccessID int64) *model.McpToken {
	for _, token := range tokens {
		if len(token.Usages) == 0 {
			continue
		}

		var usages []string
		if len(token.Usages) > 0 {
			_ = json.Unmarshal(token.Usages, &usages)
		}

		matchUser := false
		for _, usage := range usages {
			if usage == fmt.Sprintf("dify_user_id=%s", info.DifyUserID) {
				matchUser = true
			}
		}

		matchSpace := false
		for _, usage := range usages {
			if usage == fmt.Sprintf("dify_space_id=%s", info.DifySpaceID) {
				matchSpace = true
			}
		}

		matchIntelligentAccess := false
		for _, usage := range usages {
			if usage == fmt.Sprintf("intelligent_access_id=%v", intelligentAccessID) {
				matchIntelligentAccess = true
			}
		}

		if matchSpace && matchUser && matchIntelligentAccess {
			return token
		}
	}
	return nil
}
