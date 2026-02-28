package biz

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	pb "github.com/kymo-mcp/mcpcan/api/market/ai_agent"
)

// ChatImageSubDir 图片聊天文件存放的子目录（相对于 staticPath）
const ChatImageSubDir = "images/chat"

// ChatImageURLPrefix 静态文件对外访问的 URL 前缀
const ChatImageURLPrefix = "/static/images/chat"

type AiFileManager struct {
	// uploadDir 是处理后的实际存储路径（staticPath + images/chat）
	uploadDir string
}

var GAiFileManager *AiFileManager

// NewAiFileManager 使用配置中的 staticPath 创建文件管理器
func NewAiFileManager(staticPath string) *AiFileManager {
	if staticPath == "" {
		panic("staticPath is required for AiFileManager")
	}

	uploadDir := filepath.Join(staticPath, ChatImageSubDir)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		fmt.Printf("Failed to create chat image directory %s: %v\n", uploadDir, err)
	}
	return &AiFileManager{uploadDir: uploadDir}
}

// UploadChatImage 上传聊天图片
//
//   - sessionId: 会话 ID，用于生成文件名前缀
//   - file:      multipart 文件内容
//   - header:    multipart 文件头（用于获取原始文件名）
//
// 返回:
//   - RelativePath: /static/images/chat/{sessionId}-{filename}  —— 存入数据库及返回给前端的值
func (m *AiFileManager) UploadChatImage(
	_ context.Context,
	sessionId int64,
	file multipart.File,
	header *multipart.FileHeader,
) (*pb.UploadFileResponse, error) {
	// 1. 确保目录存在
	if err := os.MkdirAll(m.uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create chat image directory: %w", err)
	}

	// 2. 构造文件名: {sessionId}-{originalFilename}
	originalName := filepath.Base(header.Filename)
	// 去掉路径中的目录分隔符防注入
	originalName = strings.ReplaceAll(originalName, string(filepath.Separator), "_")
	newFileName := fmt.Sprintf("%d-%s", sessionId, originalName)

	// 3. 写入磁盘
	dstPath := filepath.Join(m.uploadDir, newFileName)
	dst, err := os.Create(dstPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	// 4. 构造相对路径（存数据库/返回给前端）
	// 格式: /static/images/chat/{sessionId}-{filename}
	relativePath := fmt.Sprintf("%s/%s", ChatImageURLPrefix, newFileName)
	relativePath = strings.ReplaceAll(relativePath, "\\", "/")

	return &pb.UploadFileResponse{
		Id:   relativePath, // 相对路径，用于删除时定位
		Url:  relativePath, // 存入数据库/返回给前端的相对路径
		Name: header.Filename,
	}, nil
}

// DeleteFiles 根据相对路径删除文件
//
// 支持格式: /static/images/chat/{filename}
func (m *AiFileManager) DeleteFiles(files []string) error {
	for _, file := range files {
		var relPath string
		if strings.HasPrefix(file, ChatImageURLPrefix) {
			// /static/images/chat/filename -> images/chat/filename
			relPath = strings.TrimPrefix(file, "/static/")
		} else if strings.HasPrefix(file, "/files/chat-files/") {
			// 兼容旧路径格式
			relPath = strings.TrimPrefix(file, "/files/")
		} else {
			// 按文件系统路径直接删除
			_ = os.Remove(file)
			continue
		}

		// 提取文件名
		fileName := filepath.Base(relPath)
		targetPath := filepath.Join(m.uploadDir, fileName)

		if err := os.Remove(targetPath); err != nil {
			if !os.IsNotExist(err) {
				fmt.Printf("Failed to delete file %s: %v\n", targetPath, err)
			}
		} else {
			fmt.Printf("Deleted file: %s\n", targetPath)
		}
	}
	return nil
}

