package http

import (
	"super-app-chonburi-go/internal/domain"
	"super-app-chonburi-go/internal/usecase"
	"super-app-chonburi-go/pkg/jwtutil"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NotificationHandler struct {
	db         *gorm.DB
	workerPool *usecase.FCMWorkerPool
}

func NewNotificationHandler(db *gorm.DB, workerPool *usecase.FCMWorkerPool) *NotificationHandler {
	return &NotificationHandler{
		db:         db,
		workerPool: workerPool,
	}
}

func (h *NotificationHandler) RegisterRoutes(router fiber.Router) {
	group := router.Group("/notifications", jwtutil.RequireAuth())

	group.Post("/token", h.RegisterToken)

	group.Get("", h.GetNotifications)

	group.Put("/:id/read", h.MarkAsRead)

	group.Post("/send-test", h.SendTestNotification)
}

func (h *NotificationHandler) RegisterToken(c fiber.Ctx) error {
	claims, ok := c.Locals("user").(*jwtutil.CustomClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "ไม่พบการล็อกอินของระบบ"})
	}

	userUUID, err := uuid.Parse(claims.ID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID ผู้ใช้งานไม่ถูกต้อง"})
	}

	var req struct {
		Token      string `json:"token"`
		DeviceType string `json:"device_type"`
	}

	if err := c.Bind().JSON(&req); err != nil || req.Token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "กรุณาระบุ Token และรูปแบบอุปกรณ์ให้ถูกต้อง"})
	}

	if req.DeviceType == "" {
		req.DeviceType = "android"
	}

	var deviceToken domain.ModuleDeviceToken

	err = h.db.Where("token = ?", req.Token).First(&deviceToken).Error
	if err == nil {

		deviceToken.UserID = userUUID
		deviceToken.DeviceType = req.DeviceType
		deviceToken.UpdatedDate = time.Now()
		if err := h.db.Save(&deviceToken).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "ไม่สามารถบันทึก Token ได้"})
		}
	} else {

		newDeviceToken := domain.ModuleDeviceToken{
			ID:          uuid.New(),
			UserID:      userUUID,
			Token:       req.Token,
			DeviceType:  req.DeviceType,
			CreatedDate: time.Now(),
			UpdatedDate: time.Now(),
		}
		if err := h.db.Create(&newDeviceToken).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "ไม่สามารถลงทะเบียน Token ใหม่ได้"})
		}
	}

	return SuccessResponse(c, fiber.Map{
		"status":  "success",
		"message": "บันทึกและจับคู่ Device Token สำเร็จ",
	})
}

func (h *NotificationHandler) GetNotifications(c fiber.Ctx) error {
	claims, ok := c.Locals("user").(*jwtutil.CustomClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "ไม่พบสิทธิ์การเข้าใช้งาน"})
	}

	userUUID, err := uuid.Parse(claims.ID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ข้อมูลผู้ใช้งานไม่ถูกต้อง"})
	}

	unreadOnly := c.Query("unread_only", "")

	var notifications []domain.ModuleNotification

	var adminCount int64
	h.db.Model(&domain.Admin{}).Where("id = ?", userUUID).Count(&adminCount)

	query := h.db.Model(&domain.ModuleNotification{})

	if adminCount > 0 {
		// Super admin ไม่ต้องมีแจ้งเตือน
		if claims.Role == "SuperAdmin" {
			return SuccessResponse(c, []domain.ModuleNotification{})
		}

		var admin domain.Admin
		if err := h.db.Preload("Departments").Where("id = ?", userUUID).First(&admin).Error; err == nil {
			var deptIDs []uuid.UUID
			for _, dept := range admin.Departments {
				if parsedDeptID, err := uuid.Parse(dept.ID); err == nil {
					deptIDs = append(deptIDs, parsedDeptID)
				}
			}

			if claims.Role == "Managers" {
				// หัวหน้างาน (Managers): เห็นเรื่องของตนเอง และ เรื่องแผนกที่เป็นระดับผู้บริหาร
				if len(deptIDs) > 0 {
					query = query.Where("type = 'admin' AND (user_admin_id = ? OR (department_id IN (?) AND role = 'Managers') OR (department_id IS NULL AND role = 'Managers'))",
						userUUID, deptIDs)
				} else {
					query = query.Where("type = 'admin' AND (user_admin_id = ? OR role = 'Managers')",
						userUUID)
				}
			} else if claims.Role == "Employees" {
				// พนักงานปฏิบัติการ (Employees): เห็นเฉพาะที่รับมอบหมายตรงๆ หรือแบบภาษีในระดับปฏิบัติการ
				if len(deptIDs) > 0 {
					query = query.Where("type = 'admin' AND (user_admin_id = ? OR (department_id IN (?) AND role = 'Employees') OR (department_id IS NULL AND role = 'Employees'))",
						userUUID, deptIDs)
				} else {
					query = query.Where("type = 'admin' AND (user_admin_id = ? OR role = 'Employees')",
						userUUID)
				}
			} else {
				query = query.Where("type = 'admin' AND user_admin_id = ?", userUUID)
			}
		} else {
			query = query.Where("type = 'admin' AND user_admin_id = ?", userUUID)
		}
	} else {

		query = query.Where("user_id = ? AND type = 'user'", userUUID)
	}

	if unreadOnly == "true" {

		query = query.Where("is_read = ?", false)
	}

	err = query.Order("created_date DESC").Limit(50).Find(&notifications).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "ไม่สามารถดึงข้อมูลรายการแจ้งเตือนได้"})
	}

	return SuccessResponse(c, notifications)
}

func (h *NotificationHandler) MarkAsRead(c fiber.Ctx) error {
	_, ok := c.Locals("user").(*jwtutil.CustomClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "ไม่พบการล็อกอินของระบบ"})
	}

	id := c.Params("id")
	notifUUID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID แจ้งเตือนไม่ถูกต้อง"})
	}

	var notif domain.ModuleNotification
	if err := h.db.Where("id = ?", notifUUID).First(&notif).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "ไม่พบข้อมูลแจ้งเตือน"})
	}

	notif.IsRead = true
	notif.State = "read"
	notif.UpdatedDate = time.Now()

	if err := h.db.Save(&notif).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "ไม่สามารถเปลี่ยนสถานะแจ้งเตือนเป็นอ่านแล้วได้"})
	}

	return SuccessResponse(c, fiber.Map{
		"status":  "success",
		"message": "เปลี่ยนสถานะแจ้งเตือนเป็นอ่านแล้วสำเร็จ",
	})
}

func (h *NotificationHandler) SendTestNotification(c fiber.Ctx) error {
	var req struct {
		TargetUserID string `json:"target_user_id"`
		Title        string `json:"title"`
		Body         string `json:"body"`
	}

	if err := c.Bind().JSON(&req); err != nil || req.TargetUserID == "" || req.Title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ข้อมูลการส่งไม่ครบถ้วน"})
	}

	targetUUID, err := uuid.Parse(req.TargetUserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID เป้าหมายผู้รับไม่ถูกต้อง"})
	}

	var deviceTokens []domain.ModuleDeviceToken
	err = h.db.Where("user_id = ?", targetUUID).Find(&deviceTokens).Error
	if err != nil || len(deviceTokens) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "ไม่พบเครื่องโทรศัพท์หรือ Token ของผู้รับคนนี้ในระบบ สำหรับส่ง Push Notification",
		})
	}

	var tokens []string
	for _, dt := range deviceTokens {
		tokens = append(tokens, dt.Token)
	}

	moduleID := uuid.New()
	newNotif := domain.ModuleNotification{
		ID:              uuid.New(),
		ModuleID:        moduleID,
		UserID:          &targetUUID,
		ReferenceID:     moduleID.String(),
		ReferenceTitle:  req.Title,
		ReferenceBody:   req.Body,
		ReferenceStatus: "sent",
		Type:            "user",
		Status:          "published",
		State:           "unread",
		IsRead:          false,
		CreatedBy:       "system_admin",
		CreatedDate:     time.Now(),
		UpdatedDate:     time.Now(),
	}

	if err := h.db.Create(&newNotif).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "ไม่สามารถบันทึกข้อมูลแจ้งเตือนเข้าระบบได้"})
	}

	payload := usecase.FCMPayload{
		Tokens: tokens,
		Title:  req.Title,
		Body:   req.Body,
		Data: map[string]string{
			"notification_id": newNotif.ID.String(),
			"module_type":     "system_test",
		},
	}

	h.workerPool.Submit(payload)

	return SuccessResponse(c, fiber.Map{
		"status":            "success",
		"message":           "ยัดแจ้งเตือนเข้าคิวของ Worker Pool เรียบร้อยแล้ว (กำลังยิงส่ง Asynchronously)",
		"notification_id":   newNotif.ID,
		"registered_tokens": len(tokens),
	})
}
