package repository

import (
	"errors"
	"super-app-chonburi-go/internal/domain"
	"time"

	"gorm.io/gorm"
)

type publicRelationRepository struct {
	db *gorm.DB
}

func NewPublicRelationRepository(db *gorm.DB) domain.PublicRelationRepository {
	return &publicRelationRepository{db: db}
}

// Dashboard
func (r *publicRelationRepository) GetDashboardStats(moduleId string) (*domain.PublicRelationDashboardStats, error) {
	var stats domain.PublicRelationDashboardStats
	now := time.Now()

	// 1. Active News Count
	r.db.Model(&domain.PublicRelation{}).
		Where("module_id = ? AND status = ? AND start_date <= ? AND end_date >= ?", moduleId, "Published", now, now).
		Count(&stats.ActiveNewsCount)

	// 2. Average Viewers
	type AvgResult struct {
		Avg float64
	}
	var avgResult AvgResult
	r.db.Raw(`
		SELECT COALESCE(AVG(count), 0) as avg 
		FROM module_public_relation_visitor_count vc
		JOIN module_public_relations pr ON pr.id = vc.module_public_relation_id
		WHERE pr.module_id = ?`, moduleId).Scan(&avgResult)
	stats.AverageViewers = int64(avgResult.Avg)

	// 3. Total News
	r.db.Model(&domain.PublicRelation{}).Where("module_id = ?", moduleId).Count(&stats.TotalNewsCount)

	// 4. Total Notifications
	r.db.Model(&domain.PublicRelationNotification{}).Where("module_id = ?", moduleId).Count(&stats.TotalNotificationCount)

	// 5. Total Likes
	r.db.Raw(`
		SELECT COUNT(*) 
		FROM module_public_relation_likes l
		JOIN module_public_relations pr ON pr.id = l.module_public_relation_id
		WHERE pr.module_id = ?`, moduleId).Scan(&stats.TotalLikesCount)

	// 6. Total Comments
	r.db.Raw(`
		SELECT COUNT(*) 
		FROM module_public_relation_comments c
		JOIN module_public_relations pr ON pr.id = c.module_public_relation_id
		WHERE pr.module_id = ?`, moduleId).Scan(&stats.TotalCommentsCount)

	// 7. Reported Comments (Status = 'reported')
	r.db.Raw(`
		SELECT COUNT(*) 
		FROM module_public_relation_comments c
		JOIN module_public_relations pr ON pr.id = c.module_public_relation_id
		WHERE pr.module_id = ? AND c.status = 'reported'`, moduleId).Scan(&stats.ReportedCommentsCount)

	return &stats, nil
}

func (r *publicRelationRepository) GetPopularNews(moduleId string, limit int) ([]domain.PublicRelation, error) {
	var prs []domain.PublicRelation
	err := r.db.Preload("VisitorCount").
		Joins("LEFT JOIN module_public_relation_visitor_count vc ON vc.module_public_relation_id = module_public_relations.id").
		Where("module_public_relations.module_id = ?", moduleId).
		Order("COALESCE(vc.count, 0) DESC").
		Limit(limit).
		Find(&prs).Error
	return prs, err
}

func (r *publicRelationRepository) GetExpiringNews(moduleId string, limit int) ([]domain.PublicRelation, error) {
	var prs []domain.PublicRelation
	now := time.Now()
	err := r.db.Where("module_id = ? AND status = ? AND end_date >= ?", moduleId, "Published", now).
		Order("end_date ASC").
		Limit(limit).
		Find(&prs).Error
	return prs, err
}

// News
func (r *publicRelationRepository) GetPaginated(moduleId string, query domain.PublicRelationQuery) (*domain.PaginatedPublicRelationResponse, error) {
	var items []domain.PublicRelation
	var total int64

	db := r.db.Model(&domain.PublicRelation{}).Where("module_id = ?", moduleId)

	if query.Title != nil && *query.Title != "" {
		db = db.Where("title ILIKE ?", "%"+*query.Title+"%")
	}
	if query.StartDate != nil && *query.StartDate != "" {
		db = db.Where("start_date >= ?", *query.StartDate)
	}
	if query.EndDate != nil && *query.EndDate != "" {
		db = db.Where("end_date <= ?", *query.EndDate+" 23:59:59")
	}

	db.Count(&total)

	offset := (query.PageNumber - 1) * query.PageSize
	err := db.Preload("Images").
		Preload("VisitorCount").
		Preload("AdminUser").
		Preload("Comments.User.Information").
		Offset(offset).Limit(query.PageSize).
		Order("created_date DESC").
		Find(&items).Error

	return &domain.PaginatedPublicRelationResponse{
		Items:       items,
		TotalItems:  total,
		PageNumber:  query.PageNumber,
		PageSize:    query.PageSize,
		TotalPages:  int((total + int64(query.PageSize) - 1) / int64(query.PageSize)),
		HasNext:     total > int64(query.PageNumber*query.PageSize),
		HasPrevious: query.PageNumber > 1,
	}, err
}

func (r *publicRelationRepository) GetByID(moduleId string, id string) (*domain.PublicRelation, error) {
	var pr domain.PublicRelation
	err := r.db.Preload("Images").
		Preload("VisitorCount").
		Preload("Comments", func(db *gorm.DB) *gorm.DB {
			return db.Preload("User.Information").Order("created_date DESC")
		}).
		Preload("AdminUser").
		Where("module_id = ? AND id = ?", moduleId, id).
		First(&pr).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &pr, nil
}

func (r *publicRelationRepository) Create(pr *domain.PublicRelation) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(pr).Error; err != nil {
			return err
		}
		// Initialize Visitor Count
		vc := domain.PublicRelationVisitorCount{
			ModulePublicRelationId: pr.ID,
			Count:                  0,
			CreatedDate:            pr.CreatedDate,
			UpdatedDate:            pr.UpdatedDate,
			CreatedBy:              pr.CreatedBy,
			UpdatedBy:              pr.UpdatedBy,
		}
		return tx.Create(&vc).Error
	})
}

func (r *publicRelationRepository) Update(pr *domain.PublicRelation) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Clean up existing images first
		if err := tx.Where("module_public_relation_id = ?", pr.ID).Delete(&domain.PublicRelationImage{}).Error; err != nil {
			return err
		}

		// Insert new images
		for i := range pr.Images {
			pr.Images[i].ModulePublicRelationId = pr.ID
			if err := tx.Create(&pr.Images[i]).Error; err != nil {
				return err
			}
		}

		return tx.Session(&gorm.Session{FullSaveAssociations: false}).Save(pr).Error
	})
}

func (r *publicRelationRepository) Delete(moduleId string, id string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Delete related records
		tx.Where("module_public_relation_id = ?", id).Delete(&domain.PublicRelationImage{})
		tx.Where("module_public_relation_id = ?", id).Delete(&domain.PublicRelationComment{})
		tx.Where("module_public_relation_id = ?", id).Delete(&domain.PublicRelationLike{})
		tx.Where("module_public_relation_id = ?", id).Delete(&domain.PublicRelationVisitorCount{})
		return tx.Where("module_id = ? AND id = ?", moduleId, id).Delete(&domain.PublicRelation{}).Error
	})
}

func (r *publicRelationRepository) HideComment(moduleId string, prId string, commentId string) error {
	// First verify news module ownership
	var count int64
	r.db.Model(&domain.PublicRelation{}).Where("module_id = ? AND id = ?", moduleId, prId).Count(&count)
	if count == 0 {
		return errors.New("public relation not found in this module")
	}

	return r.db.Model(&domain.PublicRelationComment{}).
		Where("id = ? AND module_public_relation_id = ?", commentId, prId).
		Update("status", "hidden").Error
}

func (r *publicRelationRepository) ShowComment(moduleId string, prId string, commentId string) error {
	// First verify news module ownership
	var count int64
	r.db.Model(&domain.PublicRelation{}).Where("module_id = ? AND id = ?", moduleId, prId).Count(&count)
	if count == 0 {
		return errors.New("public relation not found in this module")
	}

	return r.db.Model(&domain.PublicRelationComment{}).
		Where("id = ? AND module_public_relation_id = ?", commentId, prId).
		Update("status", "active").Error
}

// Notifications
func (r *publicRelationRepository) GetPaginatedNotifications(moduleId string, query domain.PublicRelationNotificationQuery, history bool) (*domain.PaginatedNotificationResponse, error) {
	var items []domain.PublicRelationNotification
	var total int64

	db := r.db.Model(&domain.PublicRelationNotification{}).Where("module_id = ?", moduleId)

	if history {
		db = db.Where("process_status = ?", "success")
	} else {
		db = db.Where("process_status != ?", "success")
	}

	if query.Title != nil && *query.Title != "" {
		db = db.Where("title ILIKE ?", "%"+*query.Title+"%")
	}
	if query.StartDate != nil && *query.StartDate != "" {
		db = db.Where("created_date >= ?", *query.StartDate)
	}
	if query.EndDate != nil && *query.EndDate != "" {
		db = db.Where("created_date <= ?", *query.EndDate+" 23:59:59")
	}

	db.Count(&total)

	offset := (query.PageNumber - 1) * query.PageSize
	err := db.Preload("AdminUser").
		Offset(offset).Limit(query.PageSize).
		Order("created_date DESC").
		Find(&items).Error

	return &domain.PaginatedNotificationResponse{
		Items:       items,
		TotalItems:  total,
		PageNumber:  query.PageNumber,
		PageSize:    query.PageSize,
		TotalPages:  int((total + int64(query.PageSize) - 1) / int64(query.PageSize)),
		HasNext:     total > int64(query.PageNumber*query.PageSize),
		HasPrevious: query.PageNumber > 1,
	}, err
}

func (r *publicRelationRepository) GetNotificationByID(moduleId string, id string) (*domain.PublicRelationNotification, error) {
	var n domain.PublicRelationNotification
	err := r.db.Preload("AdminUser").Where("module_id = ? AND id = ?", moduleId, id).First(&n).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &n, nil
}

func (r *publicRelationRepository) CreateNotification(notification *domain.PublicRelationNotification) error {
	return r.db.Create(notification).Error
}

func (r *publicRelationRepository) UpdateNotification(notification *domain.PublicRelationNotification) error {
	return r.db.Save(notification).Error
}

func (r *publicRelationRepository) DeleteNotification(moduleId string, id string) error {
	return r.db.Where("module_id = ? AND id = ?", moduleId, id).Delete(&domain.PublicRelationNotification{}).Error
}

// Welcome Screen
func (r *publicRelationRepository) GetWelcomeScreens() ([]domain.MunicipalityWelcomeScreen, error) {
	var screens []domain.MunicipalityWelcomeScreen
	err := r.db.Order("created_date DESC").Find(&screens).Error
	return screens, err
}

func (r *publicRelationRepository) CreateWelcomeScreen(screen *domain.MunicipalityWelcomeScreen) error {
	return r.db.Create(screen).Error
}

func (r *publicRelationRepository) UpdateWelcomeScreen(screen *domain.MunicipalityWelcomeScreen) error {
	return r.db.Save(screen).Error
}

func (r *publicRelationRepository) DeleteWelcomeScreen(id string) error {
	return r.db.Delete(&domain.MunicipalityWelcomeScreen{}, "id = ?", id).Error
}
