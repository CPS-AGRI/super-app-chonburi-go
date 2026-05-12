package usecase

import (
	"super-app-chonburi-go/internal/domain"
	"time"
	"github.com/google/uuid"
)

type complaintUseCase struct {
	complaintRepo domain.ComplaintRepository
	adminRepo     domain.AdminRepository
}

func NewComplaintUseCase(complaintRepo domain.ComplaintRepository, adminRepo domain.AdminRepository) domain.ComplaintUseCase {
	return &complaintUseCase{
		complaintRepo: complaintRepo,
		adminRepo:     adminRepo,
	}
}

func (u *complaintUseCase) GetComplaints(query domain.ComplaintQuery, adminID string) (*domain.PaginatedComplaintResponse, error) {
	admin, err := u.adminRepo.GetByID(adminID)
	if err != nil {
		return nil, err
	}

	query.IsSuperAdmin = admin.Role != nil && admin.Role.Type == "super_admin"

	// Populate Admin Department IDs
	var deptIDs []string
	var allowedModuleTypeIDs []string
	isCenter := false

	for _, dept := range admin.Departments {
		deptIDs = append(deptIDs, dept.ID)
		
		// Check if this department has the Complaint Center module
		for _, mod := range dept.Modules {
			if mod.ID == "d01b2ce5-34a9-498b-bba0-b1b8360f1ea9" || 
			   mod.NameTh == "ศูนย์ร้องทุกข์" || 
			   mod.NameEn == "Complaint Center" {
				isCenter = true
			}
		}

		// Collect allowed module type IDs for Mode 1 fallback
		for _, mod := range dept.Modules {
			for _, mt := range mod.ModuleTypes {
				allowedModuleTypeIDs = append(allowedModuleTypeIDs, mt.ID)
			}
		}
		// Also check raw IDs if present on the department object
		if len(dept.ModuleTypeIds) > 0 {
			allowedModuleTypeIDs = append(allowedModuleTypeIDs, dept.ModuleTypeIds...)
		}
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

func (u *complaintUseCase) ForwardComplaint(id string, departmentID string, adminID string) error {
	complaint, err := u.complaintRepo.GetByID(id, nil, true)
	if err != nil {
		return err
	}

	complaint.DepartmentId = &departmentID
	complaint.Status = domain.ComplaintStatusReceived
	complaint.UpdatedDate = time.Now()

	activity := &domain.ComplaintActivity{
		ID:                uuid.New(),
		ModuleComplaintId: uuid.MustParse(id),
		Description:       "มอบหมายงานไปยังหน่วยงานที่เกี่ยวข้อง (รับเรื่องแล้ว)",
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

func (u *complaintUseCase) AssignComplaint(id string, assigneeID string, adminID string) error {
	complaint, err := u.complaintRepo.GetByID(id, nil, true)
	if err != nil {
		return err
	}

	complaint.AssigneeId = &assigneeID
	complaint.Status = domain.ComplaintStatusInProgress
	complaint.UpdatedDate = time.Now()

	activity := &domain.ComplaintActivity{
		ID:                uuid.New(),
		ModuleComplaintId: uuid.MustParse(id),
		Description:       "มอบหมายงานให้ผู้รับผิดชอบ (กำลังดำเนินงาน)",
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

	complaint.DepartmentId = nil // Clear department to return to center
	complaint.Status = domain.ComplaintStatusRejected
	complaint.UpdatedDate = time.Now()

	activity := &domain.ComplaintActivity{
		ID:                uuid.New(),
		ModuleComplaintId: uuid.MustParse(id),
		Description:       "ตีกลับเรื่องไปยังศูนย์ร้องทุกข์: " + reason,
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

	return u.complaintRepo.CreateActivity(activity)
}

func (u *complaintUseCase) DeleteComplaint(id string) error {
	return u.complaintRepo.Delete(id)
}
