package http

import (
	"super-app-chonburi-go/internal/domain"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type AdminHandler struct {
	adminUseCase domain.AdminUseCase
}

func NewAdminHandler(adminUseCase domain.AdminUseCase) *AdminHandler {
	return &AdminHandler{adminUseCase: adminUseCase}
}

func (h *AdminHandler) RegisterRoutes(router fiber.Router) {
	group := router.Group("/admins")
	group.Get("", h.GetAdmins)
	group.Get("/:id", h.GetAdminByID)
	group.Post("", h.CreateAdmin)
	group.Put("/:id", h.UpdateAdmin)
	group.Delete("/:id", h.DeleteAdmin)
}

func (h *AdminHandler) GetAdmins(c fiber.Ctx) error {
	pageNumber, _ := strconv.Atoi(c.Query("page_number", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))

	query := domain.AdminQuery{
		PageNumber:   pageNumber,
		PageSize:     pageSize,
		Email:        c.Query("email"),
		Name:         c.Query("name"),
		DepartmentID: c.Query("department_id"),
	}

	response, err := h.adminUseCase.GetAdmins(query)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return SuccessResponse(c, response)
}

func (h *AdminHandler) GetAdminByID(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID format"})
	}

	admin, err := h.adminUseCase.GetAdminByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return SuccessResponse(c, admin)
}

func (h *AdminHandler) CreateAdmin(c fiber.Ctx) error {
	var admin domain.Admin
	if err := c.Bind().JSON(&admin); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "JSON Bind Error: " + err.Error()})
	}

	if err := h.adminUseCase.CreateAdmin(&admin); err != nil {
		if err.Error() == "email already exists" {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return SuccessResponse(c, admin, fiber.StatusCreated)
}

func (h *AdminHandler) UpdateAdmin(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID format"})
	}

	var admin domain.Admin
	if err := c.Bind().JSON(&admin); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}
	admin.ID = id

	if err := h.adminUseCase.UpdateAdmin(&admin); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return SuccessResponse(c, admin, fiber.StatusOK)
}

func (h *AdminHandler) DeleteAdmin(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID format"})
	}

	if err := h.adminUseCase.DeleteAdmin(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return SuccessResponse(c, nil, fiber.StatusOK)
}
