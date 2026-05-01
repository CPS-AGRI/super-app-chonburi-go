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
	group.Post("/", h.CreatePermission)
	group.Put("/:id", h.UpdatePermission)
	group.Delete("/:id", h.DeletePermission)
}

func (h *SystemPermissionHandler) CreatePermission(c fiber.Ctx) error {
	var req struct {
		Name        string `json:"name"`
		Code        string `json:"code"`
		Module      string `json:"module"`
		Description string `json:"description"`
		ParentCode  string `json:"parentCode"`
	}

	if err := c.Bind().JSON(&req); err != nil {
		return ErrorResponse(c, err.Error(), fiber.StatusBadRequest)
	}

	p := domain.SystemPermission{
		ID:          req.Code,
		NameTh:      req.Name,
		Module:      req.Module,
		Description: req.Description,
	}

	if req.ParentCode != "" {
		p.ParentID = &req.ParentCode
	}

	if err := h.useCase.CreatePermission(&p); err != nil {
		return ErrorResponse(c, err.Error(), fiber.StatusInternalServerError)
	}

	return SuccessResponse(c, p, fiber.StatusCreated)
}

func (h *SystemPermissionHandler) UpdatePermission(c fiber.Ctx) error {
	id := c.Params("id")
	var req struct {
		Name        string `json:"name"`
		Module      string `json:"module"`
		Description string `json:"description"`
		ParentCode  string `json:"parentCode"`
	}

	if err := c.Bind().JSON(&req); err != nil {
		return ErrorResponse(c, err.Error(), fiber.StatusBadRequest)
	}

	p := domain.SystemPermission{
		ID:          id,
		NameTh:      req.Name,
		Module:      req.Module,
		Description: req.Description,
	}

	if req.ParentCode != "" {
		p.ParentID = &req.ParentCode
	}

	if err := h.useCase.UpdatePermission(&p); err != nil {
		return ErrorResponse(c, err.Error(), fiber.StatusInternalServerError)
	}

	return SuccessResponse(c, p)
}

func (h *SystemPermissionHandler) DeletePermission(c fiber.Ctx) error {
	id := c.Params("id")
	if err := h.useCase.DeletePermission(id); err != nil {
		return ErrorResponse(c, err.Error(), fiber.StatusInternalServerError)
	}

	return c.SendStatus(fiber.StatusNoContent)
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
