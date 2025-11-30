package services

import (
	"fmt"
	"io"
	"math/rand"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FileUploadResponse contains the response data from a file upload
type FileUploadResponse struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
}

type MediaService struct{}

// NewMediaService creates a new MediaService instance
func NewMediaService() *MediaService {
	return &MediaService{}
}

// UploadFile handles file upload validation and storage
func (m *MediaService) UploadFile(file *multipart.FileHeader) (*FileUploadResponse, error) {
	// Validate file type
	contentType := file.Header.Get("Content-Type")
	if err := m.validateFileType(contentType); err != nil {
		return nil, err
	}

	// Validate file size based on content type
	if err := m.validateFileSize(contentType, file.Size); err != nil {
		return nil, err
	}

	// Create uploads directory if it doesn't exist
	uploadsDir := "./uploads"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create uploads directory: %w", err)
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d_%s%s", time.Now().Unix(), m.generateRandomString(10), ext)
	filePath := filepath.Join(uploadsDir, filename)

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}
	defer dst.Close()

	// Copy file contents
	if _, err := io.Copy(dst, src); err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	// Return file URL and filename
	fileURL := fmt.Sprintf("/uploads/%s", filename)
	return &FileUploadResponse{
		URL:      fileURL,
		Filename: filename,
	}, nil
}

// validateFileType checks if the file type is allowed
func (m *MediaService) validateFileType(contentType string) error {
	allowedTypes := []string{
		// Images
		"image/jpeg", "image/jpg", "image/png", "image/gif", "image/webp", "image/svg+xml",
		// Videos
		"video/mp4", "video/webm", "video/ogg", "video/quicktime",
		// Documents
		"application/pdf",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.ms-excel",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"application/vnd.ms-powerpoint",
		"application/vnd.openxmlformats-officedocument.presentationml.presentation",
		"text/plain",
		"text/csv",
	}

	if !m.contains(allowedTypes, contentType) {
		return fmt.Errorf("invalid file type. Allowed types: images, videos, PDFs, and documents")
	}

	return nil
}

// validateFileSize checks if the file size is within limits based on type
func (m *MediaService) validateFileSize(contentType string, size int64) error {
	var maxSize int64

	// Determine max size based on content type
	if m.isVideoType(contentType) {
		maxSize = int64(50 * 1024 * 1024) // 50MB for videos
	} else if m.isDocumentType(contentType) {
		maxSize = int64(10 * 1024 * 1024) // 10MB for documents/PDFs
	} else {
		maxSize = int64(5 * 1024 * 1024) // 5MB for images and others
	}

	if size > maxSize {
		sizeMB := maxSize / (1024 * 1024)
		return fmt.Errorf("file size exceeds %dMB limit", sizeMB)
	}

	return nil
}

// isVideoType checks if the content type is a video
func (m *MediaService) isVideoType(contentType string) bool {
	videoTypes := []string{
		"video/mp4", "video/webm", "video/ogg", "video/quicktime",
	}
	return m.contains(videoTypes, contentType)
}

// isDocumentType checks if the content type is a document or PDF
func (m *MediaService) isDocumentType(contentType string) bool {
	if contentType == "application/pdf" {
		return true
	}
	if strings.HasPrefix(contentType, "application/vnd.") {
		return true
	}
	if strings.HasPrefix(contentType, "application/msword") {
		return true
	}
	return false
}

// contains checks if a string exists in a slice
func (m *MediaService) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// generateRandomString generates a random string of the specified length
func (m *MediaService) generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return strings.ToLower(string(b))
}
