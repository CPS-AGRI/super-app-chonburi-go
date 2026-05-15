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

	// 1. Status Filter
	if len(query.Status) > 0 {
		db = db.Where("module_complaints.status IN ?", query.Status)
	} else {
		db = db.Where("module_complaints.status != ?", domain.ComplaintStatusDraft)
	}

	// 2. Search Filters
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
	if query.AssigneeId != nil {
		db = db.Where("module_complaints.assignee_id = ?", *query.AssigneeId)
	}

	// 3. Visibility Logic
	if !query.IsSuperAdmin {
		// Determine Global Workflow Mode from Municipality Settings
		var municipality domain.Municipality
		r.db.First(&municipality)
		isCentralMode := municipality.ComplaintMode == "central"

		isRequestingRejected := false
		for _, s := range query.Status {
			if s == domain.ComplaintStatusRejected {
				isRequestingRejected = true
				break
			}
		}

		// Determine if we are in a Dashboard-specific view
		isDashboardView := query.HasBeenAssigned != nil || isRequestingRejected

		if query.IsComplaintCenter && isDashboardView {
			// --- 1. Dashboard View (Center Admin) ---
			if !isCentralMode {
				// In Option 1, the Center Dashboard should be empty to reduce confusion
				db = db.Where("1 = 0")
			} else {
				// Option 2 Center Dashboard: Bypass department filter
				if query.HasBeenAssigned != nil && !isRequestingRejected {
					if *query.HasBeenAssigned {
						db = db.Where("module_complaints.department_id IS NOT NULL")
					} else {
						db = db.Where("module_complaints.department_id IS NULL")
					}
				}
			}
		} else {
			// --- 2. Manage Page or Regular Admin View ---
			if isCentralMode {
				// Option 2 Hybrid Logic:
				// 1. Show complaints explicitly assigned to these departments
				// 2. Show legacy complaints (department_id IS NULL) that were already handled (status != pending)
				//    and belong to categories these departments manage (based on their allowed module types).
				if len(query.AdminDepartmentIDs) > 0 {
					db = db.Where(
						"(module_complaints.department_id IN ?) OR (module_complaints.department_id IS NULL AND module_complaints.status != ? AND module_complaints.module_type_id IN ?)",
						query.AdminDepartmentIDs, domain.ComplaintStatusPending, query.AllowedModuleTypeIDs,
					)
				} else {
					// Non-center admins with no departments see nothing
					db = db.Where("1 = 0")
				}
			} else {
				// Option 1: Filter by allowed module types
				if len(query.AllowedModuleTypeIDs) > 0 {
					db = db.Where("module_complaints.module_type_id IN ?", query.AllowedModuleTypeIDs)
				} else {
					db = db.Where("1 = 0")
				}
			}

			// Hide rejected from manage page in all modes as requested (Point 67 in MD)
			if !isRequestingRejected {
				db = db.Where("module_complaints.status != ?", domain.ComplaintStatusRejected)
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

	// Map virtual UserInformation and Assignee
	if err == nil {
		assigneeIDs := []string{}
		for i := range complaints {
			if complaints[i].AssigneeId != nil {
				assigneeIDs = append(assigneeIDs, *complaints[i].AssigneeId)
			}
			
			if complaints[i].User != nil && complaints[i].User.Information != nil {
				info := *complaints[i].User.Information
				// Strip ENC_ prefix if present
				if len(info.IdentityNumberEncrypted) > 4 && info.IdentityNumberEncrypted[:4] == "ENC_" {
					info.IdentityNumberEncrypted = info.IdentityNumberEncrypted[4:]
				}
				complaints[i].UserInformation = &info
			}
		}

		if len(assigneeIDs) > 0 {
			var assignees []domain.Admin
			if r.db.Where("id IN ?", assigneeIDs).Find(&assignees).Error == nil {
				assigneeMap := make(map[string]*domain.Admin)
				for i := range assignees {
					assigneeMap[assignees[i].ID] = &assignees[i]
				}
				for i := range complaints {
					if complaints[i].AssigneeId != nil {
						complaints[i].Assignee = assigneeMap[*complaints[i].AssigneeId]
					}
				}
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

	if err == nil {
		// Populate Assignee
		if complaint.AssigneeId != nil {
			var assignee domain.Admin
			if r.db.Where("id = ?", *complaint.AssigneeId).First(&assignee).Error == nil {
				complaint.Assignee = &assignee
			}
		}

		// Populate Activities Admin
		if len(complaint.Activities) > 0 {
			adminIDs := []string{}
			for _, act := range complaint.Activities {
				if act.CreatedBy != "" {
					adminIDs = append(adminIDs, act.CreatedBy)
				}
			}

			if len(adminIDs) > 0 {
				var admins []domain.Admin
				if r.db.Where("id IN ?", adminIDs).Find(&admins).Error == nil {
					adminMap := make(map[string]*domain.Admin)
					for i := range admins {
						adminMap[admins[i].ID] = &admins[i]
					}
					for i := range complaint.Activities {
						complaint.Activities[i].Admin = adminMap[complaint.Activities[i].CreatedBy]
					}
				}
			}
		}
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

func (r *complaintRepository) GetAllowedModuleTypeIDs(deptIDs []string) ([]string, error) {
	var typeIDs []string
	err := r.db.Table("department_module_module_types").
		Joins("JOIN department_modules ON department_modules.id = department_module_module_types.department_module_id").
		Where("department_modules.department_id IN ?", deptIDs).
		Pluck("department_module_module_types.module_type_id", &typeIDs).Error

	return typeIDs, err
}
