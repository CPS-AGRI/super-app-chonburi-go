package usecase

import (
	"super-app-chonburi-go/internal/domain"
	"time"
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
	complaint.ID = domain.NewUUID()
	complaint.Status = domain.ComplaintStatusReceived
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
		ID:                domain.NewUUID(),
		ModuleComplaintId: id,
		Description:       description,
		Status:            status,
		CreatedBy:         adminID,
		UpdatedBy:         adminID,
		CreatedDate:       time.Now(),
		UpdatedDate:       time.Now(),
	}

	for i, imgUrl := range images {
		activity.Images = append(activity.Images, domain.ComplaintActivityImage{
			ID:          domain.NewUUID(),
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
	complaint.Status = domain.ComplaintStatusInProgress
	complaint.UpdatedDate = time.Now()

	activity := &domain.ComplaintActivity{
		ID:                domain.NewUUID(),
		ModuleComplaintId: id,
		Description:       "ส่งเรื่องต่อไปยังหน่วยงานที่เกี่ยวข้อง",
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
	complaint.UpdatedDate = time.Now()

	activity := &domain.ComplaintActivity{
		ID:                domain.NewUUID(),
		ModuleComplaintId: id,
		Description:       "มอบหมายงานให้ผู้รับผิดชอบ",
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
	return u.UpdateComplaintStatus(id, domain.ComplaintStatusRejected, "ปฏิเสธคำร้อง: "+reason, adminID, nil)
}

func (u *complaintUseCase) AddActivity(activity *domain.ComplaintActivity, adminID string) error {
	activity.ID = domain.NewUUID()
	activity.CreatedBy = adminID
	activity.UpdatedBy = adminID
	activity.CreatedDate = time.Now()
	activity.UpdatedDate = time.Now()
	
	for i := range activity.Images {
		activity.Images[i].ID = domain.NewUUID()
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
