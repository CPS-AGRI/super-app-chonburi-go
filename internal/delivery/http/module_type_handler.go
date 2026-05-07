package http

import (
	"super-app-chonburi-go/internal/domain"
	"github.com/gofiber/fiber/v3"
)

type ModuleTypeHandler struct {
	moduleTypeUC domain.ModuleTypeUseCase
}

func NewModuleTypeHandler(uc domain.ModuleTypeUseCase) *ModuleTypeHandler {
	return &ModuleTypeHandler{moduleTypeUC: uc}
}

func (h *ModuleTypeHandler) RegisterRoutes(router fiber.Router) {
	group := router.Group("/module-types")
	group.Get("/", h.GetAllTypes)
	group.Get("/module/:moduleId", h.GetByModule)
	group.Post("/assign", h.AssignToDepartment)
	group.Post("/", h.Create)
	group.Put("/:id", h.Update)
	group.Delete("/:id", h.Delete)
}

func (h *ModuleTypeHandler) GetAllTypes(c fiber.Ctx) error {
	types, err := h.moduleTypeUC.GetAllTypes()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": types})
}

func (h *ModuleTypeHandler) GetByModule(c fiber.Ctx) error {
	moduleID := c.Params("moduleId")
	types, err := h.moduleTypeUC.GetTypesByModule(moduleID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": types})
}

func (h *ModuleTypeHandler) AssignToDepartment(c fiber.Ctx) error {
	var req struct {
		DepartmentID string   `json:"department_id"`
		ModuleID     string   `json:"module_id"`
		TypeIDs      []string `json:"type_ids"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	err := h.moduleTypeUC.AssignTypesToDepartmentModule(req.DepartmentID, req.ModuleID, req.TypeIDs)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "message": "Module types assigned successfully"})
}

func (h *ModuleTypeHandler) Create(c fiber.Ctx) error {
	var mt domain.ModuleType
	if err := c.Bind().JSON(&mt); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := h.moduleTypeUC.CreateType(&mt); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{"success": true, "data": mt})
}

func (h *ModuleTypeHandler) Update(c fiber.Ctx) error {
	id := c.Params("id")
	var mt domain.ModuleType
	if err := c.Bind().JSON(&mt); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}
	mt.ID = id

	if err := h.moduleTypeUC.UpdateType(&mt); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "data": mt})
}

func (h *ModuleTypeHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	if err := h.moduleTypeUC.DeleteType(id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"success": true, "message": "Module type deleted successfully"})
}
