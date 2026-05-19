package http

import (
	"log"
	"strconv"
	"super-app-chonburi-go/internal/domain"
	"super-app-chonburi-go/pkg/jwtutil"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type PublicRelationHandler struct {
	uc domain.PublicRelationUseCase
}

func NewPublicRelationHandler(uc domain.PublicRelationUseCase) *PublicRelationHandler {
	return &PublicRelationHandler{uc: uc}
}

func (h *PublicRelationHandler) RegisterRoutes(router fiber.Router) {
	group := router.Group("/modules/:moduleId/public-relations")
	group.Use(jwtutil.RequireAuth())

	// Dashboard
	group.Get("/dashboard", h.GetDashboardStats)

	// Notifications (Registered before parameterized /:id route to avoid routing conflicts)
	group.Get("/notifications", h.GetPaginatedNotifications)
	group.Get("/notifications/histories", h.GetNotificationHistories)
	group.Get("/notifications/:id", h.GetNotificationByID)
	group.Post("/notifications", h.CreateNotification)
	group.Put("/notifications/:id", h.UpdateNotification)
	group.Delete("/notifications/:id", h.DeleteNotification)

	// News
	group.Get("/", h.GetPaginated)
	group.Get("/:id", h.GetByID)
	group.Post("/", h.Create)
	group.Put("/:id", h.Update)
	group.Delete("/:id", h.Delete)
	group.Put("/:id/comments/:commentId/hide", h.HideComment)
	group.Put("/:id/comments/:commentId/show", h.ShowComment)
}

func (h *PublicRelationHandler) RegisterGlobalRoutes(router fiber.Router) {
	group := router.Group("/welcome-screen")
	group.Use(jwtutil.RequireAuth())
	group.Get("/", h.GetWelcomeScreens)
	group.Put("/", h.UploadWelcomeScreen)
}

// Dashboard
func (h *PublicRelationHandler) GetDashboardStats(c fiber.Ctx) error {
	moduleId := c.Params("moduleId")
	stats, err := h.uc.GetDashboardStats(moduleId)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	popular, _ := h.uc.GetPopularNews(moduleId, 5)
	expiring, _ := h.uc.GetExpiringNews(moduleId, 5)

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"stats":    stats,
			"popular":  popular,
			"expiring": expiring,
		},
	})
}

// News
func (h *PublicRelationHandler) GetPaginated(c fiber.Ctx) error {
	moduleId := c.Params("moduleId")
	page, _ := strconv.Atoi(c.Query("page_number", "1"))
	size, _ := strconv.Atoi(c.Query("page_size", "10"))
	title := c.Query("title")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	var titlePtr *string
	if title != "" {
		titlePtr = &title
	}
	var startPtr *string
	if startDate != "" {
		startPtr = &startDate
	}
	var endPtr *string
	if endDate != "" {
		endPtr = &endDate
	}

	query := domain.PublicRelationQuery{
		PageNumber: page,
		PageSize:   size,
		Title:      titlePtr,
		StartDate:  startPtr,
		EndDate:    endPtr,
	}

	res, err := h.uc.GetPaginated(moduleId, query)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "data": res})
}

func (h *PublicRelationHandler) GetByID(c fiber.Ctx) error {
	moduleId := c.Params("moduleId")
	id := c.Params("id")
	log.Printf("[DEBUG] PublicRelation GetByID called. moduleId: %s, id: %s", moduleId, id)

	res, err := h.uc.GetByID(moduleId, id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if res == nil {
		return c.Status(404).JSON(fiber.Map{"error": "public relation not found"})
	}

	return c.JSON(fiber.Map{"success": true, "data": res})
}

func (h *PublicRelationHandler) Create(c fiber.Ctx) error {
	moduleId := c.Params("moduleId")
	userClaims, ok := c.Locals("user").(*jwtutil.CustomClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	var pr domain.PublicRelation
	if err := c.Bind().JSON(&pr); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}
	pr.ModuleId = uuid.MustParse(moduleId)

	if err := h.uc.Create(&pr, userClaims.ID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{"success": true, "data": pr})
}

func (h *PublicRelationHandler) Update(c fiber.Ctx) error {
	moduleId := c.Params("moduleId")
	id := c.Params("id")
	userClaims, ok := c.Locals("user").(*jwtutil.CustomClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	var pr domain.PublicRelation
	if err := c.Bind().JSON(&pr); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}
	pr.ID = uuid.MustParse(id)
	pr.ModuleId = uuid.MustParse(moduleId)

	if err := h.uc.Update(&pr, userClaims.ID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "data": pr})
}

func (h *PublicRelationHandler) Delete(c fiber.Ctx) error {
	moduleId := c.Params("moduleId")
	id := c.Params("id")
	userClaims, ok := c.Locals("user").(*jwtutil.CustomClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	if err := h.uc.Delete(moduleId, id, userClaims.ID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "message": "deleted successfully"})
}

func (h *PublicRelationHandler) HideComment(c fiber.Ctx) error {
	moduleId := c.Params("moduleId")
	prId := c.Params("id")
	commentId := c.Params("commentId")
	userClaims, ok := c.Locals("user").(*jwtutil.CustomClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	if err := h.uc.HideComment(moduleId, prId, commentId, userClaims.ID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "message": "comment hidden successfully"})
}

func (h *PublicRelationHandler) ShowComment(c fiber.Ctx) error {
	moduleId := c.Params("moduleId")
	prId := c.Params("id")
	commentId := c.Params("commentId")
	userClaims, ok := c.Locals("user").(*jwtutil.CustomClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	if err := h.uc.ShowComment(moduleId, prId, commentId, userClaims.ID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "message": "comment shown successfully"})
}

// Notifications
func (h *PublicRelationHandler) GetPaginatedNotifications(c fiber.Ctx) error {
	moduleId := c.Params("moduleId")
	log.Printf("[DEBUG] PublicRelation GetPaginatedNotifications called. moduleId: %s", moduleId)
	page, _ := strconv.Atoi(c.Query("page_number", "1"))
	size, _ := strconv.Atoi(c.Query("page_size", "10"))
	title := c.Query("title")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	var titlePtr *string
	if title != "" {
		titlePtr = &title
	}
	var startPtr *string
	if startDate != "" {
		startPtr = &startDate
	}
	var endPtr *string
	if endDate != "" {
		endPtr = &endDate
	}

	query := domain.PublicRelationNotificationQuery{
		PageNumber: page,
		PageSize:   size,
		Title:      titlePtr,
		StartDate:  startPtr,
		EndDate:    endPtr,
	}

	res, err := h.uc.GetPaginatedNotifications(moduleId, query, false)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "data": res})
}

func (h *PublicRelationHandler) GetNotificationHistories(c fiber.Ctx) error {
	moduleId := c.Params("moduleId")
	page, _ := strconv.Atoi(c.Query("page_number", "1"))
	size, _ := strconv.Atoi(c.Query("page_size", "10"))
	title := c.Query("title")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	var titlePtr *string
	if title != "" {
		titlePtr = &title
	}
	var startPtr *string
	if startDate != "" {
		startPtr = &startDate
	}
	var endPtr *string
	if endDate != "" {
		endPtr = &endDate
	}

	query := domain.PublicRelationNotificationQuery{
		PageNumber: page,
		PageSize:   size,
		Title:      titlePtr,
		StartDate:  startPtr,
		EndDate:    endPtr,
	}

	res, err := h.uc.GetPaginatedNotifications(moduleId, query, true)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "data": res})
}

func (h *PublicRelationHandler) GetNotificationByID(c fiber.Ctx) error {
	moduleId := c.Params("moduleId")
	id := c.Params("id")

	res, err := h.uc.GetNotificationByID(moduleId, id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if res == nil {
		return c.Status(404).JSON(fiber.Map{"error": "notification not found"})
	}

	return c.JSON(fiber.Map{"success": true, "data": res})
}

func (h *PublicRelationHandler) CreateNotification(c fiber.Ctx) error {
	moduleId := c.Params("moduleId")
	userClaims, ok := c.Locals("user").(*jwtutil.CustomClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	var n domain.PublicRelationNotification
	if err := c.Bind().JSON(&n); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}
	n.ModuleId = uuid.MustParse(moduleId)

	if err := h.uc.CreateNotification(&n, userClaims.ID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{"success": true, "data": n})
}

func (h *PublicRelationHandler) UpdateNotification(c fiber.Ctx) error {
	moduleId := c.Params("moduleId")
	id := c.Params("id")
	userClaims, ok := c.Locals("user").(*jwtutil.CustomClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	var n domain.PublicRelationNotification
	if err := c.Bind().JSON(&n); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}
	n.ID = uuid.MustParse(id)
	n.ModuleId = uuid.MustParse(moduleId)

	if err := h.uc.UpdateNotification(&n, userClaims.ID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "data": n})
}

func (h *PublicRelationHandler) DeleteNotification(c fiber.Ctx) error {
	moduleId := c.Params("moduleId")
	id := c.Params("id")
	userClaims, ok := c.Locals("user").(*jwtutil.CustomClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	if err := h.uc.DeleteNotification(moduleId, id, userClaims.ID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "message": "deleted successfully"})
}

// Welcome Screen
func (h *PublicRelationHandler) GetWelcomeScreens(c fiber.Ctx) error {
	screens, err := h.uc.GetWelcomeScreens()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": screens})
}

func (h *PublicRelationHandler) UploadWelcomeScreen(c fiber.Ctx) error {
	userClaims, ok := c.Locals("user").(*jwtutil.CustomClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	var screen domain.MunicipalityWelcomeScreen
	if err := c.Bind().JSON(&screen); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := h.uc.UploadWelcomeScreen(&screen, userClaims.ID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "data": screen})
}
