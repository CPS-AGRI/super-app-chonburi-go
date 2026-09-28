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
	"gorm.io/gorm"
)

type publicRelationUseCase struct {
	repo            domain.PublicRelationRepository
	adminRepo       domain.AdminRepository
	storageProvider storage.StorageProvider
	db              *gorm.DB
}

func NewPublicRelationUseCase(
	repo domain.PublicRelationRepository,
	adminRepo domain.AdminRepository,
	storageProvider storage.StorageProvider,
	db *gorm.DB,
) domain.PublicRelationUseCase {
	return &publicRelationUseCase{
		repo:            repo,
		adminRepo:       adminRepo,
		storageProvider: storageProvider,
		db:              db,
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

func (u *publicRelationUseCase) GetAvailableTypes(moduleId string) (*domain.PublicRelationMetadata, error) {
	return u.repo.GetAvailableTypes(moduleId)
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

	err = u.repo.Create(pr)
	if err != nil {
		return err
	}

	// ส่งแจ้งเตือนไปยัง module_notifications หากสถานะเป็น Published หรือ active
	statusLower := strings.ToLower(pr.Status)
	if statusLower == "published" || statusLower == "active" {
		notifTitle := "มีข่าวประชาสัมพันธ์ใหม่"
		notifType := "news"
		if strings.EqualFold(pr.Type, "Activity") {
			notifTitle = "มีข่าวกิจกรรมใหม่"
			notifType = "news_activity"
		}

		body := pr.Title
		if pr.DescriptionTh != nil && *pr.DescriptionTh != "" && *pr.DescriptionTh != pr.Title {
			body = fmt.Sprintf("%s: %s", pr.Title, *pr.DescriptionTh)
		}

		SendBroadcastNotification(
			pr.ModuleId,
			notifTitle,
			body,
			pr.ID.String(),
			pr.Status,
			notifType,
			pr.CreatedBy,
		)
	}

	return nil
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

// CreateNotificationComposite สร้าง notification แบบ atomic รองรับ 3 modes:
// 1. text-only (req.CreateNews == nil && req.PublicRelationId == nil)
// 2. create-news (req.CreateNews != nil) → สร้างข่าวใหม่พร้อมกันใน transaction
// 3. link-news (req.PublicRelationId != nil) → เชื่อมกับข่าวที่มีอยู่แล้ว
func (u *publicRelationUseCase) CreateNotificationComposite(
	moduleId string,
	req domain.CreateNotificationCompositeRequest,
	adminID string,
) (*domain.PublicRelationNotification, error) {
	admin, err := u.adminRepo.GetByID(adminID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	adminName := admin.Name + " " + admin.LastName
	moduleUUID := uuid.MustParse(moduleId)

	var createdNotif *domain.PublicRelationNotification

	txErr := u.db.Transaction(func(tx *gorm.DB) error {
		var linkedPrID *uuid.UUID

		// Mode 2: สร้างข่าวใหม่พร้อมกัน
		if req.CreateNews != nil {
			newsPayload := req.CreateNews

			var startDate, endDate time.Time
			if t, err := time.Parse(time.RFC3339, newsPayload.StartDate); err == nil {
				startDate = t
			} else {
				startDate = now
			}
			if t, err := time.Parse(time.RFC3339, newsPayload.EndDate); err == nil {
				endDate = t
			} else {
				endDate = now.AddDate(0, 1, 0)
			}

			newsPriority := newsPayload.Priority
			if newsPriority == "" {
				newsPriority = "Medium"
			}
			newsType := newsPayload.Type
			if newsType == "" {
				newsType = "Information"
			}
			newsStatus := newsPayload.Status
			if newsStatus == "" {
				newsStatus = "Published"
			}

			newsID := uuid.New()
			pr := domain.PublicRelation{
				ID:            newsID,
				ModuleId:      moduleUUID,
				AdminUserId:   uuid.MustParse(admin.ID),
				Title:         newsPayload.Title,
				DescriptionTh: newsPayload.DescriptionTh,
				DescriptionEn: newsPayload.DescriptionEn,
				Type:          newsType,
				Priority:      newsPriority,
				StartDate:     startDate,
				EndDate:       endDate,
				Status:        newsStatus,
				CreatedDate:   now,
				UpdatedDate:   now,
				CreatedBy:     adminName,
				UpdatedBy:     adminName,
			}

			// แปลงรูป Base64 และ upload ไป Minio
			for i, img := range newsPayload.Images {
				imgUrl := img.Url
				if strings.HasPrefix(imgUrl, "data:") {
					if uploaded, err := u.uploadBase64Image(imgUrl); err == nil {
						imgUrl = uploaded
					}
				}
				pr.Images = append(pr.Images, domain.PublicRelationImage{
					ID:                     uuid.New(),
					ModulePublicRelationId: newsID,
					Url:                    imgUrl,
					Sequence:               i + 1,
					CreatedDate:            now,
					UpdatedDate:            now,
					CreatedBy:              adminName,
					UpdatedBy:              adminName,
				})
			}

			if err := tx.Create(&pr).Error; err != nil {
				return err
			}

			// สร้าง visitor count
			vc := domain.PublicRelationVisitorCount{
				ModulePublicRelationId: newsID,
				Count:                  0,
				CreatedDate:            now,
				UpdatedDate:            now,
				CreatedBy:              adminName,
				UpdatedBy:              adminName,
			}
			if err := tx.Create(&vc).Error; err != nil {
				return err
			}

			linkedPrID = &newsID
		} else if req.PublicRelationId != nil && *req.PublicRelationId != "" {
			// Mode 3: เชื่อมกับข่าวเดิม
			prID, err := uuid.Parse(*req.PublicRelationId)
			if err != nil {
				return errors.New("invalid public_relation_id format")
			}
			linkedPrID = &prID
		}
		// Mode 1: text-only → linkedPrID remains nil

		notifType := req.Type
		if notifType == "" {
			if linkedPrID != nil {
				notifType = "news"
			} else {
				notifType = "text"
			}
		}

		notif := domain.PublicRelationNotification{
			ID:               uuid.New(),
			ModuleId:         moduleUUID,
			AdminUserId:      uuid.MustParse(admin.ID),
			PublicRelationID: linkedPrID,
			Title:            req.Title,
			Description:      req.Description,
			Type:             notifType,
			Status:           req.Status,
			ProcessStatus:    "pending",
			CreatedDate:      now,
			UpdatedDate:      now,
			CreatedBy:        adminName,
			UpdatedBy:        adminName,
		}

		// แปลง send_date ถ้ามี
		if req.SendDate != nil && *req.SendDate != "" {
			if t, err := time.Parse(time.RFC3339, *req.SendDate); err == nil {
				notif.SendDate = &t
			}
		}

		if err := tx.Create(&notif).Error; err != nil {
			return err
		}

		createdNotif = &notif
		return nil
	})

	if txErr != nil {
		return nil, txErr
	}

	// ส่งแจ้งเตือนไปยัง module_notifications หากสถานะเป็น active หรือ published
	if createdNotif != nil {
		statusLower := strings.ToLower(createdNotif.Status)
		if statusLower == "active" || statusLower == "published" {
			refID := createdNotif.ID.String()
			if createdNotif.PublicRelationID != nil {
				refID = createdNotif.PublicRelationID.String()
			}
			body := ""
			if createdNotif.Description != nil {
				body = *createdNotif.Description
			}
			SendBroadcastNotification(
				createdNotif.ModuleId,
				createdNotif.Title,
				body,
				refID,
				createdNotif.Status,
				createdNotif.Type,
				createdNotif.CreatedBy,
			)
		}
	}

	return createdNotif, nil
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
