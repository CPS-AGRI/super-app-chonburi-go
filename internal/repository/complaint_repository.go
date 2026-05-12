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

	// Always exclude drafts
	db = db.Where("module_complaints.status != ?", domain.ComplaintStatusDraft)

	// Join with user_informations for searching
	db = db.Joins("LEFT JOIN user_informations ON user_informations.user_id = module_complaints.user_id")

	// Apply Filters
	if len(query.Status) > 0 {
		db = db.Where("module_complaints.status IN ?", query.Status)
	}
	if query.AssigneeId != nil {
		db = db.Where("module_complaints.assignee_id = ?", *query.AssigneeId)
	}
	if query.HasBeenAssigned != nil {
		if *query.HasBeenAssigned {
			db = db.Where("module_complaints.department_id IS NOT NULL")
		} else {
			db = db.Where("module_complaints.department_id IS NULL")
		}
	}

	// Search Filters
	if query.Name != nil && *query.Name != "" {
		db = db.Where("user_informations.name ILIKE ?", "%"+*query.Name+"%")
	}
	if query.LastName != nil && *query.LastName != "" {
		db = db.Where("user_informations.last_name ILIKE ?", "%"+*query.LastName+"%")
	}
	if query.IdentityNumber != nil && *query.IdentityNumber != "" {
		db = db.Where("user_informations.identity_number_hash = ?", *query.IdentityNumber)
	}
	if query.DepartmentID != nil && *query.DepartmentID != "" {
		db = db.Where("module_complaints.department_id = ?", *query.DepartmentID)
	}
	if query.StartDate != nil && *query.StartDate != "" {
		db = db.Where("module_complaints.created_date >= ?", *query.StartDate)
	}
	if query.EndDate != nil && *query.EndDate != "" {
		db = db.Where("module_complaints.created_date <= ?", *query.EndDate+" 23:59:59")
	}

	// Determine Global Workflow Mode from Municipality Settings
	var municipality domain.Municipality
	r.db.First(&municipality)
	isCentralMode := municipality.ComplaintMode == "central"

	if isCentralMode {
		// Mode 2: Central Triage (Option 2)
		if !query.IsSuperAdmin {
			if query.IsComplaintCenter && query.HasBeenAssigned != nil {
				// Center Dashboard View: Show everything unassigned/assigned based on status
				// No extra department filter needed here
			} else {
				// Regular Manage Page (for both Center Admin and Regular Admin):
				// STRICTLY see only complaints already assigned to their departments.
				db = db.Where("module_complaints.department_id IN ?", query.AdminDepartmentIDs)
			}
		}
	} else {
		// Mode 1: Direct Access (Option 1)
		if query.HasBeenAssigned != nil && !*query.HasBeenAssigned {
			// Center Dashboard (Unassigned list) should be empty in Direct Mode
			db = db.Where("1 = 0")
		} else if !query.IsSuperAdmin {
			if len(query.AllowedModuleTypeIDs) > 0 {
				db = db.Where("module_complaints.module_type_id IN ?", query.AllowedModuleTypeIDs)
			} else {
				db = db.Where("1 = 0")
			}
		}
	}

	db.Count(&total)

	offset := (query.PageNumber - 1) * query.PageSize
	err := db.Preload("ModuleType").
		Preload("Department").
		Preload("User.Information").
		Preload("Images").
		Offset(offset).Limit(query.PageSize).
		Order("module_complaints.created_date DESC").
		Find(&complaints).Error

	// Map virtual UserInformation
	if err == nil {
		for i := range complaints {
			if complaints[i].User != nil && complaints[i].User.Information != nil {
				info := *complaints[i].User.Information
				// Strip ENC_ prefix if present
				if len(info.IdentityNumberEncrypted) > 4 && info.IdentityNumberEncrypted[:4] == "ENC_" {
					info.IdentityNumberEncrypted = info.IdentityNumberEncrypted[4:]
				}
				complaints[i].UserInformation = &info
			}
		}
	}

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
	err := r.db.Preload("ModuleType").
		Preload("Department").
		Preload("User.Information").
		Preload("Images").
		Preload("Activities").
		Preload("Activities.Images").
		Where("id = ?", id).
		First(&complaint).Error

	if err == nil && complaint.User != nil && complaint.User.Information != nil {
		info := *complaint.User.Information
		if len(info.IdentityNumberEncrypted) > 4 && info.IdentityNumberEncrypted[:4] == "ENC_" {
			info.IdentityNumberEncrypted = info.IdentityNumberEncrypted[4:]
		}
		complaint.UserInformation = &info
	}

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
