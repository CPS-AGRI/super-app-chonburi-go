package http

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"super-app-chonburi-go/config"
	"super-app-chonburi-go/pkg/jwtutil"
	storage "super-app-chonburi-go/pkg/storage/minio"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type UploadHandler struct {
	storage        *storage.Client
	maxUploadBytes int64
}

type UploadResponse struct {
	Bucket      string `json:"bucket"`
	ObjectKey   string `json:"object_key"`
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
	URL         string `json:"url"`
	Public      bool   `json:"public"`
}

func NewUploadHandler(storage *storage.Client, cfg config.MinIOConfig) *UploadHandler {
	maxUploadSizeMB := cfg.MaxUploadSizeMB
	if maxUploadSizeMB <= 0 {
		maxUploadSizeMB = 10
	}

	return &UploadHandler{
		storage:        storage,
		maxUploadBytes: maxUploadSizeMB * 1024 * 1024,
	}
}

func (h *UploadHandler) RegisterRoutes(app *fiber.App) {
	upload := app.Group("/api/v1/uploads")
	upload.Use(jwtutil.RequireAuth())
	upload.Post("/images", h.UploadImage)
}

func (h *UploadHandler) UploadImage(c fiber.Ctx) error {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return ErrorResponse(c, "Image file is required", fiber.StatusBadRequest)
	}

	if fileHeader.Size <= 0 {
		return ErrorResponse(c, "Image file is empty", fiber.StatusBadRequest)
	}
	if fileHeader.Size > h.maxUploadBytes {
		return ErrorResponse(c, "Image file is too large", fiber.StatusRequestEntityTooLarge)
	}

	file, err := fileHeader.Open()
	if err != nil {
		return ErrorResponse(c, "Failed to open image file")
	}
	defer file.Close()

	contentType, err := detectImageContentType(file)
	if err != nil {
		return ErrorResponse(c, err.Error(), fiber.StatusBadRequest)
	}

	objectKey := buildImageObjectKey(fileHeader, contentType)
	info, err := h.storage.PutObject(c.Context(), objectKey, file, fileHeader.Size, contentType)
	if err != nil {
		return ErrorResponse(c, "Failed to upload image")
	}

	url, err := h.storage.ObjectURL(c.Context(), info.Key)
	if err != nil {
		return ErrorResponse(c, "Failed to create image URL")
	}

	return SuccessResponse(c, UploadResponse{
		Bucket:      info.Bucket,
		ObjectKey:   info.Key,
		FileName:    fileHeader.Filename,
		ContentType: contentType,
		Size:        info.Size,
		URL:         url,
		Public:      h.storage.PublicRead,
	}, fiber.StatusCreated)
}

func detectImageContentType(file multipart.File) (string, error) {
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil {
		return "", fmt.Errorf("failed to read image file")
	}

	if _, err := file.Seek(0, 0); err != nil {
		return "", fmt.Errorf("failed to reset image file")
	}

	contentType := http.DetectContentType(buffer[:n])
	switch contentType {
	case "image/jpeg", "image/png", "image/webp", "image/gif":
		return contentType, nil
	default:
		return "", fmt.Errorf("unsupported image type")
	}
}

func buildImageObjectKey(fileHeader *multipart.FileHeader, contentType string) string {
	extension := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if extension == "" {
		extension = extensionFromContentType(contentType)
	}

	now := time.Now().UTC()
	return fmt.Sprintf("images/%04d/%02d/%02d/%s%s", now.Year(), now.Month(), now.Day(), uuid.NewString(), extension)
}

func extensionFromContentType(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	default:
		return ".bin"
	}
}
