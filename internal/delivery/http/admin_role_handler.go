package http

import (
	"github.com/gofiber/fiber/v3"
	"super-app-chonburi-go/internal/domain"
)

type AdminRoleHandler struct {
	useCase domain.AdminRoleUseCase
}

func NewAdminRoleHandler(useCase domain.AdminRoleUseCase) *AdminRoleHandler {
	return &AdminRoleHandler{useCase: useCase}
}

func (h *AdminRoleHandler) RegisterRoutes(router fiber.Router) {
	group := router.Group("/roles")
	group.Get("/", h.GetAll)
	group.Post("/", h.Create)
	group.Put("/:id", h.Update)
	group.Delete("/:id", h.Delete)
}

func (h *AdminRoleHandler) GetAll(c fiber.Ctx) error {
	roles, err := h.useCase.GetAllRoles()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"items": roles},
	})
}

func (h *AdminRoleHandler) Create(c fiber.Ctx) error {
	var role domain.AdminRole
	if err := c.Bind().JSON(&role); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := h.useCase.CreateRole(&role); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{"success": true, "data": role})
}

func (h *AdminRoleHandler) Update(c fiber.Ctx) error {
	id := c.Params("id")
	var role domain.AdminRole
	if err := c.Bind().JSON(&role); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}
	role.ID = id

	if err := h.useCase.UpdateRole(&role); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "data": role})
}

func (h *AdminRoleHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	if err := h.useCase.DeleteRole(id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Role deleted successfully"})
}
