package http

import (
	"super-app-chonburi-go/internal/domain"
	"super-app-chonburi-go/internal/usecase"
	"super-app-chonburi-go/pkg/jwtutil"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type WeatherAccountHandler struct {
	useCase usecase.WeatherAccountUseCase
}

func NewWeatherAccountHandler(useCase usecase.WeatherAccountUseCase) *WeatherAccountHandler {
	return &WeatherAccountHandler{useCase: useCase}
}

func (h *WeatherAccountHandler) RegisterRoutes(router fiber.Router) {
	group := router.Group("/weather/accounts", jwtutil.RequireAuth())
	group.Get("/", h.GetAll)
	group.Get("/:id", h.GetByID)
	group.Post("/", h.Create)
	group.Put("/:id", h.Update)
	group.Patch("/:id", h.Update)
	group.Delete("/:id", h.Delete)
}

func (h *WeatherAccountHandler) GetAll(c fiber.Ctx) error {
	ctx := c.Context()
	accounts, err := h.useCase.GetAll(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"success": true,
		"data":    accounts,
		"total":   len(accounts),
	})
}

func (h *WeatherAccountHandler) GetByID(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "invalid account id",
		})
	}

	ctx := c.Context()
	account, err := h.useCase.GetByID(ctx, id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "weather account not found",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    account,
	})
}

func (h *WeatherAccountHandler) Create(c fiber.Ctx) error {
	var input domain.CreateWeatherAccountInput
	if err := c.Bind().Body(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "invalid input: " + err.Error(),
		})
	}

	ctx := c.Context()
	account, err := h.useCase.Create(ctx, input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    account,
		"message": "Weather account created successfully",
	})
}

func (h *WeatherAccountHandler) Update(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "invalid account id",
		})
	}

	var input domain.UpdateWeatherAccountInput
	if err := c.Bind().Body(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "invalid input: " + err.Error(),
		})
	}

	ctx := c.Context()
	account, err := h.useCase.Update(ctx, id, input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    account,
		"message": "Weather account updated successfully",
	})
}

func (h *WeatherAccountHandler) Delete(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "invalid account id",
		})
	}

	ctx := c.Context()
	if err := h.useCase.Delete(ctx, id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Weather account deleted successfully",
	})
}
