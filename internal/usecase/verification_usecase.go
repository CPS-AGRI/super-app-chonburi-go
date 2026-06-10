package usecase

import (
	"errors"
	"time"

	"super-app-chonburi-go/internal/domain"
	"super-app-chonburi-go/pkg/firebase"

	"github.com/google/uuid"
)

type verificationUseCase struct {
	repo domain.AdminVerificationRepository
}

func NewVerificationUseCase(repo domain.AdminVerificationRepository) domain.AdminVerificationUseCase {
	return &verificationUseCase{repo: repo}
}

func (u *verificationUseCase) GetVerifications(query domain.VerificationQuery) (*domain.PaginatedVerificationResponse, error) {
	if query.PageNumber <= 0 {
		query.PageNumber = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 10
	}
	return u.repo.GetPaginated(query)
}

func (u *verificationUseCase) GetVerificationByID(userID uuid.UUID) (*domain.UserVerificationItem, error) {
	return u.repo.GetByID(userID)
}

func (u *verificationUseCase) ApproveVerification(req *domain.ApproveVerificationRequest, adminUserID string) error {
	userID := req.UserID
	item, err := u.repo.GetByID(userID)
	if err != nil {
		return err
	}
	if item == nil {
		return errors.New("verification request not found")
	}

	if item.VerificationStatus != string(domain.VerificationStatusPending) {
		return errors.New("cannot approve verification in status: " + item.VerificationStatus)
	}

	if err := u.repo.Approve(req, adminUserID); err != nil {
		return err
	}

	moduleIDPtr, err := u.repo.GetRegisterModuleID()
	var moduleID uuid.UUID
	if err != nil || moduleIDPtr == nil {

		moduleID = uuid.MustParse("00000000-0000-0000-0000-000000000000")
	} else {
		moduleID = *moduleIDPtr
	}

	title := "ผลการอนุมัติการยืนยันตัวตน"
	body := "การยืนยันตัวตนของคุณได้รับการอนุมัติเรียบร้อยแล้ว"

	adminUUID, _ := uuid.Parse(adminUserID)

	notif := &domain.ModuleNotification{
		ID:              uuid.New(),
		ModuleID:        moduleID,
		UserID:          &userID,
		UserAdminID:     &adminUUID,
		ReferenceID:     userID.String(),
		ReferenceTitle:  title,
		ReferenceBody:   body,
		ReferenceStatus: "approved",
		Type:            "user",
		Status:          "published",
		State:           "unread",
		CreatedBy:       adminUserID,
		CreatedDate:     time.Now(),
		UpdatedBy:       adminUserID,
		UpdatedDate:     time.Now(),
	}

	if err := u.repo.CreateNotification(notif); err != nil {

		println("Warning: failed to create DB notification record:", err.Error())
	}

	tokens, err := u.repo.GetFCMTokens(userID)
	if err == nil && len(tokens) > 0 {
		firebase.SendPushNotification(tokens, title, body)
	}

	return nil
}

func (u *verificationUseCase) RejectVerification(userID uuid.UUID, reason string, adminUserID string) error {
	item, err := u.repo.GetByID(userID)
	if err != nil {
		return err
	}
	if item == nil {
		return errors.New("verification request not found")
	}

	if item.VerificationStatus != string(domain.VerificationStatusPending) {
		return errors.New("cannot reject verification in status: " + item.VerificationStatus)
	}

	if err := u.repo.Reject(userID, reason, adminUserID); err != nil {
		return err
	}

	moduleIDPtr, err := u.repo.GetRegisterModuleID()
	var moduleID uuid.UUID
	if err != nil || moduleIDPtr == nil {
		moduleID = uuid.MustParse("00000000-0000-0000-0000-000000000000")
	} else {
		moduleID = *moduleIDPtr
	}

	title := "ผลการอนุมัติการยืนยันตัวตน"
	body := "การยืนยันตัวตนของคุณไม่ผ่านการอนุมัติ เนื่องจาก: " + reason

	adminUUID, _ := uuid.Parse(adminUserID)

	notif := &domain.ModuleNotification{
		ID:              uuid.New(),
		ModuleID:        moduleID,
		UserID:          &userID,
		UserAdminID:     &adminUUID,
		ReferenceID:     userID.String(),
		ReferenceTitle:  title,
		ReferenceBody:   body,
		ReferenceStatus: "rejected",
		Type:            "user",
		Status:          "published",
		State:           "unread",
		CreatedBy:       adminUserID,
		CreatedDate:     time.Now(),
		UpdatedBy:       adminUserID,
		UpdatedDate:     time.Now(),
	}

	if err := u.repo.CreateNotification(notif); err != nil {
		println("Warning: failed to create DB notification record:", err.Error())
	}

	tokens, err := u.repo.GetFCMTokens(userID)
	if err == nil && len(tokens) > 0 {
		firebase.SendPushNotification(tokens, title, body)
	}

	return nil
}
