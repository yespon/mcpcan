package biz

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	pb "github.com/kymo-mcp/mcpcan/api/market/ai_agent"
)

type AiFileManager struct {
	basePath string
	domain   string
}

var GAiFileManager *AiFileManager

func init() {
	// Initialize with default local path, configuration can override this later if needed
	GAiFileManager = NewAiFileManager("data/chat-files", "")
}

func NewAiFileManager(basePath string, domain string) *AiFileManager {
	// Ensure base directory exists
	if err := os.MkdirAll(basePath, 0755); err != nil {
		fmt.Printf("Failed to create base path %s: %v\n", basePath, err)
	}
	return &AiFileManager{
		basePath: basePath,
		domain:   domain,
	}
}

// DeleteFiles deletes multiple files by their paths or URLs
func (m *AiFileManager) DeleteFiles(files []string) error {
	for _, file := range files {
		// Basic sanitization/resolution
		// If it's a URL like /files/..., convert to path
		// If it's a full path, just delete
		// Current implementation of Upload returns:
		// Id: abs path, Url: relative url
		
		targetPath := file
		if strings.HasPrefix(file, "/files/chat-files/") {
			relPath := strings.TrimPrefix(file, "/files/chat-files/")
			targetPath = filepath.Join(m.basePath, relPath)
		} else if strings.HasPrefix(file, "/files/") {
			// Convert web path to fs path: /files/2024/... -> data/chat-files/2024/...
			relPath := strings.TrimPrefix(file, "/files/")
			targetPath = filepath.Join(m.basePath, relPath)
		}

		// Security check: ensure within base path
		// rel, err := filepath.Rel(m.basePath, targetPath)
		// if err != nil || strings.HasPrefix(rel, "..") { continue }

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

// UploadFile handles the file upload logic
func (m *AiFileManager) UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader) (*pb.UploadFileResponse, error) {
	// 1. Generate unique file ID and name
	fileID := uuid.New().String()
	ext := filepath.Ext(header.Filename)
	newFileName := fmt.Sprintf("%s%s", fileID, ext)

	// 2. Create directory (organized by date)
	dateDir := time.Now().Format("2006/01/02")
	relDir := filepath.Join("chat-files", dateDir) // relative path for URL construction
	absDir := filepath.Join(m.basePath, dateDir)   // absolute path for file storage

	if err := os.MkdirAll(absDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// 3. Save file
	dstPath := filepath.Join(absDir, newFileName)
	dst, err := os.Create(dstPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	// 4. Construct response
	// URL construction logic: assuming there will be a static file handler serving 'data' or similar
	// For now, return a relative URL that the frontend or backend can resolve
	// TODO: Replace with configured domain/CDN URL
	url := fmt.Sprintf("/files/%s/%s", relDir, newFileName)
	
	// If domain is configured, prepend it
	if m.domain != "" {
		url = strings.TrimRight(m.domain, "/") + url
	} else {
		// Try to use common helper if available or default relative
		// For local dev, relative is usually fine if proxied correctly
	}
	
	// Clean up URL path separators for web
	url = strings.ReplaceAll(url, "\\", "/")

	absPath, _ := filepath.Abs(dstPath)
	return &pb.UploadFileResponse{
		Id:   absPath, // Return absolute path for robust access (e.g. deletion, testing)
		Url:  url,
		Name: header.Filename,
	}, nil
}

// GetFileContent reads file content for LLM ingestion (e.g. base64 encoding)
func (m *AiFileManager) GetFileContent(path string) ([]byte, error) {
	// Security check: ensure path is within basePath
	// This is a basic check, in production use more robust path validation
	// abspath, _ := filepath.Abs(path)
	// if !strings.HasPrefix(abspath, m.basePath) { ... }
	
	return os.ReadFile(path)
}

// DetectContentType detects mime type from file content
func (m *AiFileManager) DetectContentType(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)
	_, err = f.Read(buffer)
	if err != nil && err != io.EOF {
		return "", err
	}

	return http.DetectContentType(buffer), nil
}

// Helper to check if file is an image
func (m *AiFileManager) IsImage(path string) bool {
	contentType, err := m.DetectContentType(path)
	if err != nil {
		return false
	}
	return strings.HasPrefix(contentType, "image/")
}
