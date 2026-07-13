package repository

import (
	"super-app-chonburi-go/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type cctvRepository struct {
	db *gorm.DB
}

// NewCCTVRepository creates a new AdminCCTVRepository.
func NewCCTVRepository(db *gorm.DB) domain.AdminCCTVRepository {
	return &cctvRepository{db: db}
}

func (r *cctvRepository) Create(cctv *domain.CCTV) error {
	return r.db.Create(cctv).Error
}

func (r *cctvRepository) GetPaginated(query domain.CCTVQuery) (*domain.PaginatedCCTVResponse, error) {
	var items []domain.CCTV
	var total int64

	db := r.db.Model(&domain.CCTV{})
	if query.Name != "" {
		db = db.Where("name ILIKE ?", "%"+query.Name+"%")
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	offset := (query.PageNumber - 1) * query.PageSize
	err := db.Preload("Creator").Preload("Deleter").Order("created_at DESC").Offset(offset).Limit(query.PageSize).Find(&items).Error
	if err != nil {
		return nil, err
	}

	totalPages := int((total + int64(query.PageSize) - 1) / int64(query.PageSize))
	if totalPages < 1 {
		totalPages = 1
	}

	return &domain.PaginatedCCTVResponse{
		Items:      items,
		TotalItems: total,
		PageNumber: query.PageNumber,
		TotalPages: totalPages,
	}, nil
}

func (r *cctvRepository) GetRequestsPaginated(query domain.CCTVRequestQuery) (*domain.PaginatedCCTVRequestResponse, error) {
	var items []domain.CCTVRequest
	var total int64

	db := r.db.Model(&domain.CCTVRequest{})

	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	offset := (query.PageNumber - 1) * query.PageSize
	// Preload CCTV, User, and User.Information to prevent N+1 query loops.
	err := db.Preload("CCTV").Preload("User").Preload("User.Information").Order("created_at DESC").Offset(offset).Limit(query.PageSize).Find(&items).Error
	if err != nil {
		return nil, err
	}

	totalPages := int((total + int64(query.PageSize) - 1) / int64(query.PageSize))
	if totalPages < 1 {
		totalPages = 1
	}

	return &domain.PaginatedCCTVRequestResponse{
		Items:      items,
		TotalItems: total,
		PageNumber: query.PageNumber,
		TotalPages: totalPages,
	}, nil
}

func (r *cctvRepository) GetRequestByID(id uuid.UUID) (*domain.CCTVRequest, error) {
	var item domain.CCTVRequest
	err := r.db.Preload("CCTV").Preload("User").Preload("User.Information").Where("id = ?", id).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *cctvRepository) UpdateRequest(req *domain.CCTVRequest) error {
	return r.db.Save(req).Error
}

func (r *cctvRepository) Delete(id uuid.UUID, adminID string) error {
	if err := r.db.Model(&domain.CCTV{}).Where("id = ?", id).Update("deleted_by", adminID).Error; err != nil {
		return err
	}
	return r.db.Where("id = ?", id).Delete(&domain.CCTV{}).Error
}
