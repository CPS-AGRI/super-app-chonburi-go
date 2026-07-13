package usecase

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"super-app-chonburi-go/internal/domain"
	"super-app-chonburi-go/pkg/storage"
	"time"

	"github.com/google/uuid"
)

type municipalityUseCase struct {
	repo            domain.MunicipalityRepository
	storageProvider storage.StorageProvider
}

func NewMunicipalityUseCase(
	repo domain.MunicipalityRepository,
	storageProvider storage.StorageProvider,
) domain.MunicipalityUseCase {
	return &municipalityUseCase{
		repo:            repo,
		storageProvider: storageProvider,
	}
}

func (u *municipalityUseCase) uploadBase64Image(base64Str string) (string, error) {
	if !strings.HasPrefix(base64Str, "data:") {
		return base64Str, nil
	}

	parts := strings.SplitN(base64Str, ";base64,", 2)
	if len(parts) != 2 {
		return base64Str, nil
	}

	metaPart := parts[0]
	dataPart := parts[1]

	var extension string
	if strings.Contains(metaPart, "image/png") {
		extension = ".png"
	} else if strings.Contains(metaPart, "image/jpeg") || strings.Contains(metaPart, "image/jpg") {
		extension = ".jpg"
	} else if strings.Contains(metaPart, "image/webp") {
		extension = ".webp"
	} else if strings.Contains(metaPart, "image/gif") {
		extension = ".gif"
	} else {
		extension = ".jpg"
	}

	data, err := base64.StdEncoding.DecodeString(dataPart)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	filename := fmt.Sprintf("muni_%s_%d%s", uuid.New().String(), time.Now().Unix(), extension)
	reader := bytes.NewReader(data)

	url, err := u.storageProvider.Upload(reader, filename)
	if err != nil {
		return "", fmt.Errorf("failed to upload logo to MinIO: %w", err)
	}

	return url, nil
}

func (u *municipalityUseCase) GetList() ([]domain.Municipality, error) {
	return u.repo.GetList()
}

func (u *municipalityUseCase) GetDetail(id uuid.UUID) (*domain.Municipality, error) {
	return u.repo.GetByID(id)
}

func (u *municipalityUseCase) Create(muni *domain.Municipality) error {
	muni.ID = uuid.New()

	if muni.CityLogoUrl != "" {
		minioUrl, err := u.uploadBase64Image(muni.CityLogoUrl)
		if err != nil {
			return err
		}
		muni.CityLogoUrl = minioUrl
	}

	return u.repo.Create(muni)
}

func (u *municipalityUseCase) Update(muni *domain.Municipality) error {

	if muni.CityLogoUrl != "" {
		minioUrl, err := u.uploadBase64Image(muni.CityLogoUrl)
		if err != nil {
			return err
		}
		muni.CityLogoUrl = minioUrl
	}

	return u.repo.Update(muni)
}

func (u *municipalityUseCase) Delete(id uuid.UUID) error {
	return u.repo.Delete(id)
}

func (u *municipalityUseCase) GetCurrent() (*domain.Municipality, error) {
	return u.repo.GetFirst()
}

var (
	ErrMunicipalityNotFound = errors.New("municipality not found")
)
