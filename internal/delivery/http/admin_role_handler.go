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
}

func (h *AdminRoleHandler) GetAll(c fiber.Ctx) error {
	roles, err := h.useCase.GetAllRoles()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    fiber.Map{"items": roles}, // Maintain structure for frontend compat
	})
}
