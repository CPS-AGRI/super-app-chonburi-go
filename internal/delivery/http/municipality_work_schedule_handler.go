package http

import (
	"super-app-chonburi-go/internal/domain"
	"super-app-chonburi-go/pkg/jwtutil"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type MunicipalityWorkScheduleHandler struct {
	uc domain.MunicipalityWorkScheduleUseCase
}

func NewMunicipalityWorkScheduleHandler(uc domain.MunicipalityWorkScheduleUseCase) *MunicipalityWorkScheduleHandler {
	return &MunicipalityWorkScheduleHandler{uc: uc}
}

func (h *MunicipalityWorkScheduleHandler) RegisterRoutes(router fiber.Router) {
	shifts := router.Group("/municipalities/shifts")

	protected := shifts.Group("", jwtutil.RequireAuth())
	protected.Get("/", h.GetAllShifts)
	protected.Post("/", h.CreateShift)
	protected.Put("/:id", h.UpdateShift)
	protected.Delete("/:id", h.DeleteShift)
}

func (h *MunicipalityWorkScheduleHandler) GetAllShifts(c fiber.Ctx) error {
	shifts, err := h.uc.GetAllShifts()
	if err != nil {
		return ErrorResponse(c, "Failed to fetch shifts")
	}
	return SuccessResponse(c, shifts)
}

func (h *MunicipalityWorkScheduleHandler) CreateShift(c fiber.Ctx) error {
	var schedule domain.MunicipalityWorkSchedule
	if err := c.Bind().JSON(&schedule); err != nil {
		return ErrorResponse(c, err.Error(), fiber.StatusBadRequest)
	}

	userClaims, ok := c.Locals("user").(*jwtutil.CustomClaims)
	if ok && userClaims != nil {
		schedule.CreatedBy = userClaims.Name
		schedule.UpdatedBy = userClaims.Name
	}

	if err := h.uc.SaveShift(&schedule); err != nil {
		if err == domain.ErrScheduleOverlap {
			return ErrorResponse(c, err.Error(), fiber.StatusConflict)
		}
		return ErrorResponse(c, err.Error())
	}

	return SuccessResponse(c, schedule, fiber.StatusCreated)
}

func (h *MunicipalityWorkScheduleHandler) UpdateShift(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return ErrorResponse(c, "Invalid UUID format", fiber.StatusBadRequest)
	}

	var schedule domain.MunicipalityWorkSchedule
	if err := c.Bind().JSON(&schedule); err != nil {
		return ErrorResponse(c, err.Error(), fiber.StatusBadRequest)
	}

	userClaims, ok := c.Locals("user").(*jwtutil.CustomClaims)
	if ok && userClaims != nil {
		schedule.UpdatedBy = userClaims.Name
	}

	schedule.ID = id
	if err := h.uc.SaveShift(&schedule); err != nil {
		if err == domain.ErrScheduleOverlap {
			return ErrorResponse(c, err.Error(), fiber.StatusConflict)
		}
		return ErrorResponse(c, err.Error())
	}

	return SuccessResponse(c, schedule)
}

func (h *MunicipalityWorkScheduleHandler) DeleteShift(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return ErrorResponse(c, "Invalid UUID format", fiber.StatusBadRequest)
	}

	if err := h.uc.DeleteShift(id); err != nil {
		return ErrorResponse(c, "Failed to delete shift")
	}

	return c.SendStatus(fiber.StatusNoContent)
}
