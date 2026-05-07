package http

import (
	"errors"
	"strings"
	"super-app-chonburi-go/internal/domain"
	"super-app-chonburi-go/pkg/jwtutil"
	"time"

	"github.com/gofiber/fiber/v3"
)

type ComplaintHandler struct {
	uc     domain.ComplaintUseCase
	deptUc domain.DepartmentUseCase
}

func NewComplaintHandler(uc domain.ComplaintUseCase, deptUc domain.DepartmentUseCase) *ComplaintHandler {
	return &ComplaintHandler{uc: uc, deptUc: deptUc}
}

func (h *ComplaintHandler) RegisterRoutes(router fiber.Router) {
	group := router.Group("/complaints", jwtutil.RequireAuth())
	group.Get("", h.GetComplaints)
	group.Get("/:id", h.GetComplaintByID)
	group.Post("", h.Create)
	group.Put("/:id/status", h.UpdateStatus)
	group.Post("/:id/forward", h.Forward)
	group.Put("/:id/assign", h.Assign)
	group.Put("/:id/reject", h.Reject)
	group.Post("/:id/activities", h.AddActivity)
	group.Get("/departments", h.GetDepartmentsAlias)
	group.Delete("/:id", h.Delete)
}

func getAdminIDFromClaims(c fiber.Ctx) (string, error) {
	userClaims, ok := c.Locals("user").(*jwtutil.CustomClaims)
	if !ok {
		return "", errors.New("unauthorized")
	}
	return userClaims.ID, nil
}

func (h *ComplaintHandler) GetComplaints(c fiber.Ctx) error {
	var query domain.ComplaintQuery
	if err := c.Bind().Query(&query); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid query parameters"})
	}

	statusesParam := c.Query("statuses")
	if statusesParam != "" {
		query.Status = strings.Split(statusesParam, ",")
	}

	adminID, _ := getAdminIDFromClaims(c)
	result, err := h.uc.GetComplaints(query, adminID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"data": result})
}

func (h *ComplaintHandler) GetComplaintByID(c fiber.Ctx) error {
	id := c.Params("id")
	adminID, _ := getAdminIDFromClaims(c)
	result, err := h.uc.GetComplaintByID(id, adminID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "complaint not found"})
	}

	return c.JSON(fiber.Map{"data": result})
}

func (h *ComplaintHandler) Create(c fiber.Ctx) error {
	var req domain.Complaint
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	adminID, _ := getAdminIDFromClaims(c)
	req.CreatedBy = adminID
	req.ID = domain.NewUUID()

	if err := h.uc.CreateComplaint(&req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": req})
}

func (h *ComplaintHandler) Assign(c fiber.Ctx) error {
	id := c.Params("id")
	var req struct {
		AssigneeId string `json:"assignee_id"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	adminID, _ := getAdminIDFromClaims(c)
	if err := h.uc.AssignComplaint(id, req.AssigneeId, adminID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "assigned successfully"})
}

func (h *ComplaintHandler) Reject(c fiber.Ctx) error {
	id := c.Params("id")
	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	adminID, _ := getAdminIDFromClaims(c)
	if err := h.uc.RejectComplaint(id, req.Reason, adminID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "rejected successfully"})
}

func (h *ComplaintHandler) UpdateStatus(c fiber.Ctx) error {
	id := c.Params("id")
	var req struct {
		Status      string   `json:"status"`
		Description string   `json:"description"`
		Images      []string `json:"images"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	adminID, _ := getAdminIDFromClaims(c)
	if err := h.uc.UpdateComplaintStatus(id, req.Status, req.Description, adminID, req.Images); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "status updated successfully"})
}

func (h *ComplaintHandler) Forward(c fiber.Ctx) error {
	id := c.Params("id")
	var req struct {
		DepartmentId string `json:"department_id"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	adminID, _ := getAdminIDFromClaims(c)
	if err := h.uc.ForwardComplaint(id, req.DepartmentId, adminID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "forwarded successfully"})
}

func (h *ComplaintHandler) AddActivity(c fiber.Ctx) error {
	id := c.Params("id")
	var req domain.ComplaintActivity
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	adminID, _ := getAdminIDFromClaims(c)
	req.ID = domain.NewUUID()
	req.ModuleComplaintId = id
	req.CreatedBy = adminID
	req.UpdatedBy = adminID
	req.CreatedDate = time.Now()
	req.UpdatedDate = time.Now()

	if err := h.uc.AddActivity(&req, adminID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": req})
}

func (h *ComplaintHandler) GetDepartmentsAlias(c fiber.Ctx) error {
	result, err := h.deptUc.GetAllDepartments()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"data": result})
}

func (h *ComplaintHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	if err := h.uc.DeleteComplaint(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "deleted successfully"})
}
