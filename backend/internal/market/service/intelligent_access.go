package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	pb "github.com/kymo-mcp/mcpcan/api/market/intelligent_access"
	"github.com/kymo-mcp/mcpcan/internal/market/config"
	"github.com/kymo-mcp/mcpcan/internal/market/repository"
	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/postgres"
	i18nresp "github.com/kymo-mcp/mcpcan/pkg/i18n"
	_ "github.com/lib/pq"
)

// IntelligentAccessService struct for intelligent access service
type IntelligentAccessService struct {
	ctx context.Context
}

// NewIntelligentAccessService creates a new intelligent access service
func NewIntelligentAccessService(ctx context.Context) *IntelligentAccessService {
	return &IntelligentAccessService{
		ctx: ctx,
	}
}

// CreateHandler creates intelligent access HTTP handler function
func (s *IntelligentAccessService) CreateHandler(c *gin.Context) {
	if config.GetConfig().RunKimo {
		common.GinError(c, i18nresp.CodeForbidden, "cannot create intelligent access when running in kimo mode")
		return
	}

	var req pb.CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.GinError(c, i18nresp.CodeBadRequest, err.Error())
		return
	}
	// validate request
	if req.AccessName == "" || req.AccessType == "" || req.DbHost == "" || req.DbPort == 0 || req.DbUser == "" || req.DbPassword == "" || req.DbName == "" {
		common.GinError(c, i18nresp.CodeBadRequest, "invalid request")
		return
	}
	if req.AccessType != pb.IntelligentAccessType_DifyEnterprise.String() && req.AccessType != pb.IntelligentAccessType_Dify.String() {
		common.GinError(c, i18nresp.CodeBadRequest, "invalid access type")
		return
	}

	intelligentAccess := &model.IntelligentAccess{
		AccessName: req.AccessName,
		AccessType: req.AccessType,
		DbHost:     req.DbHost,
		DbPort:     int(req.DbPort),
		DbUser:     req.DbUser,
		DbPassword: req.DbPassword,
		DbName:     req.DbName,
	}

	if err := repository.IntelligentAccessRepo.Create(s.ctx, intelligentAccess); err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to create intelligent access: %s", err.Error()))
		return
	}

	common.GinSuccess(c, &pb.CreateResponse{
		IntelligentAccess: &pb.IntelligentAccess{
			AccessID:   intelligentAccess.ID,
			AccessName: intelligentAccess.AccessName,
			AccessType: intelligentAccess.AccessType,
			DbHost:     intelligentAccess.DbHost,
			DbPort:     int32(intelligentAccess.DbPort),
			DbUser:     intelligentAccess.DbUser,
			DbPassword: intelligentAccess.DbPassword,
			DbName:     intelligentAccess.DbName,
			CreateTime: intelligentAccess.CreateTime.UnixMilli(),
			UpdateTime: intelligentAccess.UpdateTime.UnixMilli(),
		},
	})
}

// UpdateHandler updates intelligent access HTTP handler function
func (s *IntelligentAccessService) UpdateHandler(c *gin.Context) {
	if config.GetConfig().RunKimo {
		common.GinError(c, i18nresp.CodeForbidden, "cannot update intelligent access when running in kimo mode")
		return
	}

	var req pb.UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.GinError(c, i18nresp.CodeBadRequest, err.Error())
		return
	}
	// validate request
	if req.AccessID == 0 || req.AccessName == "" || req.DbHost == "" || req.DbPort == 0 || req.DbUser == "" || req.DbPassword == "" || req.DbName == "" {
		common.GinError(c, i18nresp.CodeBadRequest, "invalid request")
		return
	}

	dbIntelligentAccess, err := repository.IntelligentAccessRepo.FindByID(s.ctx, req.AccessID)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to find intelligent access: %s", err.Error()))
		return
	}

	dbIntelligentAccess.AccessName = req.AccessName
	dbIntelligentAccess.DbHost = req.DbHost
	dbIntelligentAccess.DbName = req.DbName
	dbIntelligentAccess.DbPassword = req.DbPassword
	dbIntelligentAccess.DbPort = int(req.DbPort)
	dbIntelligentAccess.DbUser = req.DbUser

	if err := repository.IntelligentAccessRepo.Update(s.ctx, dbIntelligentAccess); err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to update intelligent access: %s", err.Error()))
		return
	}

	common.GinSuccess(c, &pb.UpdateResponse{
		IntelligentAccess: &pb.IntelligentAccess{
			AccessID:   dbIntelligentAccess.ID,
			AccessName: dbIntelligentAccess.AccessName,
			AccessType: dbIntelligentAccess.AccessType,
			DbHost:     dbIntelligentAccess.DbHost,
			DbPort:     int32(dbIntelligentAccess.DbPort),
			DbUser:     dbIntelligentAccess.DbUser,
			DbPassword: dbIntelligentAccess.DbPassword,
			DbName:     dbIntelligentAccess.DbName,
			CreateTime: dbIntelligentAccess.CreateTime.UnixMilli(),
			UpdateTime: dbIntelligentAccess.UpdateTime.UnixMilli(),
		},
	})
}

// DeleteHandler deletes intelligent access HTTP handler function
func (s *IntelligentAccessService) DeleteHandler(c *gin.Context) {
	if config.GetConfig().RunKimo {
		common.GinError(c, i18nresp.CodeForbidden, "cannot delete intelligent access when running in kimo mode")
		return
	}

	var req pb.DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.GinError(c, i18nresp.CodeBadRequest, err.Error())
		return
	}

	if err := repository.IntelligentAccessRepo.Delete(s.ctx, req.AccessID); err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to delete intelligent access: %s", err.Error()))
		return
	}

	common.GinSuccess(c, &pb.DeleteResponse{})
}

// GetHandler finds intelligent access by ID HTTP handler function
func (s *IntelligentAccessService) GetHandler(c *gin.Context) {
	var req pb.GetRequest
	if err := c.ShouldBindUri(&req); err != nil {
		common.GinError(c, i18nresp.CodeBadRequest, err.Error())
		return
	}

	intelligentAccess, err := repository.IntelligentAccessRepo.FindByID(s.ctx, req.AccessID)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to find intelligent access: %s", err.Error()))
		return
	}

	common.GinSuccess(c, &pb.GetResponse{
		IntelligentAccess: &pb.IntelligentAccess{
			AccessID:   intelligentAccess.ID,
			AccessName: intelligentAccess.AccessName,
			AccessType: intelligentAccess.AccessType,
			DbHost:     intelligentAccess.DbHost,
			DbPort:     int32(intelligentAccess.DbPort),
			DbUser:     intelligentAccess.DbUser,
			DbPassword: intelligentAccess.DbPassword,
			DbName:     intelligentAccess.DbName,
			CreateTime: intelligentAccess.CreateTime.UnixMilli(),
			UpdateTime: intelligentAccess.UpdateTime.UnixMilli(),
		},
	})
}

// ListHandler finds all intelligent access HTTP handler function
func (s *IntelligentAccessService) ListHandler(c *gin.Context) {
	intelligentAccesses, err := repository.IntelligentAccessRepo.FindAll(s.ctx)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to find all intelligent access: %s", err.Error()))
		return
	}

	var pbIntelligentAccesses []*pb.IntelligentAccess
	for _, intelligentAccess := range intelligentAccesses {
		pbIntelligentAccesses = append(pbIntelligentAccesses, &pb.IntelligentAccess{
			AccessID:   intelligentAccess.ID,
			AccessName: intelligentAccess.AccessName,
			AccessType: intelligentAccess.AccessType,
			DbHost:     intelligentAccess.DbHost,
			DbPort:     int32(intelligentAccess.DbPort),
			DbUser:     intelligentAccess.DbUser,
			DbPassword: intelligentAccess.DbPassword,
			DbName:     intelligentAccess.DbName,
			CreateTime: intelligentAccess.CreateTime.UnixMilli(),
			UpdateTime: intelligentAccess.UpdateTime.UnixMilli(),
		})
	}

	common.GinSuccess(c, &pb.ListResponse{
		List: pbIntelligentAccesses,
	})
}

func (s *IntelligentAccessService) TestConnectionHandler(c *gin.Context) {
	var req pb.TestConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.GinError(c, i18nresp.CodeBadRequest, err.Error())
		return
	}

	db, err := BuildTemporaryPostgresConnection(req.DbHost, int(req.DbPort), req.DbUser, req.DbPassword, req.DbName)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to connect to database: %s", err.Error()))
		return
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to ping database: %s", err.Error()))
		return
	}

	common.GinSuccess(c, &pb.TestConnectionResponse{
		Success: true,
		Message: "PostgreSQL database connection successful",
	})
}

func BuildTemporaryPostgresConnection(host string, port int, user string, password string, database string) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Shanghai",
		host, port, user, password, database)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %s", err)
	}

	// 完全禁用连接池，每次都是新连接
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(0)
	db.SetConnMaxLifetime(0)

	return db, nil
}
func (s *IntelligentAccessService) ListDifyUserSpaceHandler(c *gin.Context) {
	var req pb.ListDifyUserSpaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.GinError(c, i18nresp.CodeBadRequest, err.Error())
		return
	}

	intelligentAccess, err := repository.IntelligentAccessRepo.FindByID(s.ctx, req.AccessID)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to find intelligent access: %s", err.Error()))
		return
	}

	// 连接数据库
	sqlDB, err := BuildTemporaryPostgresConnection(intelligentAccess.DbHost, int(intelligentAccess.DbPort), intelligentAccess.DbUser, intelligentAccess.DbPassword, intelligentAccess.DbName)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to connect to database: %s", err.Error()))
		return
	}
	defer sqlDB.Close()

	pbUserSpaces, err := GetDifyUserSpace(sqlDB)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to query user spaces: %s", err.Error()))
		return
	}

	// 给每个用户回显设置对应实例的 token
	for _, instanceID := range req.InstancesIDs {
		// 获取实例令牌
		tokens, err := mysql.McpTokenRepo.ListByInstanceID(context.Background(), instanceID)
		if err != nil {
			common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to find mcp instance token: %s", err.Error()))
			return
		}

		if len(tokens) == 0 {
			continue
		}
		// 找到默认 token
		defaultToken := tokens[0]
		for _, token := range tokens {
			usages := []string{}
			_ = json.Unmarshal(token.Usages, &usages)
			for _, usage := range usages {
				if usage == "default" {
					defaultToken = token
					break
				}
			}
		}

		for _, userSpace := range pbUserSpaces {
			userSpace.EncryptPublicKey = ""

			if userSpace.Headers == nil {
				userSpace.Headers = map[string]*pb.HeaderInfo{}
			}

			findToken := findTokenByUsage(tokens, &model.InsertIntelligentInfo{
				DifySpaceID: userSpace.TenantID,
				DifyUserID:  userSpace.UserID,
			}, req.AccessID)
			if findToken != nil {
				headers := map[string]string{}
				_ = json.Unmarshal(findToken.Headers, &headers)

				userSpace.Headers[instanceID] = &pb.HeaderInfo{
					Token:   findToken.Token,
					Headers: headers,
				}
			} else {
				headers := map[string]string{}
				_ = json.Unmarshal(defaultToken.Headers, &headers)

				userSpace.Headers[instanceID] = &pb.HeaderInfo{
					Token:   defaultToken.Token,
					Headers: headers,
				}
			}
		}
	}

	common.GinSuccess(c, &pb.ListDifyUserSpaceResponse{
		UserSpaces: pbUserSpaces,
	})
}

func GetDifyUserSpace(sqlDB *sql.DB) ([]*pb.DifyUserSpace, error) {
	userSpaces, err := postgres.GetOwnerTenantAccountJoins(sqlDB)
	if err != nil {
		return nil, fmt.Errorf("failed to query user spaces: %s", err.Error())
	}
	tenants, err := postgres.GetAllTenants(sqlDB)
	if err != nil {
		return nil, fmt.Errorf("failed to query tenants: %s", err.Error())
	}
	accounts, err := postgres.GetAllAccounts(sqlDB)
	if err != nil {
		return nil, fmt.Errorf("failed to query accounts: %s", err.Error())
	}
	tenantMap := make(map[string]postgres.Tenant)
	for _, tenant := range tenants {
		tenantMap[tenant.Id] = tenant
	}
	accountMap := make(map[string]string)
	for _, account := range accounts {
		accountMap[account.Id] = account.Name
	}

	var pbUserSpaces []*pb.DifyUserSpace
	for _, userSpace := range userSpaces {
		pbUserSpaces = append(pbUserSpaces, &pb.DifyUserSpace{
			UserID:           userSpace.AccountId,
			TenantID:         userSpace.TenantId,
			UserName:         accountMap[userSpace.AccountId],
			TenantName:       tenantMap[userSpace.TenantId].Name,
			EncryptPublicKey: tenantMap[userSpace.TenantId].EncryptPublicKey,
		})
	}
	return pbUserSpaces, nil
}
