package app

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"go.uber.org/zap"

	"github.com/kymo-mcp/mcpcan/pkg/common"
	"github.com/kymo-mcp/mcpcan/pkg/database/model"
	"github.com/kymo-mcp/mcpcan/pkg/database/repository/mysql"
	"github.com/kymo-mcp/mcpcan/pkg/logger"
	"github.com/kymo-mcp/mcpcan/pkg/openapifile"
)

// initOpenapi initializes OpenAPI document
func (a *App) initOpenapi(ctx context.Context) error {
	logger.Info("Starting OpenAPI document initialization")

	apiFile := "./init-data/openapi-file/mcpcan-openapi.yaml"
	if _, err := os.Stat(apiFile); os.IsNotExist(err) {
		logger.Info("OpenAPI document source file does not exist, skipping initialization",
			zap.String("apiFile", apiFile))
		return nil
	}

	fileName := filepath.Base(apiFile)

	codeConfig := &common.CodeConfig{
		Upload: common.UploadConfig{
			MaxFileSize: 100,
		},
	}
	fileManager := openapifile.NewOpenapiFileManager(codeConfig, a.config.Storage.OpenapiFilePath)

	// 1. Query database first to check if record exists
	existingPackages, err := mysql.McpOpenapiPackageRepo.FindAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to query openapi packages: %w", err)
	}

	// 2. Check if same file already exists in database
	var existingPkg *model.McpOpenapiPackage
	for _, pkg := range existingPackages {
		logger.Debug("Checking existing package",
			zap.String("pkgOriginalName", pkg.OriginalName),
			zap.String("pkgBaseOpenapiFileID", pkg.BaseOpenapiFileID),
			zap.String("targetFileName", fileName))
		if pkg.OriginalName == fileName && pkg.BaseOpenapiFileID == "" {
			existingPkg = pkg
			break
		}
	}

	logger.Info("Package lookup result",
		zap.Int("totalPackages", len(existingPackages)),
		zap.Bool("foundExisting", existingPkg != nil))

	// 3. If exists in database
	if existingPkg != nil {
		// Convert relative path to absolute path
		absFilePath, err := fileManager.ToAbsolutePath(existingPkg.OpenapiFilePath)
		if err != nil {
			return fmt.Errorf("failed to convert to absolute path: %w", err)
		}

		// Check if file exists on disk
		if _, err := os.Stat(absFilePath); err == nil {
			// File exists, skip completely
			logger.Info("OpenAPI package already exists with file, skipping initialization",
				zap.String("openapiFileId", existingPkg.OpenapiFileID),
				zap.String("filePath", absFilePath))
			return nil
		}

		// File missing, only copy file to original path, don't update database
		logger.Info("OpenAPI record exists but file missing, restoring file",
			zap.String("openapiFileId", existingPkg.OpenapiFileID),
			zap.String("filePath", absFilePath))

		if err := copyFile(apiFile, absFilePath); err != nil {
			return fmt.Errorf("failed to restore openapi file: %w", err)
		}

		logger.Info("OpenAPI file restored successfully",
			zap.String("filePath", absFilePath))
		return nil
	}

	// 4. Not in database, upload and create record
	file, err := os.Open(apiFile)
	if err != nil {
		return fmt.Errorf("failed to open openapi file: %w", err)
	}
	defer file.Close()

	fileStat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat openapi file: %w", err)
	}

	header := &multipart.FileHeader{
		Filename: fileName,
		Size:     fileStat.Size(),
		Header:   make(map[string][]string),
	}
	fileReader := &fileReaderWrapper{file: file}

	fileInfo, err := fileManager.UploadOpenapiFile(fileReader, header)
	if err != nil {
		return fmt.Errorf("failed to upload openapi file: %w", err)
	}

	if err := fileManager.ValidateOpenapiFile(fileInfo.OpenapiFilePath); err != nil {
		fileManager.DeleteOpenapiFile(fileInfo.OpenapiFilePath)
		return fmt.Errorf("failed to validate openapi file: %w", err)
	}

	openapiPackage := &model.McpOpenapiPackage{
		OpenapiFileID:     fileInfo.OpenapiFileID,
		OpenapiFileType:   fileInfo.OpenapiFileType,
		OpenapiFilePath:   fileInfo.OpenapiFilePath,
		OriginalName:      fileInfo.OriginalName,
		FileSize:          fileInfo.FileSize,
		BaseOpenapiFileID: "",
	}

	if err := mysql.McpOpenapiPackageRepo.Create(ctx, openapiPackage); err != nil {
		fileManager.DeleteOpenapiFile(fileInfo.OpenapiFilePath)
		return fmt.Errorf("failed to save openapi package: %w", err)
	}

	logger.Info("OpenAPI package initialized successfully",
		zap.String("openapiFileId", fileInfo.OpenapiFileID),
		zap.String("filePath", fileInfo.OpenapiFilePath))

	return nil
}

// copyFile copies file from src to dst
func copyFile(src, dst string) error {
	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}
