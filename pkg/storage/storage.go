package storage

import (
	"io"
)

type StorageProvider interface {
	Upload(file io.Reader, filename string) (string, error)
	Delete(filename string) error
	GetURL(filename string) string
}
