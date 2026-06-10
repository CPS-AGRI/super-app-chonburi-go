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

func (h *MunicipalityBankHandler) RegisterRoutes(router fiber.Router) {
	bank := router.Group("/municipalities/bank")

	bank.Get("/default", h.GetActiveBank)

	protected := bank.Group("", jwtutil.RequireAuth())
	protected.Get("/", h.GetAllBanks)
	protected.Post("/", h.CreateBank)
	protected.Put("/:id", h.UpdateBank)
	protected.Delete("/:id", h.DeleteBank)
}

func (h *MunicipalityBankHandler) GetActiveBank(c fiber.Ctx) error {
	bank, err := h.uc.GetActiveBank()
	if err != nil {
		return ErrorResponse(c, "Default bank not found", fiber.StatusNotFound)
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

func (h *MunicipalityBankHandler) CreateBank(c fiber.Ctx) error {
	var bank domain.MunicipalityBank
	if err := c.Bind().JSON(&bank); err != nil {
		return ErrorResponse(c, err.Error(), fiber.StatusBadRequest)
	}

	userClaims, ok := c.Locals("user").(*jwtutil.CustomClaims)
	if ok && userClaims != nil {
		bank.CreatedBy = userClaims.Name
		bank.UpdatedBy = userClaims.Name
	}

	if err := h.uc.SaveBank(&bank); err != nil {
		return ErrorResponse(c, err.Error())
	}

	return SuccessResponse(c, bank, fiber.StatusCreated)
}

func (h *MunicipalityBankHandler) UpdateBank(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return ErrorResponse(c, "Invalid UUID format", fiber.StatusBadRequest)
	}

	var bank domain.MunicipalityBank
	if err := c.Bind().JSON(&bank); err != nil {
		return ErrorResponse(c, err.Error(), fiber.StatusBadRequest)
	}

	userClaims, ok := c.Locals("user").(*jwtutil.CustomClaims)
	if ok && userClaims != nil {
		bank.UpdatedBy = userClaims.Name
	}

	bank.ID = id
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
