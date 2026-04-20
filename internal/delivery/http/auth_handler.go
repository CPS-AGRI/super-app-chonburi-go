package http

import (
	"super-app-chonburi-go/internal/domain"
	"super-app-chonburi-go/pkg/jwtutil"
	"time"

	"github.com/gofiber/fiber/v3"
)

type AuthHandler struct {
	uc domain.AuthUseCase
}

func NewAuthHandler(uc domain.AuthUseCase) *AuthHandler {
	return &AuthHandler{uc: uc}
}

func (h *AuthHandler) RegisterRoutes(app *fiber.App) {
	auth := app.Group("/api/v1/auth")
	auth.Post("/login", h.Login)
	auth.Post("/logout", h.Logout)
	auth.Get("/me", jwtutil.RequireAuth(), h.Me)
}

func (h *AuthHandler) Login(c fiber.Ctx) error {
	var req domain.LoginRequest
	if err := c.Bind().JSON(&req); err != nil {
		return ErrorResponse(c, "Invalid request", fiber.StatusBadRequest)
	}

	token, user, err := h.uc.Login(req.Email, req.Password)
	if err != nil {
		return ErrorResponse(c, "Invalid email or password", fiber.StatusUnauthorized)
	}

	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		HTTPOnly: true,
		SameSite: "Lax",
	})

	return SuccessResponse(c, fiber.Map{
		"user":  user,
		"token": token,
	})
}

func (h *AuthHandler) Logout(c fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
	})
	return SuccessResponse(c, fiber.Map{"message": "Logged out successfully"})
}

func (h *AuthHandler) Me(c fiber.Ctx) error {
	user := c.Locals("user")
	return SuccessResponse(c, user)
}
