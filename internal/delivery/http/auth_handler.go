package http

import (
	"super-app-chonburi-go/internal/domain"
	"super-app-chonburi-go/internal/usecase"
	"github.com/gofiber/fiber/v3"
)

type AuthHandler struct {
	authUseCase usecase.AuthUseCase
}

func NewAuthHandler(authUseCase usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{authUseCase: authUseCase}
}

func (h *AuthHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1/auth")
	api.Post("/login", h.Login)
}

func (h *AuthHandler) Login(c fiber.Ctx) error {
	var req domain.LoginRequest
	
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Bad Request",
		})
	}

	resp, err := h.authUseCase.Login(req)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.JSON(resp)
}



