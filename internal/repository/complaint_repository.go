package repository

import (
	"errors"
	"sort"
	"super-app-chonburi-go/internal/domain"
	"time"

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

	db = db.Where("module_complaints.status != ?", domain.ComplaintStatusDraft)

	db = db.Joins("LEFT JOIN user_informations ON user_informations.user_id = module_complaints.user_id")

	if len(query.Status) > 0 {
		db = db.Where("module_complaints.status IN ?", query.Status)
	} else {
		db = db.Where("module_complaints.status != ?", domain.ComplaintStatusDraft)
	}

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

	if !query.IsSuperAdmin {

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

		isDashboardView := query.HasBeenAssigned != nil || isRequestingRejected || query.IsComplaintCenter

		if query.AssigneeId != nil {

			if query.AdminRoleType == "Employees" {
				db = db.Where("module_complaints.status NOT IN ?", []string{domain.ComplaintStatusPending, domain.ComplaintStatusReceived})
			}
		} else if query.IsComplaintCenter && isDashboardView {

			if isCentralMode {

				if query.HasBeenAssigned != nil && !isRequestingRejected {
					if *query.HasBeenAssigned {
						db = db.Where("module_complaints.department_id IS NOT NULL")
					} else {
						db = db.Where("module_complaints.department_id IS NULL")
					}
				}
			} else {

				isPendingOnly := len(query.Status) == 1 && query.Status[0] == domain.ComplaintStatusPending
				if isPendingOnly {
					db = db.Where("1 = 0")
				}
			}
		} else {

			if isCentralMode {

				if len(query.AdminDepartmentIDs) > 0 {
					db = db.Where(
						"(module_complaints.department_id IN ?) OR (module_complaints.department_id IS NULL AND module_complaints.status != ? AND module_complaints.module_type_id IN ?)",
						query.AdminDepartmentIDs, domain.ComplaintStatusPending, query.AllowedModuleTypeIDs,
					)
				} else {

					db = db.Where("1 = 0")
				}
			} else {

				if len(query.AllowedModuleTypeIDs) > 0 {
					db = db.Where("module_complaints.module_type_id IN ?", query.AllowedModuleTypeIDs)
				} else {
					db = db.Where("1 = 0")
				}
			}

			if !isRequestingRejected {
				db = db.Where("module_complaints.status != ?", domain.ComplaintStatusRejected)
			}

			if query.AdminRoleType == "Employees" {
				db = db.Where("module_complaints.status NOT IN ?", []string{domain.ComplaintStatusPending, domain.ComplaintStatusReceived})
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

	if err == nil {
		assigneeIDs := []string{}
		for i := range complaints {
			if complaints[i].AssigneeId != nil {
				assigneeIDs = append(assigneeIDs, *complaints[i].AssigneeId)
			}

			if complaints[i].User != nil && complaints[i].User.Information != nil {
				info := *complaints[i].User.Information

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
		Items:       complaints,
		TotalItems:  total,
		PageNumber:  query.PageNumber,
		PageSize:    query.PageSize,
		TotalPages:  int((total + int64(query.PageSize) - 1) / int64(query.PageSize)),
		HasNext:     total > int64(query.PageNumber*query.PageSize),
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

		if complaint.AssigneeId != nil {
			var assignee domain.Admin
			if r.db.Where("id = ?", *complaint.AssigneeId).First(&assignee).Error == nil {
				complaint.Assignee = &assignee
			}
		}

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

func (r *complaintRepository) GetRatingSummaries(summaryType string) ([]domain.ComplaintRatingSummary, error) {
	var summaries []domain.ComplaintRatingSummary

	if summaryType == domain.RatingSummaryTypeAssignee {
		type RatingRow struct {
			ReferenceId   string
			ReferenceName string
			TotalRatings  int
			TotalScore    int
			DisputeCount  int
		}
		var rows []RatingRow
		err := r.db.Raw(`
			SELECT
				a.id AS reference_id,
				CONCAT(a.name, ' ', a.last_name) AS reference_name,
				COUNT(h.rating_score) AS total_ratings,
				COALESCE(SUM(h.rating_score), 0) AS total_score,
				COUNT(CASE WHEN h.is_disputed THEN 1 END) AS dispute_count
			FROM admin_users a
			LEFT JOIN module_complaint_rating_histories h ON h.assignee_id::text = a.id::text
			GROUP BY a.id, a.name, a.last_name
			HAVING COUNT(h.rating_score) > 0 OR COUNT(CASE WHEN h.is_disputed THEN 1 END) > 0
		`).Scan(&rows).Error
		if err != nil {
			return nil, err
		}

		for _, row := range rows {
			avg := 0.0
			if row.TotalRatings > 0 {
				avg = float64(row.TotalScore) / float64(row.TotalRatings)
			}
			summaries = append(summaries, domain.ComplaintRatingSummary{
				ID:            row.ReferenceId,
				SummaryType:   domain.RatingSummaryTypeAssignee,
				ReferenceId:   row.ReferenceId,
				ReferenceName: row.ReferenceName,
				TotalRatings:  row.TotalRatings,
				TotalScore:    row.TotalScore,
				AverageScore:  avg,
				DisputeCount:  row.DisputeCount,
				LastUpdatedAt: time.Now(),
			})
		}
	} else if summaryType == domain.RatingSummaryTypeDepartment {
		type RatingRow struct {
			ReferenceId   string
			ReferenceName string
			TotalRatings  int
			TotalScore    int
			DisputeCount  int
		}
		var rows []RatingRow
		err := r.db.Raw(`
			SELECT
				d.id AS reference_id,
				d.name AS reference_name,
				COUNT(h.rating_score) AS total_ratings,
				COALESCE(SUM(h.rating_score), 0) AS total_score,
				COUNT(CASE WHEN h.is_disputed THEN 1 END) AS dispute_count
			FROM departments d
			LEFT JOIN module_complaint_rating_histories h ON h.department_id::text = d.id::text
			GROUP BY d.id, d.name
			HAVING COUNT(h.rating_score) > 0 OR COUNT(CASE WHEN h.is_disputed THEN 1 END) > 0
		`).Scan(&rows).Error
		if err != nil {
			return nil, err
		}

		for _, row := range rows {
			avg := 0.0
			if row.TotalRatings > 0 {
				avg = float64(row.TotalScore) / float64(row.TotalRatings)
			}
			summaries = append(summaries, domain.ComplaintRatingSummary{
				ID:            row.ReferenceId,
				SummaryType:   domain.RatingSummaryTypeDepartment,
				ReferenceId:   row.ReferenceId,
				ReferenceName: row.ReferenceName,
				TotalRatings:  row.TotalRatings,
				TotalScore:    row.TotalScore,
				AverageScore:  avg,
				DisputeCount:  row.DisputeCount,
				LastUpdatedAt: time.Now(),
			})
		}
	}

	sort.Slice(summaries, func(i, j int) bool {
		if summaries[i].AverageScore == summaries[j].AverageScore {
			return summaries[i].TotalRatings > summaries[j].TotalRatings
		}
		return summaries[i].AverageScore > summaries[j].AverageScore
	})

	return summaries, nil
}

func (r *complaintRepository) GetOverviewStats() (*domain.ComplaintOverviewStats, error) {
	stats := &domain.ComplaintOverviewStats{}
	r.db.Model(&domain.Complaint{}).Where("status != ?", domain.ComplaintStatusDraft).Count(&stats.Total)
	r.db.Model(&domain.Complaint{}).Where("status = ?", domain.ComplaintStatusPending).Count(&stats.Pending)
	r.db.Model(&domain.Complaint{}).Where("status = ?", domain.ComplaintStatusReceived).Count(&stats.Received)
	r.db.Model(&domain.Complaint{}).Where("status = ?", domain.ComplaintStatusInProgress).Count(&stats.InProgress)
	r.db.Model(&domain.Complaint{}).Where("status = ?", domain.ComplaintStatusCompleted).Count(&stats.Completed)

	r.db.Model(&domain.Complaint{}).
		Joins("JOIN module_complaint_activities act ON act.module_complaint_id = module_complaints.id").
		Where("act.status = ? AND module_complaints.status = ?", domain.ActivityStatusDisputeRequest, domain.ComplaintStatusPending).
		Distinct("module_complaints.id").
		Count(&stats.Disputed)
	return stats, nil
}

func (r *complaintRepository) CreateRatingHistory(history *domain.ComplaintRatingHistory) error {
	return r.db.Create(history).Error
}

func (r *complaintRepository) GetCompleterInfo(complaintID string) (*string, *string, error) {

	var complaint domain.Complaint
	errComp := r.db.Select("id, assignee_id, department_id").First(&complaint, "id = ?", complaintID).Error

	var completedBy string
	err := r.db.Table("module_complaint_activities").
		Where("module_complaint_id = ? AND status = ?", complaintID, "completed").
		Order("created_date DESC").
		Limit(1).
		Pluck("created_by", &completedBy).
		Error

	var assigneeIDPtr *string
	if err == nil && completedBy != "" {
		assigneeIDPtr = &completedBy
	} else if errComp == nil {
		assigneeIDPtr = complaint.AssigneeId
	}

	var deptIDPtr *string
	if errComp == nil && complaint.DepartmentId != nil {

		deptIDPtr = complaint.DepartmentId
	} else if assigneeIDPtr != nil {

		var deptID string
		errDept := r.db.Table("admin_departments").
			Where("admin_id = ?", *assigneeIDPtr).
			Limit(1).
			Pluck("department_id", &deptID).
			Error
		if errDept == nil && deptID != "" {
			deptIDPtr = &deptID
		}
	}

	return assigneeIDPtr, deptIDPtr, nil
}
