package usecase

import (
	"errors"
	"super-app-chonburi-go/internal/domain"

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

// Helper function to resolve permissions allowed for the admin
func (u *complaintUseCase) resolveAdminPermissions(adminID uuid.UUID) ([]string, bool, error) {
	admin, err := u.adminRepo.GetByID(adminID)
	if err != nil {
		return nil, false, err
	}
	if admin == nil {
		return nil, false, errors.New("admin not found")
	}

	isSuperAdmin := admin.Role != nil && admin.Role.IsSuperAdmin

	if isSuperAdmin {
		return nil, true, nil
	}

	// Collect unique permissions from all departments
	permMap := make(map[string]bool)
	for _, dept := range admin.Departments {
		for _, perm := range dept.Permissions {
			permMap[perm.ID] = true
		}
	}

	var allowedPerms []string
	for p := range permMap {
		allowedPerms = append(allowedPerms, p)
	}

	return allowedPerms, false, nil
}

func (u *complaintUseCase) GetComplaints(query domain.ComplaintQuery, adminID uuid.UUID) (*domain.PaginatedComplaintResponse, error) {
	allowedPerms, isSuperAdmin, err := u.resolveAdminPermissions(adminID)
	if err != nil {
		return nil, err
	}

	query.AllowedPermissionIDs = allowedPerms
	query.IsSuperAdmin = isSuperAdmin

	return u.complaintRepo.GetPaginated(query)
}

func (u *complaintUseCase) GetComplaintByID(id uuid.UUID, adminID uuid.UUID) (*domain.Complaint, error) {
	allowedPerms, isSuperAdmin, err := u.resolveAdminPermissions(adminID)
	if err != nil {
		return nil, err
	}

	complaint, err := u.complaintRepo.GetByID(id, allowedPerms, isSuperAdmin)
	if err == nil && complaint != nil {
		return complaint, nil
	}

	rawComplaint, err := u.complaintRepo.GetByID(id, nil, true)
	if err == nil && rawComplaint != nil && rawComplaint.AssigneeID != nil && *rawComplaint.AssigneeID == adminID {
		return rawComplaint, nil
	}

	return complaint, err
}

func (u *complaintUseCase) CreateComplaint(complaint *domain.Complaint) error {
	// Status defaults to Received
	complaint.Status = domain.ComplaintStatusReceived
	return u.complaintRepo.Create(complaint)
}

func (u *complaintUseCase) AssignComplaint(id uuid.UUID, assignerID uuid.UUID, assigneeID uuid.UUID) error {
	// Must verify assigner and assignee are in the same department, but for now we just allow managers to assign
	// Get complaint first to ensure they have permission to access it
	complaint, err := u.GetComplaintByID(id, assignerID)
	if err != nil {
		return err
	}
	if complaint == nil {
		return errors.New("complaint not found or you don't have permission")
	}

	// Optional: Check if assigner is a Manager (Role Logic). We skip strict role checks here if not required.

	complaint.AssignerID = &assignerID
	complaint.AssigneeID = &assigneeID
	complaint.Status = domain.ComplaintStatusInProgress

	return u.complaintRepo.Update(complaint)
}

func (u *complaintUseCase) RejectComplaint(id uuid.UUID, rejecterID uuid.UUID, reason string) error {
	complaint, err := u.GetComplaintByID(id, rejecterID)
	if err != nil {
		return err
	}
	if complaint == nil {
		return errors.New("complaint not found")
	}

	complaint.RejectedByID = &rejecterID
	complaint.Status = domain.ComplaintStatusRejected

	// Add an activity for rejection
	activity := &domain.ComplaintActivity{
		ComplaintID: id,
		AdminID:     &rejecterID,
		Description: "Rejected: " + reason,
		Status:      domain.ComplaintStatusRejected,
	}

	// Save complaint first
	err = u.complaintRepo.Update(complaint)
	if err != nil {
		return err
	}

	return u.complaintRepo.CreateActivity(activity)
}

func (u *complaintUseCase) AddActivity(activity *domain.ComplaintActivity, adminID uuid.UUID) error {
	complaint, err := u.GetComplaintByID(activity.ComplaintID, adminID)
	if err != nil {
		return err
	}
	if complaint == nil {
		return errors.New("complaint not found")
	}

	activity.AdminID = &adminID

	// If activity specifies a new status (e.g. Completed), update the complaint as well
	if activity.Status != "" && activity.Status != complaint.Status {
		complaint.Status = activity.Status
		if err := u.complaintRepo.Update(complaint); err != nil {
			return err
		}
	} else {
		// Just take current status
		activity.Status = complaint.Status
	}

	return u.complaintRepo.CreateActivity(activity)
}

func (u *complaintUseCase) DeleteComplaint(id uuid.UUID, adminID uuid.UUID) error {
	// Usually only Citizens or SuperAdmins delete complaints.
	// We'll allow it if they can view it and are super admin or something, but standard allows deletion if user owns it.
	return u.complaintRepo.Delete(id)
}
