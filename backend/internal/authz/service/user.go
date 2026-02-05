package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
	"go.uber.org/zap"

	"github.com/kymo-mcp/mcpcan/api/authz/user"
	"github.com/kymo-mcp/mcpcan/internal/authz/biz"
	"github.com/kymo-mcp/mcpcan/internal/authz/config"
	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	i18nresp "github.com/kymo-mcp/mcpcan/pkg/i18n"
	"github.com/kymo-mcp/mcpcan/pkg/logger"
	"github.com/kymo-mcp/mcpcan/pkg/utils"
)

// UserService user HTTP service
type UserService struct {
	userBiz *biz.UserBiz
}

// NewUserService creates user service instance
func NewUserService() *UserService {
	return &UserService{
		userBiz: biz.NewUserBiz(),
	}
}

// CreateUser creates user
func (s *UserService) CreateUser(c *gin.Context) {
	var req user.CreateUserRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}
	userInfo, err := utils.GetCurrentUser(c)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	// Convert request to model
	userModel := s.convertCreateRequestToModel(&req)
	userModel.CreateBy = &userInfo.Username
	userModel.UpdateBy = &userInfo.Username

	// Create user
	if err := mysql.SysUserRepo.Create(c.Request.Context(), userModel); err != nil {
		logger.Error("Failed to create user", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	// If password is provided, hash it
	if req.Password != "" {
		if err := s.userBiz.SetUserPassword(c.Request.Context(), userModel, req.Password); err != nil {
			logger.Error("Failed to set user password", zap.Error(err))
			common.GinError(c, i18nresp.CodeInternalError, "Failed to set user password")
			return
		}
	}

	// Create user roles associations
	associations := make([]*model.SysUsersRoles, 0, len(req.RoleIds))
	for _, roleID := range req.RoleIds {
		associations = append(associations, &model.SysUsersRoles{
			UserID: userModel.UserID,
			RoleID: uint(roleID),
		})
	}
	if err := mysql.SysUsersRolesRepo.BatchCreate(c.Request.Context(), associations); err != nil {
		logger.Error("Failed to create user roles associations", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	// Return created user information
	response := s.convertModelToProto(userModel)
	common.GinSuccess(c, response)
}

// UpdateUser updates user
func (s *UserService) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, "Invalid user ID")
		return
	}
	userInfo, err := utils.GetCurrentUser(c)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	var req user.UpdateUserRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// Get existing user
	existingUser, err := mysql.SysUserRepo.FindByID(c.Request.Context(), uint(id))
	if err != nil {
		logger.Error("Failed to get user", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	// Update model
	s.updateModelFromRequest(existingUser, &req)
	existingUser.UpdateBy = &userInfo.Username

	//// If password is provided, hash it
	//if req.Password != "" {
	//	if err := s.userBiz.SetUserPassword(c.Request.Context(), existingUser, req.Password); err != nil {
	//		logger.Error("Failed to set user password", zap.Error(err))
	//		common.GinError(c, i18nresp.CodeInternalError, "Failed to set user password")
	//		return
	//	}
	//}

	// Update user
	if err := mysql.SysUserRepo.Update(c.Request.Context(), existingUser); err != nil {
		logger.Error("Failed to update user", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}
	// delete role associations
	if err := mysql.SysUsersRolesRepo.BatchDeleteByUserID(c.Request.Context(), []uint{existingUser.UserID}); err != nil {
		logger.Error("Failed to delete user roles associations", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}
	// Create user roles associations
	associations := make([]*model.SysUsersRoles, 0, len(req.RoleIds))
	for _, roleID := range req.RoleIds {
		associations = append(associations, &model.SysUsersRoles{
			UserID: existingUser.UserID,
			RoleID: uint(roleID),
		})
	}
	if err := mysql.SysUsersRolesRepo.BatchCreate(c.Request.Context(), associations); err != nil {
		logger.Error("Failed to create user roles associations", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	// Return updated user information
	userProto := s.convertModelToProto(existingUser)
	common.GinSuccess(c, userProto)
}

// DeleteUser deletes user
func (s *UserService) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.GinError(c, i18nresp.CodeInternalError, "Invalid user ID")
		return
	}

	if err := s.userBiz.BatchDeleteByUserID(c.Request.Context(), []uint{uint(id)}); err != nil {
		logger.Error("Failed to delete user", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	common.GinSuccess(c, nil)
}

func (s *UserService) BatchDelete(c *gin.Context) {
	var req user.BatchDeleteUserRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	var deleteUserId []uint
	for _, userId := range req.UserIds {
		deleteUserId = append(deleteUserId, uint(userId))
	}

	if err := s.userBiz.BatchDeleteByUserID(c.Request.Context(), deleteUserId); err != nil {
		logger.Error("Failed to delete user", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	common.GinSuccess(c, nil)
}

// ListUsers gets user list
func (s *UserService) ListUsers(c *gin.Context) {
	var req user.ListUsersRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// Set default pagination parameters
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 {
		req.Size = 10
	}
	if req.Size > 100 {
		req.Size = 100
	}

	var enable *bool
	if req.Status == user.UserStatus_UserStatusEnabled {
		enable = &[]bool{true}[0]
	} else if req.Status == user.UserStatus_UserStatusDisabled {
		enable = &[]bool{false}[0]
	}
	var depIds []uint
	for _, depId := range req.DeptIds {
		depIds = append(depIds, uint(depId))
	}

	// Get user list
	users, total, err := mysql.SysUserRepo.FindWithPagination(c.Request.Context(), int(req.Page), int(req.Size), req.Blurry, enable, depIds)
	if err != nil {
		logger.Error("Failed to get user list", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	// Convert to response format
	userProtos := make([]*user.SysUser, len(users))
	for i, u := range users {
		userProtos[i] = s.convertModelToProto(u)
	}

	// Set user role and department info
	err = setUsersRoleAndDeptInfo(c.Request.Context(), userProtos)
	if err != nil {
		logger.Error("Failed to set user role and dept info", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	response := &user.ListUsersResponse{
		List:  userProtos,
		Total: total,
	}
	common.GinSuccess(c, response)
}

// convertCreateRequestToModel converts create request to model
func (s *UserService) convertCreateRequestToModel(req *user.CreateUserRequest) *model.SysUser {
	userModel := &model.SysUser{
		Username: &req.Username,
		NickName: &req.FullName,
		Email:    &req.Email,
		Phone:    &req.Phone,
	}

	if req.DeptId > 0 {
		deptID := uint(req.DeptId)
		userModel.DeptID = &deptID
	}

	if req.Status == user.UserStatus_UserStatusEnabled {
		userModel.Enabled = &[]bool{true}[0]
	} else {
		userModel.Enabled = &[]bool{false}[0]
	}

	// Handle password
	if req.Password != "" {
		userModel.Password = &req.Password
	}

	return userModel
}

// updateModelFromRequest updates model from update request
func (s *UserService) updateModelFromRequest(userModel *model.SysUser, req *user.UpdateUserRequest) {
	if req.FullName != "" {
		userModel.NickName = &req.FullName
	}
	if req.Email != "" {
		userModel.Email = &req.Email
	}
	if req.Phone != "" {
		userModel.Phone = &req.Phone
	}
	if req.DeptId > 0 {
		deptID := uint(req.DeptId)
		userModel.DeptID = &deptID
	}
	switch req.Status {
	case user.UserStatus_UserStatusEnabled:
		userModel.Enabled = &[]bool{true}[0]
	case user.UserStatus_UserStatusDisabled:
		userModel.Enabled = &[]bool{false}[0]
	}
}

// UpdatePassword updates user password
func (s *UserService) UpdatePassword(c *gin.Context) {
	var req user.UpdatePasswordRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	userId := c.GetInt64("userId")

	// Basic parameter validation
	if userId <= 0 {
		common.GinError(c, i18nresp.CodeUserIDInvalid, "")
		return
	}

	if req.OldPassword == "" {
		common.GinError(c, i18nresp.CodeOldPasswordEmpty, "")
		return
	}

	if req.NewPassword == "" {
		common.GinError(c, i18nresp.CodeNewPasswordEmpty, "")
		return
	}

	if req.ConfirmPassword == "" {
		common.GinError(c, i18nresp.CodeConfirmPasswordEmpty, "")
		return
	}

	// Verify new password and confirm password match
	if req.NewPassword != req.ConfirmPassword {
		common.GinError(c, i18nresp.CodePasswordMismatch, "")
		return
	}

	// Verify new password strength
	if isValid, errorCode := common.ValidatePasswordStrengthWithI18n(req.NewPassword); !isValid {
		common.GinError(c, errorCode, "")
		return
	}

	// Verify new password can't be the same as old password
	if req.OldPassword == req.NewPassword {
		common.GinError(c, i18nresp.CodePasswordSameAsOld, "")
		return
	}

	// Use business layer's UpdatePassword method for password update
	if err := s.userBiz.UpdatePassword(c.Request.Context(), uint(userId), req.OldPassword, req.NewPassword); err != nil {
		logger.Error("Failed to update password", zap.Error(err), zap.Int64("userId", userId))
		// Return corresponding error code based on error type
		if strings.Contains(err.Error(), "旧密码不正确") || strings.Contains(err.Error(), "old password") {
			common.GinError(c, i18nresp.CodeOldPasswordIncorrect, "")
		} else if strings.Contains(err.Error(), "用户不存在") || strings.Contains(err.Error(), "user not found") {
			common.GinError(c, i18nresp.CodeUserNotFoundError, "")
		} else {
			common.GinError(c, i18nresp.CodeUpdatePasswordFailure, "")
		}
		return
	}

	common.GinSuccess(c, nil)
}

// UpdateAvatar updates user avatar
func (s *UserService) UpdateAvatar(c *gin.Context) {
	userId := c.GetInt64("userId")
	if userId <= 0 {
		common.GinError(c, i18nresp.CodeUserIDInvalid, "")
		return
	}
	// Get uploaded file
	imageFile, err := c.FormFile("image")
	if err != nil {
		logger.Error("Failed to get image file", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "No image file provided")
		return
	}

	// Open file
	file, err := imageFile.Open()
	if err != nil {
		logger.Error("Failed to open image file", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "Failed to open image file")
		return
	}
	defer file.Close()

	// Read file content
	imageData, err := io.ReadAll(file)
	if err != nil {
		logger.Error("Failed to read image file", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "Failed to read image file")
		return
	}

	// Validate image type
	if !utils.IsValidImageType(imageData) {
		logger.Error("Invalid image type")
		common.GinError(c, i18nresp.CodeInternalError, "Unsupported image type")
		return
	}

	// Validate file size (5MB limit)
	maxSize := int64(5 * 1024 * 1024)
	if int64(len(imageData)) > maxSize {
		logger.Error("Image file too large", zap.Int("size", len(imageData)))
		common.GinError(c, i18nresp.CodeInternalError, "Image file too large")
		return
	}

	// Get user information
	userModel, err := mysql.SysUserRepo.FindByID(c.Request.Context(), uint(userId))
	if err != nil {
		logger.Error("Failed to get user", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	// Generate file name
	ext := utils.GetImageFileExtension(imageData)
	if ext == "" {
		ext = "jpg"
	}
	fileName := fmt.Sprintf("%d.%s", userId, ext)
	// Build storage path
	storageDir := filepath.Join(config.GlobalConfig.Storage.StaticPath, strings.Trim(common.AvatarPath, "/"))
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		logger.Error("Failed to create storage directory", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "Failed to create storage directory")
		return
	}
	// Save file
	filePath := filepath.Join(storageDir, fileName)
	if err := os.WriteFile(filePath, imageData, 0644); err != nil {
		logger.Error("Failed to save image file", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "Failed to save image file")
		return
	}

	// Update avatar path
	imagePath := filepath.Join(common.StaticPrefix, common.AvatarPath, fileName)
	userModel.AvatarPath = &imagePath

	if err := mysql.SysUserRepo.Update(c.Request.Context(), userModel); err != nil {
		logger.Error("Failed to update avatar", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, err.Error())
		return
	}

	resp := &user.UpdateAvatarResponse{
		Path: imagePath,
		Size: int64(len(imageData)),
		Mime: fmt.Sprintf("image/%s", ext),
	}
	common.GinSuccess(c, resp)
}

// convertModelToProto converts model to Proto
func (s *UserService) convertModelToProto(userModel *model.SysUser) *user.SysUser {
	userProto := &user.SysUser{
		Id:       int64(userModel.UserID),
		Username: userModel.GetUsername(),
		FullName: userModel.GetNickName(),
		Email:    userModel.GetEmail(),
		Phone:    userModel.GetPhone(),
		Avatar:   userModel.GetAvatarPath(),
	}

	if userModel.DeptID != nil {
		userProto.DeptId = int64(*userModel.DeptID)
	}

	if *userModel.Enabled {
		userProto.Status = user.UserStatus_UserStatusEnabled
	} else {
		userProto.Status = user.UserStatus_UserStatusDisabled
	}

	if userModel.CreateTime != nil {
		userProto.CreatedAt = userModel.CreateTime.Unix()
	}
	if userModel.UpdateTime != nil {
		userProto.UpdatedAt = userModel.UpdateTime.Unix()
	}

	return userProto
}

func setUsersRoleAndDeptInfo(ctx context.Context, users []*user.SysUser) error {
	var deptIDs []uint
	var userIDs []uint
	var userMap = map[uint]*user.SysUser{}
	for _, u := range users {
		userMap[uint(u.Id)] = u
		userIDs = append(userIDs, uint(u.Id))
		if u.DeptId > 0 {
			deptIDs = append(deptIDs, uint(u.DeptId))
		}
	}

	// 设置用户角色
	userRoles, err := mysql.SysUsersRolesRepo.BatchFindByUserID(ctx, userIDs)
	if err != nil {
		return fmt.Errorf("failed to query user roles: %v", err)
	}
	var roleIDs []uint
	for _, ur := range userRoles {
		roleIDs = append(roleIDs, ur.RoleID)
		u := userMap[ur.UserID]
		if u != nil {
			u.RoleIds = append(u.RoleIds, int64(ur.RoleID))
		}
	}
	roles, err := mysql.SysRoleRepo.FindByIDs(ctx, roleIDs)
	if err != nil {
		return fmt.Errorf("failed to query roles: %v", err)
	}
	var roleMap = make(map[uint]*model.SysRole)
	for _, role := range roles {
		roleMap[role.RoleID] = role
	}

	// 设置部门名称
	depts, err := mysql.SysDeptRepo.FindByIDs(ctx, deptIDs)
	if err != nil {
		return fmt.Errorf("failed to query departments: %v", err)
	}
	var deptMap = make(map[uint]*model.SysDept)
	for _, dept := range depts {
		deptMap[dept.DeptID] = dept
	}

	// 设置部门名称，和角色信息
	for _, u := range users {
		dept := deptMap[uint(u.DeptId)]
		if dept != nil {
			u.DeptName = dept.Name
		}
		for _, roleID := range u.RoleIds {
			role := roleMap[uint(roleID)]
			if role != nil {
				u.RoleNames = append(u.RoleNames, role.Name)
			}
		}
	}
	return nil
}
