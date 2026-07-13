package http

import (
	"super-app-chonburi-go/internal/domain"
	"super-app-chonburi-go/pkg/jwtutil"

	"github.com/gofiber/fiber/v3"
)

type ModuleHandler struct {
	uc domain.ModuleUseCase
}

func NewModuleHandler(uc domain.ModuleUseCase) *ModuleHandler {
	return &ModuleHandler{uc: uc}
}

func (h *ModuleHandler) RegisterRoutes(router fiber.Router) {
	group := router.Group("/modules")
	group.Get("/me", h.GetMyModules)
	group.Get("/all", h.GetAllModules)
	group.Post("/assign", h.AssignToDepartment)
	group.Post("/", h.Create)
	group.Put("/:id", h.Update)
	group.Delete("/:id", h.Delete)
}

func (h *ModuleHandler) GetMyModules(c fiber.Ctx) error {
	userClaims, ok := c.Locals("user").(*jwtutil.CustomClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	modules, err := h.uc.GetModulesForUser(userClaims.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "data": modules})
}

func (h *ModuleHandler) GetAllModules(c fiber.Ctx) error {
	modules, err := h.uc.GetAllModules()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": modules})
}

func (h *ModuleHandler) AssignToDepartment(c fiber.Ctx) error {
	var req struct {
		DepartmentId string   `json:"department_id"`
		ModuleIds    []string `json:"module_ids"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	if err := h.uc.AssignModulesToDepartment(req.DepartmentId, req.ModuleIds); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "modules assigned successfully"})
}

func (h *ModuleHandler) Create(c fiber.Ctx) error {
	var module domain.Module
	if err := c.Bind().JSON(&module); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := h.uc.CreateModule(&module); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{"success": true, "data": module})
}

func (h *ModuleHandler) Update(c fiber.Ctx) error {
	id := c.Params("id")
	var module domain.Module
	if err := c.Bind().JSON(&module); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}
	module.ID = id

	if err := h.uc.UpdateModule(&module); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "data": module})
}

func (h *ModuleHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	if err := h.uc.DeleteModule(id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Module deleted successfully"})
}
