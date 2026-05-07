package repository

import (
	"errors"
	"super-app-chonburi-go/internal/domain"

	"gorm.io/gorm"
)

type complaintRepository struct {
	db *gorm.DB
}

func NewComplaintRepository(db *gorm.DB) domain.ComplaintRepository {
	return &complaintRepository{db: db}
}

func (r *complaintRepository) GetPaginated(query domain.ComplaintQuery) (*domain.PaginatedComplaintResponse, error) {
	var complaints []domain.Complaint
	var total int64

	db := r.db.Model(&domain.Complaint{})

	// Apply Filters
	if len(query.Status) > 0 {
		db = db.Where("status IN ?", query.Status)
	}
	if query.AssigneeId != nil {
		db = db.Where("assignee_id = ?", *query.AssigneeId)
	}
	if !query.IsSuperAdmin {
		if len(query.AllowedModuleTypeIDs) > 0 {
			db = db.Where("module_type_id IN ?", query.AllowedModuleTypeIDs)
		}
	}

	db.Count(&total)

	offset := (query.PageNumber - 1) * query.PageSize
	err := db.Preload("ModuleType").
		Preload("Department").
		Preload("Images").
		Offset(offset).Limit(query.PageSize).
		Order("created_date DESC").
		Find(&complaints).Error

	return &domain.PaginatedComplaintResponse{
		Items:      complaints,
		TotalItems: total,
		PageNumber: query.PageNumber,
		PageSize:   query.PageSize,
		TotalPages: int((total + int64(query.PageSize) - 1) / int64(query.PageSize)),
		HasNext:    total > int64(query.PageNumber*query.PageSize),
		HasPrevious: query.PageNumber > 1,
	}, err
}

func (r *complaintRepository) GetByID(id string, allowedModuleTypeIDs []string, isSuperAdmin bool) (*domain.Complaint, error) {
	var complaint domain.Complaint
	db := r.db.Preload("ModuleType").
		Preload("Department").
		Preload("Images").
		Preload("Activities").
		Preload("Activities.Images")

	err := db.First(&complaint, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &complaint, nil
}

func (r *complaintRepository) Create(complaint *domain.Complaint) error {
	return r.db.Create(complaint).Error
}

func (r *complaintRepository) Update(complaint *domain.Complaint) error {
	return r.db.Save(complaint).Error
}

func (r *complaintRepository) CreateActivity(activity *domain.ComplaintActivity) error {
	return r.db.Create(activity).Error
}

func (r *complaintRepository) Delete(id string) error {
	return r.db.Delete(&domain.Complaint{}, "id = ?", id).Error
}
