package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aelhady03/twerlo-chat-app/internal/config"
	"github.com/aelhady03/twerlo-chat-app/internal/models"
)

type MediaHandler struct {
	config *config.Config
}

func NewMediaHandler(config *config.Config) *MediaHandler {
	return &MediaHandler{
		config: config,
	}
}

// UploadMedia handles file uploads
func (h *MediaHandler) UploadMedia(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	claims, err := getUserFromContext(r.Context())
	if err != nil {
		writeErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	// Parse multipart form
	err = r.ParseMultipartForm(h.config.Upload.MaxSize)
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_FORM", "Invalid multipart form")
		return
	}

	// Get file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "MISSING_FILE", "No file provided")
		return
	}
	defer file.Close()

	// Validate file size
	if header.Size > h.config.Upload.MaxSize {
		writeErrorResponse(w, http.StatusBadRequest, "FILE_TOO_LARGE",
			fmt.Sprintf("File size exceeds maximum allowed size of %d bytes", h.config.Upload.MaxSize))
		return
	}

	// Validate file type
	allowedTypes := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".pdf":  true,
		".doc":  true,
		".docx": true,
		".txt":  true,
		".mp4":  true,
		".avi":  true,
		".mov":  true,
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedTypes[ext] {
		writeErrorResponse(w, http.StatusBadRequest, "INVALID_FILE_TYPE",
			"File type not allowed. Allowed types: jpg, jpeg, png, gif, pdf, doc, docx, txt, mp4, avi, mov")
		return
	}

	// Generate unique filename
	filename := fmt.Sprintf("%s_%s_%s",
		claims.UserID.String(),
		time.Now().Format("20060102_150405"),
		header.Filename)

	// Create upload directory if it doesn't exist
	uploadDir := h.config.Upload.Path
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "UPLOAD_FAILED", "Failed to create upload directory")
		return
	}

	// Create destination file
	filePath := filepath.Join(uploadDir, filename)
	dst, err := os.Create(filePath)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "UPLOAD_FAILED", "Failed to create destination file")
		return
	}
	defer dst.Close()

	// Copy file content
	_, err = io.Copy(dst, file)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "UPLOAD_FAILED", "Failed to save file")
		return
	}

	// Determine file type
	fileType := "file"
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif":
		fileType = "image"
	case ".mp4", ".avi", ".mov":
		fileType = "video"
	}

	// Create response
	response := models.UploadResponse{
		Filename: filename,
		URL:      fmt.Sprintf("/uploads/%s", filename),
		Size:     header.Size,
		Type:     fileType,
	}

	writeSuccessResponse(w, http.StatusOK, "File uploaded successfully", response)
}

// ServeMedia serves uploaded media files
func (h *MediaHandler) ServeMedia(w http.ResponseWriter, r *http.Request) {
	// Get filename from URL
	filename := r.URL.Path[len("/uploads/"):]
	if filename == "" {
		http.NotFound(w, r)
		return
	}

	// Construct file path
	filePath := filepath.Join(h.config.Upload.Path, filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}

	// Serve file
	http.ServeFile(w, r, filePath)
}
