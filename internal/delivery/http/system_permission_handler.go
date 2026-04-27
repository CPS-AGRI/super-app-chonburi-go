package http

import (
	"super-app-chonburi-go/internal/domain"

	"github.com/gofiber/fiber/v3"
)

type SystemPermissionHandler struct {
	useCase domain.SystemPermissionUseCase
}

func NewSystemPermissionHandler(useCase domain.SystemPermissionUseCase) *SystemPermissionHandler {
	return &SystemPermissionHandler{useCase: useCase}
}

func (h *SystemPermissionHandler) RegisterRoutes(router fiber.Router) {
	group := router.Group("/permissions")
	group.Get("/", h.GetAllPermissions)
}

func (h *SystemPermissionHandler) GetAllPermissions(c fiber.Ctx) error {
	permissions, err := h.useCase.GetAllPermissions()
	if err != nil {
		return ErrorResponse(c, err.Error(), fiber.StatusInternalServerError)
	}
	return SuccessResponse(c, permissions)
}
