package http

import (
	"super-app-chonburi-go/internal/domain"
	"super-app-chonburi-go/pkg/jwtutil"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type TaxHandler struct {
	uc domain.TaxUseCase
}

func NewTaxHandler(uc domain.TaxUseCase) *TaxHandler {
	return &TaxHandler{uc: uc}
}

func (h *TaxHandler) RegisterRoutes(router fiber.Router) {
	group := router.Group("/tax", jwtutil.RequireAuth())
	
	// Import related
	group.Post("/import", h.Import)
	
	// Query related
	// group.Get("/imports", h.GetImports)
	// group.Get("/informations", h.GetInformations)
	// group.Get("/informations/:id", h.GetInformationByID)
	
	// Action related
	group.Put("/informations/:id/status", h.UpdateStatus)
	group.Put("/informations/:id/link/:userId", h.LinkUser)
}

func (h *TaxHandler) Import(c fiber.Ctx) error {
	var req struct {
		Name         string                               `json:"name"`
		Year         string                               `json:"year"`
		ModuleTypeId string                               `json:"module_type_id"`
		Records      []domain.ModuleOnlineTaxPaymentInformation `json:"records"`
	}

	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	adminIDStr, err := GetAdminIDFromClaims(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	adminID := uuid.MustParse(adminIDStr)

	importHead := domain.ModuleOnlineTaxPayment{
		Name:           req.Name,
		Year:           req.Year,
		ModuleTypeId:   uuid.MustParse(req.ModuleTypeId),
	}

	h.uc.ImportTaxRecords(&importHead, req.Records, adminID)

	return c.JSON(fiber.Map{
		"message": "Import started in background",
		"import_id": importHead.ID,
	})
}

func (h *TaxHandler) UpdateStatus(c fiber.Ctx) error {
	id := c.Params("id")
	var req struct {
		Status     string  `json:"status"`
		ReceiptUrl *string `json:"receipt_url"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	adminIDStr, _ := GetAdminIDFromClaims(c)
	adminID := uuid.MustParse(adminIDStr)

	if err := h.uc.UpdateTaxStatus(uuid.MustParse(id), req.Status, adminID, req.ReceiptUrl); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "status updated successfully"})
}

func (h *TaxHandler) LinkUser(c fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Params("userId")

	adminIDStr, _ := GetAdminIDFromClaims(c)
	adminID := uuid.MustParse(adminIDStr)

	if err := h.uc.LinkUser(uuid.MustParse(id), uuid.MustParse(userID), adminID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "user linked successfully"})
}
