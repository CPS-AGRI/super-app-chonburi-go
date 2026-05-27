package usecase

import (
	"encoding/json"
	"fmt"
	"strings"
	"super-app-chonburi-go/internal/domain"
	"time"

	"github.com/google/uuid"
)

type complaintUseCase struct {
	complaintRepo domain.ComplaintRepository
	adminRepo     domain.AdminRepository
	muniRepo      domain.MunicipalityRepository
}

func NewComplaintUseCase(complaintRepo domain.ComplaintRepository, adminRepo domain.AdminRepository, muniRepo domain.MunicipalityRepository) domain.ComplaintUseCase {
	return &complaintUseCase{
		complaintRepo: complaintRepo,
		adminRepo:     adminRepo,
		muniRepo:      muniRepo,
	}
}

func (u *complaintUseCase) GetComplaints(query domain.ComplaintQuery, adminID string) (*domain.PaginatedComplaintResponse, error) {
	admin, err := u.adminRepo.GetByID(adminID)
	if err != nil {
		return nil, err
	}

	query.IsSuperAdmin = admin.Role != nil && admin.Role.Type == "super_admin"
	if admin.Role != nil {
		query.AdminRoleType = admin.Role.Type
	}

	// Populate Admin Department IDs
	var deptIDs []string
	var allowedModuleTypeIDs []string
	isCenter := false

	for _, dept := range admin.Departments {
		deptIDs = append(deptIDs, dept.ID)

		// Check if this department has the Complaint Center module
		for _, mod := range dept.Modules {
			nameTh := strings.ToLower(mod.NameTh)
			nameEn := strings.ToLower(mod.NameEn)
			if mod.ID == "d01b2ce5-34a9-498b-bba0-b1b8360f1ea9" ||
				strings.Contains(nameTh, "ศูนย์ร้องทุกข์") ||
				strings.Contains(nameTh, "ศูนย์รับเรื่อง") ||
				strings.Contains(nameEn, "complaint center") {
				isCenter = true
			}
		}
	}

	// Fix Filtering Bug: Get strictly assigned ModuleType IDs from bridge table
	if len(deptIDs) > 0 {
		allowedModuleTypeIDs, _ = u.complaintRepo.GetAllowedModuleTypeIDs(deptIDs)
	}

	query.AdminDepartmentIDs = deptIDs
	query.IsComplaintCenter = isCenter
	query.AllowedModuleTypeIDs = allowedModuleTypeIDs
	
	// Default pagination
	if query.PageNumber < 1 { query.PageNumber = 1 }
	if query.PageSize < 1 { query.PageSize = 10 }

	return u.complaintRepo.GetPaginated(query)
}

func (u *complaintUseCase) GetComplaintByID(id string, adminID string) (*domain.Complaint, error) {
	admin, err := u.adminRepo.GetByID(adminID)
	if err != nil {
		return nil, err
	}

	return u.complaintRepo.GetByID(id, nil, admin.Role != nil && admin.Role.Type == "super_admin")
}

func (u *complaintUseCase) CreateComplaint(complaint *domain.Complaint) error {
	complaint.ID = uuid.New()
	complaint.Status = domain.ComplaintStatusPending
	complaint.CreatedDate = time.Now()
	complaint.UpdatedDate = time.Now()

	return u.complaintRepo.Create(complaint)
}

func (u *complaintUseCase) UpdateComplaintStatus(id string, status string, description string, adminID string, images []string) error {
	complaint, err := u.complaintRepo.GetByID(id, nil, true)
	if err != nil {
		return err
	}


	complaint.Status = status
	complaint.UpdatedDate = time.Now()

	activity := &domain.ComplaintActivity{
		ID:                uuid.New(),
		ModuleComplaintId: uuid.MustParse(id),
		Description:       description,
		Status:            status,
		CreatedBy:         adminID,
		UpdatedBy:         adminID,
		CreatedDate:       time.Now(),
		UpdatedDate:       time.Now(),
	}

	for i, imgUrl := range images {
		activity.Images = append(activity.Images, domain.ComplaintActivityImage{
			ID:          uuid.New(),
			Url:         imgUrl,
			Sequence:    i + 1,
			CreatedBy:   adminID,
			UpdatedBy:   adminID,
			CreatedDate: time.Now(),
			UpdatedDate: time.Now(),
		})
	}

	if err := u.complaintRepo.CreateActivity(activity); err != nil {
		return err
	}

	return u.complaintRepo.Update(complaint)
}

func (u *complaintUseCase) ForwardComplaint(id string, departmentID string, description string, adminID string) error {
	complaint, err := u.complaintRepo.GetByID(id, nil, true)
	if err != nil {
		return err
	}

	complaint.DepartmentId = &departmentID
	complaint.Status = domain.ComplaintStatusReceived
	complaint.UpdatedDate = time.Now()

	desc := description
	if desc == "" {
		desc = "มอบหมายงานไปยังหน่วยงานที่เกี่ยวข้อง (รับเรื่องแล้ว)"
	}

	activity := &domain.ComplaintActivity{
		ID:                uuid.New(),
		ModuleComplaintId: uuid.MustParse(id),
		Description:       desc,
		Status:            complaint.Status,
		CreatedBy:         adminID,
		UpdatedBy:         adminID,
		CreatedDate:       time.Now(),
		UpdatedDate:       time.Now(),
	}

	if err := u.complaintRepo.CreateActivity(activity); err != nil {
		return err
	}

	return u.complaintRepo.Update(complaint)
}

func (u *complaintUseCase) AssignComplaint(id string, assigneeID string, description string, adminID string) error {
	complaint, err := u.complaintRepo.GetByID(id, nil, true)
	if err != nil {
		return err
	}

	complaint.AssigneeId = &assigneeID
	complaint.Status = domain.ComplaintStatusInProgress
	complaint.UpdatedDate = time.Now()

	desc := description
	if desc == "" {
		desc = "มอบหมายงานให้ผู้รับผิดชอบ (กำลังดำเนินงาน)"
	}

	activity := &domain.ComplaintActivity{
		ID:                uuid.New(),
		ModuleComplaintId: uuid.MustParse(id),
		Description:       desc,
		Status:            complaint.Status,
		CreatedBy:         adminID,
		UpdatedBy:         adminID,
		CreatedDate:       time.Now(),
		UpdatedDate:       time.Now(),
	}

	if err := u.complaintRepo.CreateActivity(activity); err != nil {
		return err
	}

	return u.complaintRepo.Update(complaint)
}

func (u *complaintUseCase) RejectComplaint(id string, reason string, adminID string) error {
	complaint, err := u.complaintRepo.GetByID(id, nil, true)
	if err != nil {
		return err
	}

	// Handle Workflow Modes
	muni, err := u.muniRepo.GetFirst()
	status := domain.ComplaintStatusRejected
	var activityDesc string

	// Always clear department on reject so it disappears from the department's manage table
	complaint.DepartmentId = nil

	if err == nil && muni.ComplaintMode == "direct" {
		// Option 1: Direct Mode - mark as rejected for internal review/history
		activityDesc = "ตีกลับเรื่อง (ถูกปฏิเสธ): " + reason
	} else {
		// Option 2: Central Mode - Return to Center
		activityDesc = "ตีกลับเรื่องไปยังศูนย์ร้องทุกข์: " + reason
	}

	complaint.Status = status
	complaint.UpdatedDate = time.Now()

	activity := &domain.ComplaintActivity{
		ID:                uuid.New(),
		ModuleComplaintId: uuid.MustParse(id),
		Description:       activityDesc,
		Status:            status,
		CreatedBy:         adminID,
		UpdatedBy:         adminID,
		CreatedDate:       time.Now(),
		UpdatedDate:       time.Now(),
	}

	if err := u.complaintRepo.CreateActivity(activity); err != nil {
		return err
	}

	return u.complaintRepo.Update(complaint)
}

func (u *complaintUseCase) AddActivity(activity *domain.ComplaintActivity, adminID string) error {
	activity.ID = uuid.New()
	activity.CreatedBy = adminID
	activity.UpdatedBy = adminID
	activity.CreatedDate = time.Now()
	activity.UpdatedDate = time.Now()
	
	for i := range activity.Images {
		activity.Images[i].ID = uuid.New()
		activity.Images[i].CreatedBy = adminID
		activity.Images[i].UpdatedBy = adminID
		activity.Images[i].CreatedDate = time.Now()
		activity.Images[i].UpdatedDate = time.Now()
		if activity.Images[i].Sequence == 0 {
			activity.Images[i].Sequence = i + 1
		}
	}

	// Fetch main complaint to update its UpdatedDate and Status
	complaint, err := u.complaintRepo.GetByID(activity.ModuleComplaintId.String(), nil, true)
	if err == nil && complaint != nil {
		if activity.Status != "" {
			complaint.Status = activity.Status
		}
		complaint.UpdatedDate = time.Now()
		_ = u.complaintRepo.Update(complaint)
	}

	return u.complaintRepo.CreateActivity(activity)
}

func (u *complaintUseCase) DeleteComplaint(id string) error {
	return u.complaintRepo.Delete(id)
}

func (u *complaintUseCase) RateComplaint(id string, userID string, rating int, comment string) error {
	complaint, err := u.complaintRepo.GetByID(id, nil, true)
	if err != nil {
		return err
	}

	if complaint.Status != domain.ComplaintStatusCompleted {
		return fmt.Errorf("can only rate a completed complaint")
	}

	// 1. Get completer info for assignee_id and department_id
	assigneeID, departmentID, _ := u.complaintRepo.GetCompleterInfo(id)

	// 2. Create history record
	history := &domain.ComplaintRatingHistory{
		ID:                uuid.New(),
		ModuleComplaintId: uuid.MustParse(id),
		AssigneeId:        assigneeID,
		DepartmentId:      departmentID,
		RatingScore:       &rating,
		IsDisputed:        false,
		CreatedDate:       time.Now(),
		UpdatedDate:       time.Now(),
		CreatedBy:         userID,
		UpdatedBy:         userID,
	}
	if err := u.complaintRepo.CreateRatingHistory(history); err != nil {
		return err
	}

	// 3. Reset active dispute flag on the main complaint
	complaint.IsDisputed = false 
	if err := u.complaintRepo.Update(complaint); err != nil {
		return err
	}

	descBytes, _ := json.Marshal(domain.ComplaintRating{Rating: rating, Comment: comment})

	activity := &domain.ComplaintActivity{
		ID:                uuid.New(),
		ModuleComplaintId: uuid.MustParse(id),
		Description:       string(descBytes),
		Status:            domain.ActivityStatusUserRating,
		CreatedBy:         userID,
		UpdatedBy:         userID,
		CreatedDate:       time.Now(),
		UpdatedDate:       time.Now(),
	}

	return u.complaintRepo.CreateActivity(activity)
}

func (u *complaintUseCase) DisputeComplaint(id string, userID string, reason string, images []string) error {
	complaint, err := u.complaintRepo.GetByID(id, nil, true)
	if err != nil {
		return err
	}

	if complaint.Status != domain.ComplaintStatusCompleted {
		return fmt.Errorf("can only dispute a completed complaint")
	}

	// 1. Get completer info for assignee_id and department_id (the person who failed to complete the job)
	assigneeID, departmentID, _ := u.complaintRepo.GetCompleterInfo(id)

	// 2. Create history record
	history := &domain.ComplaintRatingHistory{
		ID:                uuid.New(),
		ModuleComplaintId: uuid.MustParse(id),
		AssigneeId:        assigneeID,
		DepartmentId:      departmentID,
		RatingScore:       nil,
		IsDisputed:        true,
		CreatedDate:       time.Now(),
		UpdatedDate:       time.Now(),
		CreatedBy:         userID,
		UpdatedBy:         userID,
	}
	if err := u.complaintRepo.CreateRatingHistory(history); err != nil {
		return err
	}

	muni, muniErr := u.muniRepo.GetFirst()

	activityDesc := "ลูกค้าไม่พอใจผลการดำเนินงาน (direct): " + reason
	if muniErr == nil && muni.ComplaintMode == "central" {
		activityDesc = "ลูกค้าไม่พอใจผลการดำเนินงาน (central): " + reason
		complaint.DepartmentId = nil
	}

	complaint.AssigneeId = nil
	complaint.Status = domain.ComplaintStatusPending
	complaint.UpdatedDate = time.Now()
	complaint.IsDisputed = true // Mark as disputed in db to flag current active dispute state

	activity := &domain.ComplaintActivity{
		ID:                uuid.New(),
		ModuleComplaintId: uuid.MustParse(id),
		Description:       activityDesc,
		Status:            domain.ActivityStatusDisputeRequest,
		CreatedBy:         userID,
		UpdatedBy:         userID,
		CreatedDate:       time.Now(),
		UpdatedDate:       time.Now(),
	}

	for i, imgUrl := range images {
		activity.Images = append(activity.Images, domain.ComplaintActivityImage{
			ID:                        uuid.New(),
			ModuleComplaintActivityId: activity.ID,
			Url:                       imgUrl,
			Sequence:                  i + 1,
			CreatedBy:                 userID,
			UpdatedBy:                 userID,
			CreatedDate:               time.Now(),
			UpdatedDate:               time.Now(),
		})
	}

	if err := u.complaintRepo.CreateActivity(activity); err != nil {
		return err
	}

	return u.complaintRepo.Update(complaint)
}

func (u *complaintUseCase) GetRatingSummaries(summaryType string) ([]domain.ComplaintRatingSummary, error) {
	return u.complaintRepo.GetRatingSummaries(summaryType)
}

func (u *complaintUseCase) GetOverviewStats() (*domain.ComplaintOverviewStats, error) {
	return u.complaintRepo.GetOverviewStats()
}
