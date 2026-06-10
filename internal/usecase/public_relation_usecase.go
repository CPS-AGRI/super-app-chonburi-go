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

type publicRelationUseCase struct {
	repo            domain.PublicRelationRepository
	adminRepo       domain.AdminRepository
	storageProvider storage.StorageProvider
}

func NewPublicRelationUseCase(
	repo domain.PublicRelationRepository,
	adminRepo domain.AdminRepository,
	storageProvider storage.StorageProvider,
) domain.PublicRelationUseCase {
	return &publicRelationUseCase{
		repo:            repo,
		adminRepo:       adminRepo,
		storageProvider: storageProvider,
	}
}

func (u *publicRelationUseCase) uploadBase64Image(base64Str string) (string, error) {
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
		return "", err
	}

	filename := fmt.Sprintf("%s%s", uuid.New().String(), extension)
	reader := bytes.NewReader(data)

	url, err := u.storageProvider.Upload(reader, filename)
	if err != nil {
		return "", err
	}

	return url, nil
}

func (u *publicRelationUseCase) GetDashboardStats(moduleId string) (*domain.PublicRelationDashboardStats, error) {
	return u.repo.GetDashboardStats(moduleId)
}

func (u *publicRelationUseCase) GetPopularNews(moduleId string, limit int) ([]domain.PublicRelation, error) {
	if limit <= 0 {
		limit = 5
	}
	return u.repo.GetPopularNews(moduleId, limit)
}

func (u *publicRelationUseCase) GetExpiringNews(moduleId string, limit int) ([]domain.PublicRelation, error) {
	if limit <= 0 {
		limit = 5
	}
	return u.repo.GetExpiringNews(moduleId, limit)
}

func (u *publicRelationUseCase) GetPaginated(moduleId string, query domain.PublicRelationQuery) (*domain.PaginatedPublicRelationResponse, error) {
	if query.PageNumber < 1 {
		query.PageNumber = 1
	}
	if query.PageSize < 1 {
		query.PageSize = 10
	}
	return u.repo.GetPaginated(moduleId, query)
}

func (u *publicRelationUseCase) GetByID(moduleId string, id string) (*domain.PublicRelation, error) {
	return u.repo.GetByID(moduleId, id)
}

func (u *publicRelationUseCase) Create(pr *domain.PublicRelation, adminID string) error {
	admin, err := u.adminRepo.GetByID(adminID)
	if err != nil {
		return err
	}

	pr.ID = uuid.New()
	pr.AdminUserId = uuid.MustParse(admin.ID)
	pr.CreatedDate = time.Now()
	pr.UpdatedDate = time.Now()
	pr.CreatedBy = admin.Name + " " + admin.LastName
	pr.UpdatedBy = admin.Name + " " + admin.LastName

	for i := range pr.Images {
		if strings.HasPrefix(pr.Images[i].Url, "data:") {
			minioUrl, err := u.uploadBase64Image(pr.Images[i].Url)
			if err == nil {
				pr.Images[i].Url = minioUrl
			}
		}
		pr.Images[i].ID = uuid.New()
		pr.Images[i].ModulePublicRelationId = pr.ID
		pr.Images[i].Sequence = i + 1
		pr.Images[i].CreatedDate = pr.CreatedDate
		pr.Images[i].UpdatedDate = pr.UpdatedDate
		pr.Images[i].CreatedBy = pr.CreatedBy
		pr.Images[i].UpdatedBy = pr.UpdatedBy
	}

	return u.repo.Create(pr)
}

func (u *publicRelationUseCase) Update(pr *domain.PublicRelation, adminID string) error {
	admin, err := u.adminRepo.GetByID(adminID)
	if err != nil {
		return err
	}

	existing, err := u.repo.GetByID(pr.ModuleId.String(), pr.ID.String())
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("public relation not found")
	}

	existing.Title = pr.Title
	existing.DescriptionTh = pr.DescriptionTh
	existing.DescriptionEn = pr.DescriptionEn
	existing.Type = pr.Type
	existing.Priority = pr.Priority
	existing.StartDate = pr.StartDate
	existing.EndDate = pr.EndDate
	existing.Status = pr.Status
	existing.UpdatedDate = time.Now()
	existing.UpdatedBy = admin.Name + " " + admin.LastName

	existing.Images = pr.Images
	for i := range existing.Images {
		if strings.HasPrefix(existing.Images[i].Url, "data:") {
			minioUrl, err := u.uploadBase64Image(existing.Images[i].Url)
			if err == nil {
				existing.Images[i].Url = minioUrl
			}
		}
		existing.Images[i].ID = uuid.New()
		existing.Images[i].ModulePublicRelationId = existing.ID
		existing.Images[i].Sequence = i + 1
		existing.Images[i].CreatedDate = existing.CreatedDate
		existing.Images[i].UpdatedDate = existing.UpdatedDate
		existing.Images[i].CreatedBy = existing.CreatedBy
		existing.Images[i].UpdatedBy = existing.UpdatedBy
	}

	return u.repo.Update(existing)
}

func (u *publicRelationUseCase) Delete(moduleId string, id string, adminID string) error {
	return u.repo.Delete(moduleId, id)
}

func (u *publicRelationUseCase) HideComment(moduleId string, prId string, commentId string, adminID string) error {
	return u.repo.HideComment(moduleId, prId, commentId)
}

func (u *publicRelationUseCase) ShowComment(moduleId string, prId string, commentId string, adminID string) error {
	return u.repo.ShowComment(moduleId, prId, commentId)
}

func (u *publicRelationUseCase) GetPaginatedNotifications(moduleId string, query domain.PublicRelationNotificationQuery, history bool) (*domain.PaginatedNotificationResponse, error) {
	if query.PageNumber < 1 {
		query.PageNumber = 1
	}
	if query.PageSize < 1 {
		query.PageSize = 10
	}
	return u.repo.GetPaginatedNotifications(moduleId, query, history)
}

func (u *publicRelationUseCase) GetNotificationByID(moduleId string, id string) (*domain.PublicRelationNotification, error) {
	return u.repo.GetNotificationByID(moduleId, id)
}

func (u *publicRelationUseCase) CreateNotification(notification *domain.PublicRelationNotification, adminID string) error {
	admin, err := u.adminRepo.GetByID(adminID)
	if err != nil {
		return err
	}

	notification.ID = uuid.New()
	notification.AdminUserId = uuid.MustParse(admin.ID)
	notification.ProcessStatus = "pending"
	notification.CreatedDate = time.Now()
	notification.UpdatedDate = time.Now()
	notification.CreatedBy = admin.Name + " " + admin.LastName
	notification.UpdatedBy = admin.Name + " " + admin.LastName

	return u.repo.CreateNotification(notification)
}

func (u *publicRelationUseCase) UpdateNotification(notification *domain.PublicRelationNotification, adminID string) error {
	admin, err := u.adminRepo.GetByID(adminID)
	if err != nil {
		return err
	}

	existing, err := u.repo.GetNotificationByID(notification.ModuleId.String(), notification.ID.String())
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("notification not found")
	}

	existing.Title = notification.Title
	existing.Description = notification.Description
	existing.SendDate = notification.SendDate
	existing.Type = notification.Type
	existing.Status = notification.Status
	existing.UpdatedDate = time.Now()
	existing.UpdatedBy = admin.Name + " " + admin.LastName

	return u.repo.UpdateNotification(existing)
}

func (u *publicRelationUseCase) DeleteNotification(moduleId string, id string, adminID string) error {
	return u.repo.DeleteNotification(moduleId, id)
}

func (u *publicRelationUseCase) GetWelcomeScreens() ([]domain.MunicipalityWelcomeScreen, error) {
	return u.repo.GetWelcomeScreens()
}

func (u *publicRelationUseCase) UploadWelcomeScreen(screen *domain.MunicipalityWelcomeScreen, adminID string) error {
	admin, err := u.adminRepo.GetByID(adminID)
	if err != nil {
		return err
	}

	if strings.HasPrefix(screen.ImageUrl, "data:") {
		minioUrl, err := u.uploadBase64Image(screen.ImageUrl)
		if err == nil {
			screen.ImageUrl = minioUrl
		}
	}

	if screen.IsActive {
		existing, err := u.repo.GetWelcomeScreens()
		if err == nil {
			for _, ext := range existing {
				if ext.IsActive && ext.ID != screen.ID {
					ext.IsActive = false
					ext.UpdatedDate = time.Now()
					ext.UpdatedBy = admin.Name + " " + admin.LastName
					_ = u.repo.UpdateWelcomeScreen(&ext)
				}
			}
		}
	}

	if screen.ID == uuid.Nil {
		screen.ID = uuid.New()
		screen.CreatedDate = time.Now()
		screen.UpdatedDate = time.Now()
		screen.CreatedBy = admin.Name + " " + admin.LastName
		screen.UpdatedBy = admin.Name + " " + admin.LastName
		return u.repo.CreateWelcomeScreen(screen)
	} else {
		screen.UpdatedDate = time.Now()
		screen.UpdatedBy = admin.Name + " " + admin.LastName
		return u.repo.UpdateWelcomeScreen(screen)
	}
}
