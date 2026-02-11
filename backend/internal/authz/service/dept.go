package service

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kymo-mcp/mcpcan/internal/authz/biz"
	"github.com/kymo-mcp/mcpcan/pkg/utils"
	"go.uber.org/zap"

	pb "github.com/kymo-mcp/mcpcan/api/authz/dept"
	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
	i18nresp "github.com/kymo-mcp/mcpcan/pkg/i18n"
	"github.com/kymo-mcp/mcpcan/pkg/logger"
)

// DeptService department HTTP service
type DeptService struct {
	biz *biz.DeptData
}

// NewDeptService creates department service instance
func NewDeptService() *DeptService {
	return &DeptService{
		biz: biz.NewDeptData(context.Background()),
	}
}

// CreateDept creates department
func (s *DeptService) CreateDept(c *gin.Context) {
	var req pb.CreateDeptRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	if req.Name == "" {
		common.GinError(c, i18nresp.CodeInternalError, "Department name is required")
		return
	}
	if req.Sort <= 0 {
		common.GinError(c, i18nresp.CodeInternalError, "Department sort is required")
		return
	}

	userInfo, err := utils.GetCurrentUser(c)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	// Convert request to model
	deptModel := &model.SysDept{
		Name:     req.Name,
		DeptSort: int(req.Sort),
		Source:   model.DeptSourcePlatform,
		CreateBy: &[]string{userInfo.Username}[0],
		UpdateBy: &[]string{userInfo.Username}[0],
	}
	if req.Status == pb.DeptStatus_DeptStatusEnabled {
		deptModel.Enabled = 1
	} else {
		deptModel.Enabled = 0
	}

	if req.ImageURL != "" {
		imageURL := req.ImageURL
		deptModel.ImageURL = &imageURL
	}

	// Check if department name already exists
	existDept, _ := mysql.SysDeptRepo.FindByName(c.Request.Context(), deptModel.Name)
	if existDept != nil {
		common.GinError(c, i18nresp.CodeInternalError, "Department name already exists")
		return
	}

	// Set parent department ID if provided
	var parentDept *model.SysDept
	if req.ParentId > 0 {
		parentID := uint(req.ParentId)
		parentDept, err = mysql.SysDeptRepo.FindByID(c.Request.Context(), parentID)
		if err != nil {
			logger.Error("Failed to get parent department", zap.Error(err))
			common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("Failed to get parent department: %v", err))
			return
		}
		if parentDept == nil {
			common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("Parent department with ID %d does not exist", parentID))
			return
		}
		deptModel.PID = &parentID
	}

	// Create department
	if err := mysql.SysDeptRepo.Create(c.Request.Context(), deptModel); err != nil {
		logger.Error("Failed to create department", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("Failed to create department: %v", err))
		return
	}

	// Convert model to response
	respData := s.convertModelToProto(deptModel)

	common.GinSuccess(c, respData)
}

// UpdateDept updates department
func (s *DeptService) UpdateDept(c *gin.Context) {
	var req pb.UpdateDeptRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	if req.Name == "" {
		common.GinError(c, i18nresp.CodeInternalError, "Department name is required")
		return
	}
	if req.Sort <= 0 {
		common.GinError(c, i18nresp.CodeInternalError, "Department sort is required")
		return
	}
	if req.ParentId > 0 && req.ParentId == req.Id {
		common.GinError(c, i18nresp.CodeInternalError, "Parent cannot be the same as the current department")
		return
	}

	userInfo, err := utils.GetCurrentUser(c)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	// Get department by ID
	deptModel, err := mysql.SysDeptRepo.FindByID(c.Request.Context(), uint(req.Id))
	if err != nil {
		logger.Error("Failed to get department", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("Failed to get department: %v", err))
		return
	}

	// Update department fields
	deptModel.Name = req.Name
	deptModel.DeptSort = int(req.Sort)
	deptModel.ImageURL = &req.ImageURL
	deptModel.UpdateBy = &userInfo.Username
	if req.Status == pb.DeptStatus_DeptStatusEnabled {
		deptModel.Enabled = 1
	} else {
		deptModel.Enabled = 0
	}

	// Update parent department ID if provided
	if req.ParentId > 0 {
		parentID := uint(req.ParentId)
		deptModel.PID = &parentID
	} else {
		deptModel.PID = nil
	}

	// Check if department name already exists
	existDept, _ := mysql.SysDeptRepo.FindByName(c.Request.Context(), deptModel.Name)
	if existDept != nil && existDept.DeptID != deptModel.DeptID {
		common.GinError(c, i18nresp.CodeInternalError, "Department name already exists")
		return
	}

	// Update department
	if err := mysql.SysDeptRepo.Update(c.Request.Context(), deptModel); err != nil {
		logger.Error("Failed to update department", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("Failed to update department: %v", err))
		return
	}

	// Convert model to response
	respData := s.convertModelToProto(deptModel)

	common.GinSuccess(c, respData)
}

// DeleteDept deletes department
func (s *DeptService) DeleteDept(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.GinError(c, i18nresp.CodeBadRequest, "Invalid department ID")
		return
	}

	if err := s.biz.BatchDeleteDepts(c.Request.Context(), []uint{uint(id)}); err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("Failed to delete department: %v", err))
		return
	}

	common.GinSuccess(c, nil)
}

// BatchDeleteDept batch deletes departments
func (s *DeptService) BatchDeleteDept(c *gin.Context) {
	var req pb.BatchDeleteDeptRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// Convert IDs to uint slice
	var deptIDs []uint
	for _, id := range req.Ids {
		deptIDs = append(deptIDs, uint(id))
	}
	if err := s.biz.BatchDeleteDepts(c.Request.Context(), deptIDs); err != nil {
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("Failed to delete department: %v", err))
		return
	}

	common.GinSuccess(c, nil)
}

// UpdateDeptStatus updates department status
func (s *DeptService) UpdateDeptStatus(c *gin.Context) {
	var req pb.UpdateDeptStatusRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	userInfo, err := utils.GetCurrentUser(c)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	// Get department by ID
	deptModel, err := mysql.SysDeptRepo.FindByID(c.Request.Context(), uint(req.Id))
	if err != nil {
		logger.Error("Failed to get department", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("Failed to get department: %v", err))
		return
	}

	deptModel.UpdateBy = &[]string{userInfo.Username}[0]
	// Update status
	if req.Status == int32(pb.DeptStatus_DeptStatusEnabled) {
		deptModel.Enabled = 1
	} else {
		deptModel.Enabled = 0
	}

	// Update department
	if err := mysql.SysDeptRepo.Update(c.Request.Context(), deptModel); err != nil {
		logger.Error("Failed to update department status", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("Failed to update department status: %v", err))
		return
	}

	common.GinSuccess(c, nil)
}

// FindDepts gets department list with pagination
func (s *DeptService) FindDepts(c *gin.Context) {
	var req pb.ListDeptsRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	var pid *uint
	if req.ParentId > 0 {
		pid = &[]uint{uint(req.ParentId)}[0]
	}

	// 如果需要通过名称模糊查询，则将 pid 设置为0，为0的时候代表全量查询
	if req.Name != "" {
		pid = &[]uint{uint(0)}[0]
	}

	// Get departments with pagination
	depts, _, err := mysql.SysDeptRepo.FindWithPagination(c.Request.Context(), 1, 99999, req.Name, pid, req.Status)
	if err != nil {
		logger.Error("Failed to get departments with pagination", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("Failed to get departments: %v", err))
		return
	}

	parentIDS := make([]uint, 0)
	for _, dept := range depts {
		parentIDS = append(parentIDS, dept.DeptID)
	}
	// Get sub departments
	subDepts, err := mysql.SysDeptRepo.FindByParentID(c.Request.Context(), parentIDS)
	if err != nil {
		logger.Error("Failed to get sub departments", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("Failed to get sub departments: %v", err))
		return
	}
	var childs = map[uint][]*model.SysDept{}
	for _, dept := range depts {
		childs[dept.DeptID] = make([]*model.SysDept, 0)
	}
	for _, dept := range subDepts {
		childs[*dept.PID] = append(childs[*dept.PID], dept)
	}

	// Convert models to response
	var respDepts []*pb.SysDept
	for _, dept := range depts {
		result := s.convertModelToProto(dept)
		if children, ok := childs[dept.DeptID]; ok {
			result.HasChildren = len(children) > 0
		}
		respDepts = append(respDepts, result)
	}

	common.GinSuccess(c, pb.ListDeptsResponse{
		List: respDepts,
	})
}

// convertModelToProto converts model.SysDept to pb.SysDept
func (s *DeptService) convertModelToProto(dept *model.SysDept) *pb.SysDept {
	resp := &pb.SysDept{
		Id:        int64(dept.DeptID),
		Name:      dept.Name,
		Sort:      int32(dept.DeptSort),
		Status:    pb.DeptStatus_DeptStatusDisabled,
		CreatedAt: 0,
		UpdatedAt: 0,
	}

	if dept.Enabled == 1 {
		resp.Status = pb.DeptStatus_DeptStatusEnabled
	}

	if dept.PID != nil {
		resp.ParentId = int64(*dept.PID)
	}

	if dept.CreateTime != nil {
		resp.CreatedAt = dept.CreateTime.UnixMilli()
	}

	if dept.UpdateTime != nil {
		resp.UpdatedAt = dept.UpdateTime.UnixMilli()
	}

	if dept.CreateBy != nil {
		resp.CreatedBy = *dept.CreateBy
	}

	if dept.UpdateBy != nil {
		resp.UpdatedBy = *dept.UpdateBy
	}

	return resp
}

// GetDeptTree gets department tree
func (s *DeptService) GetDeptTree(c *gin.Context) {
	var req pb.GetDeptTreeRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// Get departments with pagination
	depts, _, err := mysql.SysDeptRepo.FindWithPagination(c.Request.Context(), 1, 99999, "", &[]uint{0}[0], int32(pb.DeptStatus_DeptStatusEnabled))
	if err != nil {
		logger.Error("Failed to get departments with pagination", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, fmt.Sprintf("Failed to get departments: %v", err))
		return
	}

	var deptMap map[uint]*pb.SysDept = make(map[uint]*pb.SysDept)
	for _, dept := range depts {
		deptMap[dept.DeptID] = s.convertModelToProto(dept)
	}

	var rootDepts []*pb.SysDept
	for _, dept := range depts {
		if dept.PID == nil || *dept.PID == 0 {
			rootDepts = append(rootDepts, deptMap[dept.DeptID])
		} else {
			if parent, ok := deptMap[*dept.PID]; ok {
				parent.Children = append(parent.Children, s.convertModelToProto(dept))
				deptMap[*dept.PID] = parent
			}
		}
	}

	common.GinSuccess(c, rootDepts)
}
