package service

import (
	"context"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kymo-mcp/mcpcan/api/market/openapi_file"
	"github.com/kymo-mcp/mcpcan/internal/market/config"
	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
	i18nresp "github.com/kymo-mcp/mcpcan/pkg/i18n"
	"github.com/kymo-mcp/mcpcan/pkg/logger"
	"github.com/kymo-mcp/mcpcan/pkg/openapifile"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// OpenapiService provides OpenAPI document management services
type OpenapiService struct {
	openapiPackageRepo *mysql.McpOpenapiPackageRepository
	fileManager        *openapifile.OpenapiFileManager
}

// NewOpenapiService creates a new OpenapiService instance
func NewOpenapiService() *OpenapiService {
	return &OpenapiService{
		openapiPackageRepo: mysql.McpOpenapiPackageRepo,
		fileManager:        openapifile.NewOpenapiFileManager(&config.GlobalConfig.Code, config.GlobalConfig.Storage.OpenapiFilePath),
	}
}

// UploadOpenapiFile uploads an OpenAPI document
func (s *OpenapiService) UploadOpenapiFile(c *gin.Context) {
	// Block upload in demo mode
	if config.IsDemoMode() {
		common.GinError(c, i18nresp.CodeForbidden, "operation forbidden in demo mode")
		return
	}
	// Record upload start time
	startTime := time.Now()
	logger.Info("Starting OpenAPI document upload request",
		zap.String("client_ip", c.ClientIP()),
		zap.String("request_id", c.GetString("RequestID")),
		zap.String("content_type", c.ContentType()))

	// Get uploaded file
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		logger.Error("Failed to get uploaded file",
			zap.Error(err),
			zap.String("client_ip", c.ClientIP()),
			zap.String("request_id", c.GetString("RequestID")))
		common.GinError(c, i18nresp.OpenapiFileUploadFailed, "failed to get uploaded file")
		return
	}
	defer file.Close()

	// Record detailed information of uploaded file
	logger.Info("File received for upload",
		zap.String("filename", header.Filename),
		zap.Int64("size", header.Size),
		zap.Int("configured_max_size_mb", config.GlobalConfig.Code.Upload.MaxFileSize),
		zap.Float64("size_mb", float64(header.Size)/(1024*1024)),
		zap.String("content_type", header.Header.Get("Content-Type")))

	// Use OpenAPI file manager to handle upload
	fileInfo, err := s.fileManager.UploadOpenapiFile(file, header)
	if err != nil {
		logger.Error("Failed to upload OpenAPI document", zap.Error(err))
		// 根据错误类型返回不同的错误码
		if strings.Contains(err.Error(), "unsupported file type") {
			common.GinError(c, i18nresp.OpenapiFileTypeNotSupported, err.Error())
		} else if strings.Contains(err.Error(), "file size") {
			common.GinError(c, i18nresp.OpenapiFileSizeExceeded, err.Error())
		} else {
			common.GinError(c, i18nresp.OpenapiFileUploadFailed, err.Error())
		}
		return
	}

	// Validate the uploaded OpenAPI file
	if err := s.fileManager.ValidateOpenapiFile(fileInfo.OpenapiFilePath); err != nil {
		logger.Error("OpenAPI file validation failed", zap.Error(err))
		// Clean up uploaded file
		s.fileManager.DeleteOpenapiFile(fileInfo.OpenapiFilePath)
		common.GinError(c, i18nresp.OpenapiFileValidationFailed, fmt.Sprintf("invalid OpenAPI document: %v", err))
		return
	}

	ctx := context.Background()

	// Save to database
	openapiPackage := &model.McpOpenapiPackage{
		OpenapiFileID:     fileInfo.OpenapiFileID,
		OpenapiFileType:   fileInfo.OpenapiFileType,
		OpenapiFilePath:   fileInfo.OpenapiFilePath,
		OriginalName:      fileInfo.OriginalName,
		FileSize:          fileInfo.FileSize,
		BaseOpenapiFileID: c.Request.FormValue("baseOpenapiFileID"),
	}

	if err := s.openapiPackageRepo.Create(ctx, openapiPackage); err != nil {
		logger.Error("Failed to save OpenAPI document to database", zap.Error(err))
		// Clean up uploaded file
		s.fileManager.DeleteOpenapiFile(fileInfo.OpenapiFilePath)
		// 检查是否是重复文件
		if strings.Contains(err.Error(), "Duplicate") || strings.Contains(err.Error(), "duplicate") {
			common.GinError(c, i18nresp.OpenapiFileDuplicate, "OpenAPI document already exists")
		} else {
			common.GinError(c, i18nresp.OpenapiFileUploadFailed, "failed to save OpenAPI document information")
		}
		return
	}

	// Calculate total elapsed time
	totalElapsed := time.Since(startTime)
	logger.Info("OpenAPI document uploaded successfully",
		zap.String("openapiFileId", fileInfo.OpenapiFileID),
		zap.String("filename", fileInfo.OriginalName),
		zap.String("filePath", fileInfo.OpenapiFilePath),
		zap.String("fileType", string(fileInfo.OpenapiFileType)),
		zap.Duration("total_elapsed", totalElapsed),
		zap.Float64("total_elapsed_seconds", totalElapsed.Seconds()))

	common.GinSuccess(c, &openapi_file.UploadOpenapiFileResponse{
		OpenapiFileId: fileInfo.OpenapiFileID,
		FilePath:      fileInfo.OpenapiFilePath,
		FileType:      convertOpenapiFileType(fileInfo.OpenapiFileType),
	})
}

// GetOpenapiFileContent retrieves the content of an OpenAPI document
func (s *OpenapiService) GetOpenapiFileContent(c *gin.Context) {
	var req openapi_file.GetOpenapiFileContentRequest
	if err := common.BindAndValidateQuery(c, &req); err != nil {
		return
	}

	ctx := context.Background()

	// Find OpenAPI document
	openapiPackage, err := s.openapiPackageRepo.FindByOpenapiFileID(ctx, req.OpenapiFileId)
	if err != nil {
		logger.Error("Failed to find OpenAPI document", zap.String("openapiFileId", req.OpenapiFileId), zap.Error(err))
		common.GinError(c, i18nresp.OpenapiFileNotFound, "OpenAPI document not found")
		return
	}

	// Use file manager to get file content
	content, err := s.fileManager.GetOpenapiFileContent(openapiPackage.OpenapiFilePath)
	if err != nil {
		logger.Error("Failed to read OpenAPI file content",
			zap.String("openapiFileId", req.OpenapiFileId),
			zap.String("filePath", openapiPackage.OpenapiFilePath),
			zap.Error(err))
		if os.IsNotExist(err) {
			common.GinError(c, i18nresp.OpenapiFileNotFound, "OpenAPI document file not found")
		} else {
			common.GinError(c, i18nresp.OpenapiFileParseError, "failed to read OpenAPI document")
		}
		return
	}

	// 检查内容是否为空
	if content == "" {
		common.GinError(c, i18nresp.OpenapiFileContentEmpty, "OpenAPI document content is empty")
		return
	}

	common.GinSuccess(c, &openapi_file.GetOpenapiFileContentResponse{
		Content:           content,
		BaseOpenapiFileID: openapiPackage.BaseOpenapiFileID,
		FileType:          convertOpenapiFileType(openapiPackage.OpenapiFileType),
	})
}

// EditOpenapiFile edits an OpenAPI document
func (s *OpenapiService) EditOpenapiFile(c *gin.Context) {
	var req openapi_file.EditOpenapiFileRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		return
	}

	// 检查内容是否为空
	if req.Content == "" {
		common.GinError(c, i18nresp.OpenapiFileContentEmpty, "OpenAPI document content cannot be empty")
		return
	}

	ctx := context.Background()

	// Find OpenAPI document
	openapiPackage, err := s.openapiPackageRepo.FindByOpenapiFileID(ctx, req.OpenapiFileId)
	if err != nil {
		logger.Error("Failed to find OpenAPI document", zap.String("openapiFileId", req.OpenapiFileId), zap.Error(err))
		common.GinError(c, i18nresp.OpenapiFileNotFound, "OpenAPI document not found")
		return
	}

	// Use file manager to update file content
	if err := s.fileManager.UpdateOpenapiFileContent(openapiPackage.OpenapiFilePath, req.Content); err != nil {
		logger.Error("Failed to update OpenAPI file content",
			zap.String("openapiFileId", req.OpenapiFileId),
			zap.String("filePath", openapiPackage.OpenapiFilePath),
			zap.Error(err))
		common.GinError(c, i18nresp.OpenapiFileUpdateFailed, "failed to update OpenAPI document")
		return
	}

	// Validate the updated OpenAPI file
	if err := s.fileManager.ValidateOpenapiFile(openapiPackage.OpenapiFilePath); err != nil {
		logger.Error("OpenAPI file validation failed after edit", zap.Error(err))
		common.GinError(c, i18nresp.OpenapiFileValidationFailed, fmt.Sprintf("invalid OpenAPI document after edit: %v", err))
		return
	}

	// Update file size in database
	fileInfo, err := s.fileManager.GetFileInfo(openapiPackage.OpenapiFilePath)
	if err == nil && fileInfo != nil {
		openapiPackage.FileSize = fileInfo.FileSize
		openapiPackage.PrepareForUpdate()
		if err := s.openapiPackageRepo.Update(ctx, openapiPackage); err != nil {
			logger.Warn("Failed to update file size in database", zap.Error(err))
		}
	}

	common.GinSuccess(c, &openapi_file.EditOpenapiFileResponse{
		Success: true,
		Message: "OpenAPI document edited successfully",
	})
}

// DownloadOpenapiFile handles OpenAPI document download requests
func (s *OpenapiService) DownloadOpenapiFile(c *gin.Context) {
	req := &openapi_file.DownloadOpenapiFileRequest{}
	if err := common.BindAndValidateQuery(c, req); err != nil {
		return
	}
	openapiFileID := req.OpenapiFileId

	// Parameter validation
	if openapiFileID == "" {
		common.GinError(c, i18nresp.CodeBadRequest, "OpenAPI file ID is required")
		return
	}

	// Find OpenAPI document
	openapiPackage, err := s.openapiPackageRepo.FindByOpenapiFileID(c, openapiFileID)
	if err != nil {
		logger.Error("Failed to find OpenAPI document", zap.String("openapiFileId", openapiFileID), zap.Error(err))
		common.GinError(c, i18nresp.OpenapiFileNotFound, "OpenAPI document not found")
		return
	}

	// Convert relative path to absolute path
	absFilePath, err := s.fileManager.ToAbsolutePath(openapiPackage.OpenapiFilePath)
	if err != nil {
		logger.Error("Failed to convert to absolute path",
			zap.String("relativePath", openapiPackage.OpenapiFilePath),
			zap.Error(err))
		common.GinError(c, i18nresp.OpenapiFileNotFound, "invalid file path")
		return
	}

	// Check if file exists
	if _, err := os.Stat(absFilePath); os.IsNotExist(err) {
		logger.Error("File not found", zap.String("filePath", absFilePath), zap.Error(err))
		common.GinError(c, i18nresp.OpenapiFileNotFound, "OpenAPI document file not found")
		return
	}

	// Determine the actual filename to use for download
	downloadFileName := openapiPackage.OriginalName

	// Set proper Content-Type based on file extension
	ext := filepath.Ext(downloadFileName)
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		// Default to application/octet-stream for unknown file types
		contentType = "application/octet-stream"
	}

	// Set response headers for proper file download
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", downloadFileName))
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	// Serve the file
	c.File(absFilePath)
}

// GetOpenapiFileList gets OpenAPI document list
func (s *OpenapiService) GetOpenapiFileList(c *gin.Context) {
	var req openapi_file.OpenapiFileListRequest
	if err := common.BindAndValidateQuery(c, &req); err != nil {
		logger.Error("Failed to bind request", zap.Error(err))
		common.GinError(c, i18nresp.CodeBadRequest, "invalid request parameters")
		return
	}

	// Set default values
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	filters := map[string]interface{}{}
	if req.Name != "" {
		filters["name"] = req.Name
	}
	if len(req.Types) > 0 {
		// Convert to model type
		var fileTypes []model.OpenapiFileType
		for _, t := range req.Types {
			modelType, _ := common.ConvertToModelOpenapiFileType(t)
			fileTypes = append(fileTypes, modelType)
		}
		if len(fileTypes) > 0 {
			filters["types"] = fileTypes
		}
	}
	// Query OpenAPI document list
	packages, total, err := s.openapiPackageRepo.FindWithPagination(c.Request.Context(), req.Page, req.PageSize, filters)
	if err != nil {
		logger.Error("Failed to query OpenAPI documents", zap.Error(err))
		common.GinError(c, i18nresp.CodeInternalError, "failed to query OpenAPI documents")
		return
	}

	// Convert to response format
	var fileList []*openapi_file.OpenapiFileInfo
	for _, pkg := range packages {
		fileInfo := &openapi_file.OpenapiFileInfo{
			Id:                pkg.OpenapiFileID,
			Name:              pkg.OriginalName,
			Path:              pkg.OpenapiFilePath,
			Size:              pkg.FileSize,
			Type:              convertOpenapiFileType(pkg.OpenapiFileType),
			CreatedAt:         pkg.CreatedAt.String(),
			UpdatedAt:         pkg.UpdatedAt.String(),
			BaseOpenapiFileID: pkg.BaseOpenapiFileID,
		}
		fileList = append(fileList, fileInfo)
	}

	response := &openapi_file.OpenapiFileListResponse{
		List:     fileList,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}
	common.GinSuccess(c, response)
}

// DeleteOpenapiFile deletes an OpenAPI document and its associated files
func (s *OpenapiService) DeleteOpenapiFile(c *gin.Context) {
	// Block deletion in demo mode
	if config.IsDemoMode() {
		common.GinError(c, i18nresp.CodeForbidden, "operation forbidden in demo mode")
		return
	}
	var req openapi_file.DeleteOpenapiFileRequest
	if err := common.BindAndValidate(c, &req); err != nil {
		logger.Error("Failed to bind request", zap.Error(err))
		common.GinError(c, i18nresp.CodeBadRequest, "invalid request parameters")
		return
	}

	// Validate file ID
	if req.OpenapiFileId == "" {
		logger.Warn("Empty OpenAPI file ID provided for deletion")
		common.GinError(c, i18nresp.CodeBadRequest, "OpenAPI file ID is required")
		return
	}

	ctx := context.Background()

	// Find the OpenAPI document
	openapiPackage, err := s.openapiPackageRepo.FindByOpenapiFileID(ctx, req.OpenapiFileId)
	if err != nil {
		logger.Error("Failed to find OpenAPI document", zap.String("openapiFileId", req.OpenapiFileId), zap.Error(err))
		common.GinError(c, i18nresp.OpenapiFileNotFound, "OpenAPI document not found")
		return
	}

	if openapiPackage == nil {
		logger.Warn("OpenAPI document not found", zap.String("openapiFileId", req.OpenapiFileId))
		common.GinError(c, i18nresp.OpenapiFileNotFound, "OpenAPI document not found")
		return
	}

	// Delete physical files using file manager
	if openapiPackage.OpenapiFilePath != "" {
		if err := s.fileManager.DeleteOpenapiFile(openapiPackage.OpenapiFilePath); err != nil {
			logger.Error("Failed to delete OpenAPI document files",
				zap.String("openapiFileId", req.OpenapiFileId),
				zap.String("filePath", openapiPackage.OpenapiFilePath),
				zap.Error(err))
			// Continue with database deletion even if file deletion fails
			logger.Warn("Continuing with database deletion despite file deletion failure")
		}
	}

	// Delete database record
	if err := s.openapiPackageRepo.DeleteByOpenapiFileID(ctx, req.OpenapiFileId); err != nil {
		logger.Error("Failed to delete OpenAPI document from database",
			zap.String("openapiFileId", req.OpenapiFileId),
			zap.Error(err))
		common.GinError(c, i18nresp.OpenapiFileDeleteFailed, "failed to delete OpenAPI document record")
		return
	}

	logger.Info("OpenAPI document deleted successfully", zap.String("openapiFileId", req.OpenapiFileId))

	go func() {
		openapiFileList, _, err := s.openapiPackageRepo.FindWithPagination(context.Background(), 1, 99999, map[string]interface{}{
			"baseOpenapiFileID": req.OpenapiFileId,
		})
		if err != nil {
			logger.Error("Failed to list openapiFile by baseOpenapiFileID ", zap.Error(err),
				zap.String("baseOpenapiFileID", req.OpenapiFileId),
				zap.Error(err))
			return
		}
		for _, openapiFile := range openapiFileList {
			if err := s.fileManager.DeleteOpenapiFile(openapiFile.OpenapiFilePath); err != nil {
				logger.Error("Failed to delete OpenAPI document files",
					zap.String("openapiFileId", openapiFile.OpenapiFileID),
					zap.String("filePath", openapiFile.OpenapiFilePath),
					zap.Error(err))
			}
			if err := s.openapiPackageRepo.DeleteByOpenapiFileID(context.Background(), openapiFile.OpenapiFileID); err != nil {
				logger.Error("Failed to delete OpenAPI document from database",
					zap.String("openapiFileId", openapiFile.OpenapiFileID),
					zap.Error(err))
			}
		}
	}()

	// Return success response
	response := &openapi_file.DeleteOpenapiFileResponse{
		Success: true,
		Message: "OpenAPI document deleted successfully",
	}

	common.GinSuccess(c, response)
}

// convertOpenapiFileType converts OpenAPI file type
func convertOpenapiFileType(modelType model.OpenapiFileType) openapi_file.OpenapiFileType {
	switch modelType {
	case model.OpenapiFileTypeJson:
		return openapi_file.OpenapiFileType_OpenapiFileTypeJson
	case model.OpenapiFileTypeYaml:
		return openapi_file.OpenapiFileType_OpenapiFileTypeYaml
	default:
		return openapi_file.OpenapiFileType_OpenapiFileTypeUnspecified
	}
}
