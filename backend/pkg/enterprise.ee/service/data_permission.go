package service

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/kymo-mcp/mcpcan/api/authz/user_auth"
	datapermissionpb "github.com/kymo-mcp/mcpcan/api/market/data_permission"
	"github.com/kymo-mcp/mcpcan/pkg/enterprise.ee/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/enterprise.ee/database/mysql"
	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/gomap"
	i18nresp "github.com/kymo-mcp/mcpcan/pkg/i18n"
)

type DataPermissionService struct{}

func NewDataPermissionService() *DataPermissionService {
	return &DataPermissionService{}
}

func (s *DataPermissionService) SaveHandler(c *gin.Context) {
	var req datapermissionpb.SaveDataPermissionRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	var currentUserID int64
	if u, ok := gomap.Get(common.UserInfoContextKey).(*user_auth.UserInfo); ok && u != nil {
		currentUserID = u.UserId
	}

	var permissions []*model.SysDataPermission

	// Add access authorization (whitelist)
	if req.IsAllPersonnel {
		permissions = append(permissions, &model.SysDataPermission{
			DataType:    req.DataType,
			DataID:      req.DataId,
			TargetType:  datapermissionpb.TargetType_ALL.String(),
			TargetID:    0,
			IsBlacklist: 0,
			CreatedBy:   currentUserID,
		})
	} else {
		for _, deptID := range req.DeptIds {
			permissions = append(permissions, &model.SysDataPermission{
				DataType:    req.DataType,
				DataID:      req.DataId,
				TargetType:  datapermissionpb.TargetType_DEPT.String(),
				TargetID:    deptID,
				IsBlacklist: 0,
				CreatedBy:   currentUserID,
			})
		}
		for _, roleID := range req.RoleIds {
			permissions = append(permissions, &model.SysDataPermission{
				DataType:    req.DataType,
				DataID:      req.DataId,
				TargetType:  datapermissionpb.TargetType_ROLE.String(),
				TargetID:    roleID,
				IsBlacklist: 0,
				CreatedBy:   currentUserID,
			})
		}
		for _, userID := range req.UserIds {
			permissions = append(permissions, &model.SysDataPermission{
				DataType:    req.DataType,
				DataID:      req.DataId,
				TargetType:  datapermissionpb.TargetType_USER.String(),
				TargetID:    userID,
				IsBlacklist: 0,
				CreatedBy:   currentUserID,
			})
		}
	}

	// Add access blacklist
	for _, userID := range req.BlacklistUserIds {
		permissions = append(permissions, &model.SysDataPermission{
			DataType:    req.DataType,
			DataID:      req.DataId,
			TargetType:  datapermissionpb.TargetType_USER.String(),
			TargetID:    userID,
			IsBlacklist: 1,
			CreatedBy:   currentUserID,
		})
	}

	err := mysql.SysDataPermissionRepo.BatchSave(c.Request.Context(), req.DataType, req.DataId, permissions)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to save data permissions: %s", err.Error()))
		return
	}

	common.GinSuccess(c, &datapermissionpb.SaveDataPermissionResponse{})
}

func (s *DataPermissionService) GetHandler(c *gin.Context) {
	var req datapermissionpb.GetDataPermissionRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		common.GinError(c, i18nresp.CodeParameterInvalid, err.Error())
		return
	}

	permissions, err := mysql.SysDataPermissionRepo.GetByData(c.Request.Context(), req.DataType, req.DataId)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("failed to get data permissions: %s", err.Error()))
		return
	}

	resp := &datapermissionpb.GetDataPermissionResponse{
		DataType:         req.DataType,
		DataId:           req.DataId,
		DeptIds:          []int64{},
		RoleIds:          []int64{},
		UserIds:          []int64{},
		BlacklistUserIds: []int64{},
	}

	for _, p := range permissions {
		if p.IsBlacklist == 1 {
			if p.TargetType == datapermissionpb.TargetType_USER.String() {
				resp.BlacklistUserIds = append(resp.BlacklistUserIds, p.TargetID)
			}
		} else {
			switch p.TargetType {
			case datapermissionpb.TargetType_ALL.String():
				resp.IsAllPersonnel = true
			case datapermissionpb.TargetType_DEPT.String():
				resp.DeptIds = append(resp.DeptIds, p.TargetID)
			case datapermissionpb.TargetType_ROLE.String():
				resp.RoleIds = append(resp.RoleIds, p.TargetID)
			case datapermissionpb.TargetType_USER.String():
				resp.UserIds = append(resp.UserIds, p.TargetID)
			}
		}
	}

	common.GinSuccess(c, resp)
}
