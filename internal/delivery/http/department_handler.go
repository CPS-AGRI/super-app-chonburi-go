package http

import (
	"super-app-chonburi-go/internal/domain"
	"super-app-chonburi-go/pkg/jwtutil"

	"github.com/gofiber/fiber/v3"
)

type DepartmentHandler struct {
	uc domain.DepartmentUseCase
}

func NewDepartmentHandler(uc domain.DepartmentUseCase) *DepartmentHandler {
	return &DepartmentHandler{uc: uc}
}

func (h *DepartmentHandler) RegisterRoutes(router fiber.Router) {
	group := router.Group("/departments", jwtutil.RequireAuth())
	group.Get("", h.GetDepartments)
	group.Get("/all", h.GetDepartmentsAll)
	group.Get("/list", h.GetDepartmentsAll)
	group.Get("/:id", h.GetByID)
	group.Post("", h.Create)
	group.Put("/:id", h.Update)
	group.Delete("/:id", h.Delete)
}

func (h *DepartmentHandler) GetDepartments(c fiber.Ctx) error {
	var query domain.DepartmentQuery
	if err := c.Bind().Query(&query); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid query parameters"})
	}

	result, err := h.uc.GetDepartments(query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "data": result})
}

func (h *DepartmentHandler) GetDepartmentsAll(c fiber.Ctx) error {
	result, err := h.uc.GetAllDepartments()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": result})
}

func (h *DepartmentHandler) GetByID(c fiber.Ctx) error {
	id := c.Params("id")
	result, err := h.uc.GetDepartmentByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "department not found"})
	}
	return c.JSON(fiber.Map{"success": true, "data": result})
}

func (h *DepartmentHandler) Create(c fiber.Ctx) error {
	var req domain.Department
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	req.ID = domain.NewUUID()
	if err := h.uc.CreateDepartment(&req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"success": true, "data": req})
}

func (h *DepartmentHandler) Update(c fiber.Ctx) error {
	id := c.Params("id")
	var req domain.Department
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	req.ID = id
	if err := h.uc.UpdateDepartment(&req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "data": req})
}

func (h *DepartmentHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	if err := h.uc.DeleteDepartment(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "message": "deleted successfully"})
}
