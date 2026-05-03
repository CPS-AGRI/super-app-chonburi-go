package http

import (
	"super-app-chonburi-go/internal/domain"
	"super-app-chonburi-go/pkg/jwtutil"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type CityShiftHandler struct {
	uc domain.CityShiftUseCase
}

func NewCityShiftHandler(uc domain.CityShiftUseCase) *CityShiftHandler {
	return &CityShiftHandler{uc: uc}
}

func (h *CityShiftHandler) RegisterRoutes(app *fiber.App) {
	group := app.Group("/api/v1/municipalities/shifts")
	group.Use(jwtutil.RequireAuth())
	group.Get("/", h.GetAllShifts)
	group.Post("/", h.SaveShift)
	group.Put("/:id", h.SaveShift)
	group.Delete("/:id", h.DeleteShift)
}

func (h *CityShiftHandler) GetAllShifts(c fiber.Ctx) error {
	shifts, err := h.uc.GetAllShifts()
	if err != nil {
		return ErrorResponse(c, "Failed to retrieve shifts", fiber.StatusInternalServerError)
	}

	return SuccessResponse(c, shifts)
}

func (h *CityShiftHandler) SaveShift(c fiber.Ctx) error {
	var shift domain.CityShift
	if err := c.Bind().JSON(&shift); err != nil {
		return ErrorResponse(c, "Invalid input format", fiber.StatusBadRequest)
	}

	// If PUT request, get ID from params
	if c.Method() == fiber.MethodPut {
		idParam := c.Params("id")
		id, err := uuid.Parse(idParam)
		if err != nil {
			return ErrorResponse(c, "Invalid ID format", fiber.StatusBadRequest)
		}
		shift.ID = id
	}

	// Basic validation
	if shift.OfficerName == "" || shift.WorkingDay == "" || shift.WorkingHoursStart == "" || shift.WorkingHoursEnd == "" {
		return ErrorResponse(c, "Missing required fields", fiber.StatusBadRequest)
	}

	// Set CreatedBy and UpdatedBy from Token
	userClaims, ok := c.Locals("user").(*jwtutil.CustomClaims)
	if ok && userClaims != nil {
		if shift.ID == uuid.Nil {
			shift.CreatedBy = userClaims.Name
		}
		shift.UpdatedBy = userClaims.Name
	}

	err := h.uc.SaveShift(&shift)
	if err != nil {
		// Differentiate overlap error (client error) from server errors
		if err.Error() == "ไม่สามารถเพิ่มได้เนื่องจากเวลาคาบเกี่ยวกัน" {
			return ErrorResponse(c, err.Error(), fiber.StatusConflict) // 409 Conflict
		}
		return ErrorResponse(c, "Failed to save shift: "+err.Error(), fiber.StatusInternalServerError)
	}

	return SuccessResponse(c, shift)
}

func (h *CityShiftHandler) DeleteShift(c fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return ErrorResponse(c, "Invalid ID format", fiber.StatusBadRequest)
	}

	err = h.uc.DeleteShift(id)
	if err != nil {
		return ErrorResponse(c, "Failed to delete shift", fiber.StatusInternalServerError)
	}

	return SuccessResponse(c, nil)
}
