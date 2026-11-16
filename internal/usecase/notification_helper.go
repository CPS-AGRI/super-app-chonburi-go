package usecase

import (
	"log"
	"super-app-chonburi-go/internal/domain"
	"super-app-chonburi-go/pkg/database"
	"time"

	"github.com/google/uuid"
)

var (
	moduleIDComplaint         = uuid.MustParse("5b630777-4de8-42f7-926a-2335879e0f6d")
	moduleIDComplaintCenter   = uuid.MustParse("d01b2ce5-34a9-498b-bba0-b1b8360f1ea9")
	moduleIDIdentityVerify    = uuid.MustParse("8c7ce421-5d1f-41de-840b-14ac192d4778")
	moduleIDTaxPayment        = uuid.MustParse("eadaf67c-db3e-4557-82cd-8bf721c3e2e7")
	moduleIDPublicRelations   = uuid.MustParse("d04d6d51-b35d-43df-be91-31178986aae3")
	moduleIDCCTV              = uuid.MustParse("2c8e1925-5d1f-41de-840b-14ac192d4778")
	moduleIDWeather           = uuid.MustParse("10e5e192-8254-479b-ab83-86bfd09504eb")
)

func resolveNotificationModuleID(notifType string, fallback uuid.UUID) uuid.UUID {
	switch notifType {
	case "complaint", "complaint_assignee":
		return moduleIDComplaint
	case "complaint_center":
		return moduleIDComplaintCenter
	case "tax":
		return moduleIDTaxPayment
	case "verification", "user_verify":
		return moduleIDIdentityVerify
	case "pr", "news", "advertise":
		return moduleIDPublicRelations
	case "cctv":
		return moduleIDCCTV
	case "weather":
		return moduleIDWeather
	default:
		if parsed, err := uuid.Parse(notifType); err == nil {
			return parsed
		}
		return fallback
	}
}

func buildDepartmentNotification(notifType, departmentID, role, title, body, refID, refStatus string) domain.ModuleNotification {
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

	return domain.ModuleNotification{
		ID:              uuid.New(),
		ModuleID:        resolveNotificationModuleID(notifType, moduleIDComplaintCenter),
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
}

func SendNotificationToUser(userID uuid.UUID, title, body, refID, refStatus, notifType string) {
	db := database.DB
	if db == nil {
		log.Println("[Notif-Helper] ข้อผิดพลาด: database.DB ยังไม่ได้เชื่อมต่อ")
		return
	}

	moduleID := resolveNotificationModuleID(notifType, moduleIDComplaint)

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

	var tokens []string
	seen := make(map[string]bool)

	// 1. ดึงจาก user_fcm_tokens (ตารางหลักที่ Mobile App ลงทะเบียน)
	var fcmTokens []domain.UserFCMToken
	if err := db.Where("user_id = ? AND token != ''", userID).Find(&fcmTokens).Error; err == nil {
		for _, ut := range fcmTokens {
			if ut.Token != "" && !seen[ut.Token] {
				tokens = append(tokens, ut.Token)
				seen[ut.Token] = true
			}
		}
	} else {
		log.Printf("[Notif-Helper] ข้อผิดพลาดในการค้นหา user_fcm_tokens: %v", err)
	}

	// 2. Fallback / Backward-compatibility: ดึงจาก module_device_tokens เผื่อมี token เก่าค้างอยู่
	var deviceTokens []domain.ModuleDeviceToken
	if err := db.Where("user_id = ? AND token != ''", userID).Find(&deviceTokens).Error; err == nil {
		for _, dt := range deviceTokens {
			if dt.Token != "" && !seen[dt.Token] {
				tokens = append(tokens, dt.Token)
				seen[dt.Token] = true
			}
		}
	}

	if len(tokens) > 0 {
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
	SendNotificationToDepartmentWithModule("complaint_center", departmentID, role, title, body, refID, refStatus)
}

func SendNotificationToDepartmentWithModule(notifType, departmentID, role, title, body, refID, refStatus string) {
	db := database.DB
	if db == nil {
		log.Println("[Notif-Helper] ข้อผิดพลาด: database.DB ยังไม่ได้เชื่อมต่อ")
		return
	}

	newNotif := buildDepartmentNotification(notifType, departmentID, role, title, body, refID, refStatus)

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

	moduleID := resolveNotificationModuleID(notifType, moduleIDComplaint)

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

func SendBroadcastNotification(moduleID uuid.UUID, title, body, refID, refStatus, notifType, createdBy string) {
	db := database.DB
	if db == nil {
		log.Println("[Notif-Helper] ข้อผิดพลาด: database.DB ยังไม่ได้เชื่อมต่อ")
		return
	}

	if createdBy == "" {
		createdBy = "usecase_trigger"
	}

	newNotif := domain.ModuleNotification{
		ID:              uuid.New(),
		ModuleID:        moduleID,
		UserID:          nil,
		ReferenceID:     refID,
		ReferenceTitle:  title,
		ReferenceBody:   body,
		ReferenceStatus: refStatus,
		Type:            "user",
		Status:          "published",
		State:           "unread",
		IsRead:          false,
		CreatedBy:       createdBy,
		CreatedDate:     time.Now(),
		UpdatedDate:     time.Now(),
	}

	if err := db.Create(&newNotif).Error; err != nil {
		log.Printf("[Notif-Helper] ข้อผิดพลาดในการบันทึกแจ้งเตือนสาธารณะ: %v", err)
		return
	}

	var tokens []string
	seen := make(map[string]bool)

	// 1. ดึง Distinct Tokens จาก user_fcm_tokens (ตารางหลักของ Mobile App)
	var fcmTokens []string
	if err := db.Model(&domain.UserFCMToken{}).Where("token != ''").Distinct().Pluck("token", &fcmTokens).Error; err == nil {
		for _, t := range fcmTokens {
			if t != "" && !seen[t] {
				tokens = append(tokens, t)
				seen[t] = true
			}
		}
	} else {
		log.Printf("[Notif-Helper] ข้อผิดพลาดในการดึง user_fcm_tokens สำหรับ Broadcast: %v", err)
	}

	// 2. Fallback / Backward-compatibility: ดึง Distinct Tokens จาก module_device_tokens เผื่อมี Token ค้างอยู่
	var deviceTokens []string
	if err := db.Model(&domain.ModuleDeviceToken{}).Where("token != ''").Distinct().Pluck("token", &deviceTokens).Error; err == nil {
		for _, t := range deviceTokens {
			if t != "" && !seen[t] {
				tokens = append(tokens, t)
				seen[t] = true
			}
		}
	}

	if len(tokens) > 0 {
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
		log.Println("[Notif-Helper] ยังไม่มีอุปกรณ์ที่ลงทะเบียน Device Token สำหรับรับแจ้งเตือนสาธารณะ")
	}
}

