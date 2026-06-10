package http

import (
	"super-app-chonburi-go/internal/domain"
	"super-app-chonburi-go/pkg/jwtutil"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type MunicipalityHandler struct {
	uc domain.MunicipalityUseCase
}

func NewMunicipalityHandler(uc domain.MunicipalityUseCase) *MunicipalityHandler {
	return &MunicipalityHandler{uc: uc}
}

func (h *MunicipalityHandler) RegisterRoutes(router fiber.Router) {
	router.Get("/municipality/public", h.GetPublicInfo)

	muni := router.Group("/municipalities")
	muni.Use(jwtutil.RequireAuth())

	muni.Get("/", h.GetList)
	muni.Post("/", h.Create)
	muni.Get("/:id", h.GetDetail)
	muni.Put("/:id", h.Update)
	muni.Delete("/:id", h.Delete)
}

func (h *MunicipalityHandler) GetPublicInfo(c fiber.Ctx) error {
	muni, err := h.uc.GetCurrent()
	if err != nil {
		return SuccessResponse[*domain.Municipality](c, nil)
	}
	return SuccessResponse(c, muni)
}

func (h *MunicipalityHandler) GetList(c fiber.Ctx) error {
	list, err := h.uc.GetList()
	if err != nil {
		return ErrorResponse(c, "Failed to fetch municipalities")
	}
	return SuccessResponse(c, list)
}

func (h *MunicipalityHandler) GetDetail(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return ErrorResponse(c, "Invalid UUID format", fiber.StatusBadRequest)
	}

	muni, err := h.uc.GetDetail(id)
	if err != nil {
		return ErrorResponse(c, "Municipality not found", fiber.StatusNotFound)
	}
	return SuccessResponse(c, muni)
}

func (h *MunicipalityHandler) Create(c fiber.Ctx) error {
	var muni domain.Municipality
	if err := c.Bind().JSON(&muni); err != nil {
		return ErrorResponse(c, err.Error(), fiber.StatusBadRequest)
	}

	userClaims, ok := c.Locals("user").(*jwtutil.CustomClaims)
	if ok && userClaims != nil {
		muni.CreatedBy = userClaims.Name
		muni.UpdatedBy = userClaims.Name
	}

	if err := h.uc.Create(&muni); err != nil {
		return ErrorResponse(c, err.Error())
	}

	return SuccessResponse(c, muni, fiber.StatusCreated)
}

func (h *MunicipalityHandler) Update(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return ErrorResponse(c, "Invalid UUID format", fiber.StatusBadRequest)
	}

	var muni domain.Municipality
	if err := c.Bind().JSON(&muni); err != nil {
		return ErrorResponse(c, err.Error(), fiber.StatusBadRequest)
	}

	userClaims, ok := c.Locals("user").(*jwtutil.CustomClaims)
	if ok && userClaims != nil {
		muni.UpdatedBy = userClaims.Name
	}

	muni.ID = id
	if err := h.uc.Update(&muni); err != nil {
		return ErrorResponse(c, err.Error())
	}

	return SuccessResponse(c, muni)
}

func (h *MunicipalityHandler) Delete(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return ErrorResponse(c, "Invalid UUID format", fiber.StatusBadRequest)
	}

	if err := h.uc.Delete(id); err != nil {
		return ErrorResponse(c, "Failed to delete municipality")
	}

	return c.SendStatus(fiber.StatusNoContent)
}
