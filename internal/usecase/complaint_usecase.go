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

	var deptIDs []string
	var allowedModuleTypeIDs []string
	isCenter := false

	for _, dept := range admin.Departments {
		deptIDs = append(deptIDs, dept.ID)

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

	if len(deptIDs) > 0 {
		allowedModuleTypeIDs, _ = u.complaintRepo.GetAllowedModuleTypeIDs(deptIDs)
	}

	query.AdminDepartmentIDs = deptIDs
	query.IsComplaintCenter = isCenter
	query.AllowedModuleTypeIDs = allowedModuleTypeIDs

	if query.PageNumber < 1 {
		query.PageNumber = 1
	}
	if query.PageSize < 1 {
		query.PageSize = 10
	}

	return u.complaintRepo.GetPaginated(query)
}

func (u *complaintUseCase) GetComplaintByID(id string, adminID string) (*domain.Complaint, error) {
	admin, err := u.adminRepo.GetByID(adminID)
	if err != nil {
		return nil, err
	}

	return u.complaintRepo.GetByID(id, nil, admin.Role != nil && admin.Role.Type == "super_admin")
}

func getThaiComplaintStatus(status string) string {
	switch status {
	case domain.ComplaintStatusPending:
		return "รอดำเนินการ"
	case domain.ComplaintStatusReceived:
		return "รับเรื่องแล้ว"
	case domain.ComplaintStatusInProgress:
		return "กำลังดำเนินการ"
	case domain.ComplaintStatusCompleted:
		return "ดำเนินการเสร็จสิ้น"
	case domain.ComplaintStatusRejected:
		return "ปฏิเสธคำร้อง"
	default:
		return status
	}
}

func (u *complaintUseCase) CreateComplaint(complaint *domain.Complaint) error {
	complaint.ID = uuid.New()
	complaint.Status = domain.ComplaintStatusPending
	complaint.CreatedDate = time.Now()
	complaint.UpdatedDate = time.Now()

	err := u.complaintRepo.Create(complaint)
	if err != nil {
		return err
	}

	SendNotificationToUser(
		complaint.UserId,
		"ยื่นคำร้องร้องเรียนสำเร็จ",
		fmt.Sprintf("เราได้รับเรื่องร้องเรียนเลขที่ %s เรียบร้อยแล้ว (สถานะ: รอดำเนินการ)", complaint.DocumentId),
		complaint.ID.String(),
		"pending",
		"complaint",
	)

	deptID := ""
	if complaint.DepartmentId != nil {
		deptID = *complaint.DepartmentId
	}
	SendNotificationToDepartment(
		deptID,
		"Managers",
		"มีเรื่องร้องเรียนใหม่ส่งเข้ามา",
		fmt.Sprintf("เรื่องร้องเรียนเลขที่ %s เรื่อง: %s รอนุมัติรับเรื่องและส่งต่อทำงาน", complaint.DocumentId, complaint.Description),
		complaint.ID.String(),
		"pending",
	)

	return nil
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

	err = u.complaintRepo.Update(complaint)
	if err != nil {
		return err
	}

	SendNotificationToUser(
		complaint.UserId,
		"อัปเดตสถานะเรื่องร้องเรียน",
		fmt.Sprintf("เรื่องร้องเรียนเลขที่ %s ของท่านเปลี่ยนสถานะเป็น: %s (%s)", complaint.DocumentId, getThaiComplaintStatus(status), description),
		complaint.ID.String(),
		status,
		"complaint",
	)

	return nil
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

	err = u.complaintRepo.Update(complaint)
	if err != nil {
		return err
	}

	SendNotificationToDepartment(
		departmentID,
		"Managers",
		"มีเรื่องร้องเรียนใหม่ส่งมอบหมายเข้ามา",
		fmt.Sprintf("เรื่องร้องเรียนเลขที่ %s ได้รับการมอบหมายมายังแผนกของท่านเพื่อประสานงานดำเนินงาน", complaint.DocumentId),
		complaint.ID.String(),
		"received",
	)

	SendNotificationToUser(
		complaint.UserId,
		"เรื่องร้องเรียนถูกส่งมอบหมาย",
		fmt.Sprintf("เรื่องร้องเรียนเลขที่ %s ของท่านได้รับการรับเรื่องแล้วและอยู่ระหว่างส่งมอบหมายหน่วยงานดูแล", complaint.DocumentId),
		complaint.ID.String(),
		"received",
		"complaint",
	)

	return nil
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

	err = u.complaintRepo.Update(complaint)
	if err != nil {
		return err
	}

	if parsedAssignee, err := uuid.Parse(assigneeID); err == nil {
		SendNotificationToAdmin(
			parsedAssignee,
			"ท่านได้รับมอบหมายงานร้องทุกข์ใหม่",
			fmt.Sprintf("คุณได้รับการแต่งตั้งเป็นผู้ดูแลเคสรหัส %s โปรดดำเนินการตรวจสอบเรื่องและอัปเดตความก้าวหน้า", complaint.DocumentId),
			complaint.ID.String(),
			"in_progress",
			"complaint_assignee",
		)
	}

	SendNotificationToUser(
		complaint.UserId,
		"เจ้าหน้าที่เริ่มดำเนินงานร้องเรียนของท่าน",
		fmt.Sprintf("เรื่องร้องเรียนเลขที่ %s ของท่าน อยู่ระหว่างการลงพื้นที่จัดการโดยเจ้าหน้าที่ฝ่ายปฏิบัติงาน", complaint.DocumentId),
		complaint.ID.String(),
		"in_progress",
		"complaint",
	)

	return nil
}

func (u *complaintUseCase) RejectComplaint(id string, reason string, adminID string) error {
	complaint, err := u.complaintRepo.GetByID(id, nil, true)
	if err != nil {
		return err
	}

	muni, err := u.muniRepo.GetFirst()
	status := domain.ComplaintStatusRejected
	var activityDesc string

	complaint.DepartmentId = nil

	if err == nil && muni.ComplaintMode == "direct" {

		activityDesc = "ตีกลับเรื่อง (ถูกปฏิเสธ): " + reason
	} else {

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

	err = u.complaintRepo.Update(complaint)
	if err != nil {
		return err
	}

	SendNotificationToUser(
		complaint.UserId,
		"เรื่องร้องเรียนได้รับการปฏิเสธคำร้อง/ตีกลับ",
		fmt.Sprintf("คำร้องขอนอกเหนือนโยบาย เลขที่ %s: %s", complaint.DocumentId, reason),
		complaint.ID.String(),
		"rejected",
		"complaint",
	)

	deptID := ""
	if complaint.DepartmentId != nil {
		deptID = *complaint.DepartmentId
	}
	SendNotificationToDepartment(
		deptID,
		"Managers",
		"เรื่องร้องเรียนได้รับการปฏิเสธคำร้อง/ตีกลับ",
		fmt.Sprintf("เรื่องร้องเรียนเลขที่ %s ถูกตีกลับ/ปฏิเสธโดยเจ้าหน้าที่: %s", complaint.DocumentId, reason),
		complaint.ID.String(),
		"rejected",
	)

	return nil
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

	complaint, err := u.complaintRepo.GetByID(activity.ModuleComplaintId.String(), nil, true)
	if err == nil && complaint != nil {
		if activity.Status != "" {
			oldStatus := complaint.Status
			complaint.Status = activity.Status
			complaint.UpdatedDate = time.Now()
			_ = u.complaintRepo.Update(complaint)

			if oldStatus != activity.Status {
				SendNotificationToUser(
					complaint.UserId,
					"อัปเดตสถานะเรื่องร้องเรียน",
					fmt.Sprintf("เรื่องร้องเรียนเลขที่ %s ของท่านเปลี่ยนสถานะเป็น: %s (%s)", complaint.DocumentId, getThaiComplaintStatus(activity.Status), activity.Description),
					complaint.ID.String(),
					activity.Status,
					"complaint",
				)

				deptID := ""
				if complaint.DepartmentId != nil {
					deptID = *complaint.DepartmentId
				}
				SendNotificationToDepartment(
					deptID,
					"Managers",
					"มีการอัปเดตสถานะเรื่องร้องเรียน",
					fmt.Sprintf("เรื่องร้องเรียนเลขที่ %s เปลี่ยนสถานะเป็น: %s โดยเจ้าหน้าที่ (%s)", complaint.DocumentId, getThaiComplaintStatus(activity.Status), activity.Description),
					complaint.ID.String(),
					activity.Status,
				)
			}
		} else {
			complaint.UpdatedDate = time.Now()
			_ = u.complaintRepo.Update(complaint)
		}
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

	assigneeID, departmentID, _ := u.complaintRepo.GetCompleterInfo(id)

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

	assigneeID, departmentID, _ := u.complaintRepo.GetCompleterInfo(id)

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
	if muniErr == nil && muni.ComplaintMode == "central" {
		complaint.DepartmentId = nil
	}

	complaint.AssigneeId = nil
	complaint.Status = domain.ComplaintStatusPending
	complaint.UpdatedDate = time.Now()
	complaint.IsDisputed = true

	activity := &domain.ComplaintActivity{
		ID:                uuid.New(),
		ModuleComplaintId: uuid.MustParse(id),
		Description:       reason,
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
