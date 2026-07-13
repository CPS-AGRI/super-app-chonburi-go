package usecase

import (
	"log"
	"super-app-chonburi-go/internal/domain"
	"super-app-chonburi-go/pkg/database"
	"time"

	"github.com/google/uuid"
)

func SendNotificationToUser(userID uuid.UUID, title, body, refID, refStatus, notifType string) {
	db := database.DB
	if db == nil {
		log.Println("[Notif-Helper] ข้อผิดพลาด: database.DB ยังไม่ได้เชื่อมต่อ")
		return
	}

	var moduleID uuid.UUID
	switch notifType {
	case "complaint", "complaint_assignee":
		moduleID = uuid.MustParse("5b630777-4de8-42f7-926a-2335879e0f6d")
	case "tax":
		moduleID = uuid.MustParse("eadaf67c-db3e-4557-82cd-8bf721c3e2e7")
	default:
		if parsed, err := uuid.Parse(notifType); err == nil {
			moduleID = parsed
		} else {
			moduleID = uuid.MustParse("5b630777-4de8-42f7-926a-2335879e0f6d")
		}
	}

	newNotif := domain.ModuleNotification{
		ID:              uuid.New(),
		ModuleID:        moduleID,
		UserID:          &userID,
		ReferenceID:     refID,
		ReferenceTitle:  title,
		ReferenceBody:   body,
		ReferenceStatus: refStatus,
		Type:            "user",
		Status:          "published",
		State:           "unread",
		IsRead:          false,
		CreatedBy:       "usecase_trigger",
		CreatedDate:     time.Now(),
		UpdatedDate:     time.Now(),
	}

	if err := db.Create(&newNotif).Error; err != nil {
		log.Printf("[Notif-Helper] ข้อผิดพลาดในการบันทึกแจ้งเตือนยูสเซอร์: %v", err)
	}

	var deviceTokens []domain.ModuleDeviceToken
	err := db.Where("user_id = ?", userID).Find(&deviceTokens).Error
	if err == nil && len(deviceTokens) > 0 {
		var tokens []string
		for _, dt := range deviceTokens {
			tokens = append(tokens, dt.Token)
		}

		payload := FCMPayload{
			Tokens:     tokens,
			Title:      title,
			Body:       body,
			RetryCount: 3,
			Data: map[string]string{
				"notification_id": newNotif.ID.String(),
				"reference_id":    refID,
				"module_type":     notifType,
			},
		}

		if GlobalFCMWorkerPool != nil {
			GlobalFCMWorkerPool.Submit(payload)
		} else {
			log.Println("[Notif-Helper] GlobalFCMWorkerPool ยังไม่ได้ Initialize")
		}
	} else {
		log.Printf("[Notif-Helper] ยูสเซอร์ %s ยังไม่ได้ผูกโทรศัพท์เพื่อรับ Push Notification (บันทึกเฉพาะ DB)", userID)
	}
}

func SendNotificationToDepartment(departmentID string, role string, title, body, refID, refStatus string) {
	db := database.DB
	if db == nil {
		log.Println("[Notif-Helper] ข้อผิดพลาด: database.DB ยังไม่ได้เชื่อมต่อ")
		return
	}

	var deptUUID *uuid.UUID
	if departmentID != "" {
		if parsed, err := uuid.Parse(departmentID); err == nil {
			deptUUID = &parsed
		}
	}

	var roleStr *string
	if role != "" {
		roleStr = &role
	}

	newNotif := domain.ModuleNotification{
		ID:              uuid.New(),
		ModuleID:        uuid.MustParse("d01b2ce5-34a9-498b-bba0-b1b8360f1ea9"),
		DepartmentID:    deptUUID,
		Role:            roleStr,
		ReferenceID:     refID,
		ReferenceTitle:  title,
		ReferenceBody:   body,
		ReferenceStatus: refStatus,
		Type:            "admin",
		Status:          "published",
		State:           "unread",
		IsRead:          false,
		CreatedBy:       "usecase_trigger",
		CreatedDate:     time.Now(),
		UpdatedDate:     time.Now(),
	}

	if err := db.Create(&newNotif).Error; err != nil {
		log.Printf("[Notif-Helper] ไม่สามารถบันทึกแจ้งเตือนกลุ่มแอดมินได้: %v", err)
	}
}

func SendNotificationToAdmin(adminID uuid.UUID, title, body, refID, refStatus, notifType string) {
	db := database.DB
	if db == nil {
		log.Println("[Notif-Helper] ข้อผิดพลาด: database.DB ยังไม่ได้เชื่อมต่อ")
		return
	}

	var moduleID uuid.UUID
	switch notifType {
	case "complaint", "complaint_assignee":
		moduleID = uuid.MustParse("5b630777-4de8-42f7-926a-2335879e0f6d")
	case "tax":
		moduleID = uuid.MustParse("eadaf67c-db3e-4557-82cd-8bf721c3e2e7")
	default:
		if parsed, err := uuid.Parse(notifType); err == nil {
			moduleID = parsed
		} else {
			moduleID = uuid.MustParse("5b630777-4de8-42f7-926a-2335879e0f6d")
		}
	}

	newNotif := domain.ModuleNotification{
		ID:              uuid.New(),
		ModuleID:        moduleID,
		UserAdminID:     &adminID,
		ReferenceID:     refID,
		ReferenceTitle:  title,
		ReferenceBody:   body,
		ReferenceStatus: refStatus,
		Type:            "admin",
		Status:          "published",
		State:           "unread",
		IsRead:          false,
		CreatedBy:       "usecase_trigger",
		CreatedDate:     time.Now(),
		UpdatedDate:     time.Now(),
	}

	if err := db.Create(&newNotif).Error; err != nil {
		log.Printf("[Notif-Helper] ข้อผิดพลาดในการบันทึกแจ้งเตือนแอดมิน: %v", err)
	}
}

