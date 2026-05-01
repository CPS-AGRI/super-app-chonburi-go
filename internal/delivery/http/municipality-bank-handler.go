package http

import (
	"super-app-chonburi-go/internal/domain"
	"super-app-chonburi-go/pkg/jwtutil"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type MunicipalityBankHandler struct {
	uc domain.MunicipalityBankUseCase
}

func NewMunicipalityBankHandler(uc domain.MunicipalityBankUseCase) *MunicipalityBankHandler {
	return &MunicipalityBankHandler{uc: uc}
}

func (h *MunicipalityBankHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1/municipalities")
	api.Use(jwtutil.RequireAuth())

	bank := api.Group("/bank")
	bank.Get("/default", h.GetActiveBank)
	bank.Get("/all", h.GetAllBanks)
	bank.Post("/", h.SaveBank)
	bank.Put("/:id", h.SaveBank)
	bank.Delete("/:id", h.DeleteBank)
}

func (h *MunicipalityBankHandler) GetActiveBank(c fiber.Ctx) error {
	bank, err := h.uc.GetActiveBank()
	if err != nil {
		return SuccessResponse(c, nil) // Return nil if no active bank
	}
	return SuccessResponse(c, bank)
}

func (h *MunicipalityBankHandler) GetAllBanks(c fiber.Ctx) error {
	banks, err := h.uc.GetAllBanks()
	if err != nil {
		return ErrorResponse(c, "Failed to fetch banks")
	}
	return SuccessResponse(c, banks)
}

func (h *MunicipalityBankHandler) SaveBank(c fiber.Ctx) error {
	var bank domain.MunicipalityBank
	if err := c.Bind().JSON(&bank); err != nil {
		return ErrorResponse(c, err.Error(), fiber.StatusBadRequest)
	}

	idParam := c.Params("id")
	if idParam != "" {
		id, err := uuid.Parse(idParam)
		if err == nil {
			bank.ID = id
		}
	}

	userClaims, ok := c.Locals("user").(*jwtutil.CustomClaims)
	if ok && userClaims != nil {
		if bank.ID == uuid.Nil {
			bank.CreatedBy = userClaims.Name
		}
		bank.UpdatedBy = userClaims.Name
	}

	if err := h.uc.SaveBank(&bank); err != nil {
		return ErrorResponse(c, err.Error())
	}

	return SuccessResponse(c, bank)
}

func (h *MunicipalityBankHandler) DeleteBank(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return ErrorResponse(c, "Invalid UUID format", fiber.StatusBadRequest)
	}

	if err := h.uc.DeleteBank(id); err != nil {
		return ErrorResponse(c, "Failed to delete bank")
	}

	return c.SendStatus(fiber.StatusNoContent)
}
