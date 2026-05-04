package repository

import (
	"errors"
	"math"
	"super-app-chonburi-go/internal/domain"

	"github.com/google/uuid"
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
	var totalItems int64

	dbQuery := r.db.Model(&domain.Complaint{})

	// Apply filtering for PermissionIDs if not SuperAdmin
	if !query.IsSuperAdmin {
		if query.AssigneeID != nil {
			dbQuery = dbQuery.Where("assignee_id = ?", query.AssigneeID)
		} else {
			if len(query.AllowedPermissionIDs) == 0 {
				// User has no permissions, return empty immediately
				return &domain.PaginatedComplaintResponse{
					Items:       []domain.Complaint{},
					TotalItems:  0,
					PageNumber:  query.PageNumber,
					PageSize:    query.PageSize,
					TotalPages:  0,
					HasNext:     false,
					HasPrevious: false,
				}, nil
			}
			dbQuery = dbQuery.Where("permission_id IN ?", query.AllowedPermissionIDs)
		}
	}

	// Apply status filter
	if len(query.Status) > 0 {
		dbQuery = dbQuery.Where("status IN ?", query.Status)
	}

	// Apply assignee filter if it wasn't already handled above
	if query.AssigneeID != nil && query.IsSuperAdmin {
		dbQuery = dbQuery.Where("assignee_id = ?", query.AssigneeID)
	}

	// Count total
	if err := dbQuery.Count(&totalItems).Error; err != nil {
		return nil, err
	}

	// Pagination
	offset := (query.PageNumber - 1) * query.PageSize
	
	// Fetch items with relations
	err := dbQuery.
		Preload("Permission").
		Preload("Assigner").
		Preload("Assignee").
		Preload("UserInformation").
		Order("created_at desc").
		Offset(offset).
		Limit(query.PageSize).
		Find(&complaints).Error

	if err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(totalItems) / float64(query.PageSize)))

	return &domain.PaginatedComplaintResponse{
		Items:       complaints,
		TotalItems:  totalItems,
		PageNumber:  query.PageNumber,
		PageSize:    query.PageSize,
		TotalPages:  totalPages,
		HasNext:     query.PageNumber < totalPages,
		HasPrevious: query.PageNumber > 1,
	}, nil
}

func (r *complaintRepository) GetByID(id uuid.UUID, allowedPermissionIDs []string, isSuperAdmin bool) (*domain.Complaint, error) {
	var complaint domain.Complaint
	
	dbQuery := r.db.
		Preload("Permission").
		Preload("Assigner").
		Preload("Assignee").
		Preload("RejectedBy").
		Preload("UserInformation").
		Preload("Images").
		Preload("Activities").
		Preload("Activities.Admin").
		Preload("Activities.Images").
		Where("id = ?", id)

	if !isSuperAdmin {
		if len(allowedPermissionIDs) == 0 {
			return nil, gorm.ErrRecordNotFound
		}
		dbQuery = dbQuery.Where("permission_id IN ?", allowedPermissionIDs)
	}

	err := dbQuery.First(&complaint).Error
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

func (r *complaintRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&domain.Complaint{}, "id = ?", id).Error
}
