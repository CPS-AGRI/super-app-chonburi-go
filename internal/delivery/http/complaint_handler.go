package http

import (
	"strconv"
	"strings"
	"super-app-chonburi-go/internal/domain"
	"super-app-chonburi-go/pkg/jwtutil"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type ComplaintHandler struct {
	usecase domain.ComplaintUseCase
}

func NewComplaintHandler(usecase domain.ComplaintUseCase) *ComplaintHandler {
	return &ComplaintHandler{usecase: usecase}
}

func (h *ComplaintHandler) RegisterRoutes(router fiber.Router) {
	group := router.Group("/complaints")
	group.Get("", h.GetComplaints)
	group.Get("/:id", h.GetComplaintByID)
	group.Post("", h.CreateComplaint) // Called by Citizen App
	group.Put("/:id/assign", h.AssignComplaint)
	group.Put("/:id/reject", h.RejectComplaint)
	group.Post("/:id/activities", h.AddActivity)
	group.Delete("/:id", h.DeleteComplaint)
}

func getAdminIDFromClaims(c fiber.Ctx) (uuid.UUID, error) {
	userClaims, ok := c.Locals("user").(*jwtutil.CustomClaims)
	if !ok {
		return uuid.Nil, fiber.ErrUnauthorized
	}
	return uuid.Parse(userClaims.ID)
}

func (h *ComplaintHandler) GetComplaints(c fiber.Ctx) error {
	adminID, err := getAdminIDFromClaims(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	pageNumber, _ := strconv.Atoi(c.Query("page_number", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "50"))
	
	var statusList []string
	if statuses := c.Query("statuses"); statuses != "" {
		statusList = strings.Split(statuses, ",")
	}

	query := domain.ComplaintQuery{
		PageNumber: pageNumber,
		PageSize:   pageSize,
		Status:     statusList,
	}

	if assignee := c.Query("assignee_id"); assignee != "" {
		if id, err := uuid.Parse(assignee); err == nil {
			query.AssigneeID = &id
		}
	}

	response, err := h.usecase.GetComplaints(query, adminID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return SuccessResponse(c, response)
}

func (h *ComplaintHandler) GetComplaintByID(c fiber.Ctx) error {
	adminID, err := getAdminIDFromClaims(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID format"})
	}

	complaint, err := h.usecase.GetComplaintByID(id, adminID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return SuccessResponse(c, complaint)
}

// CreateComplaint simulates a citizen submitting a complaint
func (h *ComplaintHandler) CreateComplaint(c fiber.Ctx) error {
	var req domain.Complaint
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	// Generate a mock DocumentID if missing
	if req.DocumentID == "" {
		req.DocumentID = "CMP-AUTO-" + uuid.New().String()[:8]
	}

	if err := h.usecase.CreateComplaint(&req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    req,
	})
}

func (h *ComplaintHandler) AssignComplaint(c fiber.Ctx) error {
	adminID, err := getAdminIDFromClaims(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID format"})
	}

	var req struct {
		AssigneeID string `json:"assigneeId"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	assigneeID, err := uuid.Parse(req.AssigneeID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid assignee ID format"})
	}

	if err := h.usecase.AssignComplaint(id, adminID, assigneeID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

func (h *ComplaintHandler) RejectComplaint(c fiber.Ctx) error {
	adminID, err := getAdminIDFromClaims(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID format"})
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	if err := h.usecase.RejectComplaint(id, adminID, req.Reason); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

func (h *ComplaintHandler) AddActivity(c fiber.Ctx) error {
	adminID, err := getAdminIDFromClaims(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID format"})
	}

	var activity domain.ComplaintActivity
	if err := c.Bind().Body(&activity); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	activity.ComplaintID = id

	if err := h.usecase.AddActivity(&activity, adminID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    activity,
	})
}

func (h *ComplaintHandler) DeleteComplaint(c fiber.Ctx) error {
	adminID, err := getAdminIDFromClaims(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID format"})
	}

	if err := h.usecase.DeleteComplaint(id, adminID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return SuccessResponse(c, fiber.Map{"message": "Complaint deleted successfully"})
}
