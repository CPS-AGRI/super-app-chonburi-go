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
	allPermissions, err := h.useCase.GetAllPermissions()
	if err != nil {
		return ErrorResponse(c, err.Error(), fiber.StatusInternalServerError)
	}

	// Filter out sensitive permissions that shouldn't be managed via UI
	var filtered []domain.SystemPermission
	for _, p := range allPermissions {
		if p.ID != "MANAGE_CITY_SETTINGS" {
			filtered = append(filtered, p)
		}
	}

	return SuccessResponse(c, filtered)
}
