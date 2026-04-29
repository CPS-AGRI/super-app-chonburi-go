package http

import (
	"strconv"

	"github.com/google/uuid"
	"github.com/gofiber/fiber/v3"
	"super-app-chonburi-go/internal/domain"
	"super-app-chonburi-go/pkg/jwtutil"
)

type DepartmentHandler struct {
	useCase domain.DepartmentUseCase
}

func NewDepartmentHandler(useCase domain.DepartmentUseCase) *DepartmentHandler {
	return &DepartmentHandler{useCase}
}

func (h *DepartmentHandler) RegisterRoutes(router fiber.Router) {
	group := router.Group("/departments")
	group.Get("/", h.GetAll)
	group.Get("/list", h.GetList)
	group.Get("/:id", h.GetByID)
	group.Post("/", h.Create)
	group.Put("/:id", h.Update)
	group.Delete("/:id", h.Delete)
}

func (h *DepartmentHandler) GetAll(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	size, _ := strconv.Atoi(c.Query("size", "10"))
	name := c.Query("name")
	status := c.Query("status")

	res, err := h.useCase.GetDepartments(domain.DepartmentQuery{
		PageNumber: page,
		PageSize:   size,
		Name:       name,
		Status:     status,
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    res,
	})
}

func (h *DepartmentHandler) GetList(c fiber.Ctx) error {
	depts, err := h.useCase.GetAllDepartments()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": depts})
}

func (h *DepartmentHandler) GetByID(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}

	dept, err := h.useCase.GetDepartmentByID(id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if dept == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Department not found"})
	}

	return c.JSON(fiber.Map{"success": true, "data": dept})
}

func (h *DepartmentHandler) Create(c fiber.Ctx) error {
	var dept domain.Department
	if err := c.Bind().Body(&dept); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// Set Audit fields
	if user, ok := c.Locals("user").(*jwtutil.CustomClaims); ok {
		dept.CreatedBy = user.Name
		dept.UpdatedBy = user.Name
	}

	if err := h.useCase.CreateDepartment(&dept); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{"success": true, "data": dept})
}

func (h *DepartmentHandler) Update(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}

	var dept domain.Department
	if err := c.Bind().Body(&dept); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	dept.ID = id

	// Set Audit fields
	if user, ok := c.Locals("user").(*jwtutil.CustomClaims); ok {
		dept.UpdatedBy = user.Name
	}

	if err := h.useCase.UpdateDepartment(&dept); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "data": dept})
}

func (h *DepartmentHandler) Delete(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}

	if err := h.useCase.DeleteDepartment(id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Deleted successfully"})
}
