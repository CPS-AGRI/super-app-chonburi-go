package storage

import (
	"bytes"
	"context"
	"io"
	"path/filepath"
	"strings"
	"time"

	"super-app-chonburi-go/pkg/storage/minio"
)

type minioStorage struct {
	client *minio.Client
}

func NewMinIOStorage(client *minio.Client) StorageProvider {
	return &minioStorage{client: client}
}

func (s *minioStorage) Upload(file io.Reader, filename string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	objectKey := "uploads/" + filename

	contentType := "application/octet-stream"
	switch filepath.Ext(filename) {
	case ".pdf":
		contentType = "application/pdf"
	case ".png":
		contentType = "image/png"
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	reader := bytes.NewReader(data)
	size := int64(len(data))

	info, err := s.client.PutObject(ctx, objectKey, reader, size, contentType)
	if err != nil {
		return "", err
	}

	return s.client.ObjectURL(ctx, info.Key)
}

func (s *minioStorage) Delete(filename string) error {
	if filename == "" {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cleanKey := strings.TrimSpace(filename)
	if idx := strings.Index(cleanKey, "/uploads/"); idx != -1 {
		cleanKey = cleanKey[idx+1:]
	} else if !strings.HasPrefix(cleanKey, "uploads/") {
		cleanKey = "uploads/" + filepath.Base(cleanKey)
	}

	return s.client.RemoveObject(ctx, cleanKey)
}

func (s *minioStorage) GetURL(filename string) string {
	ctx := context.Background()
	url, _ := s.client.ObjectURL(ctx, "uploads/"+filename)
	return url
}
