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

func (h *AuthHandler) RegisterRoutes(app fiber.Router) {
	auth := app.Group("/auth")
	auth.Post("/login", h.Login)
	auth.Post("/refresh", h.Refresh)
	auth.Post("/logout", h.Logout)
	auth.Get("/me", jwtutil.RequireAuth(), h.Me)
	auth.Post("/forgot-password", h.ForgotPassword)
	auth.Get("/verify-token", h.VerifyToken)
	auth.Post("/reset-password", h.ResetPassword)
}

func (h *AuthHandler) Login(c fiber.Ctx) error {
	var req domain.LoginRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	token, refreshToken, user, err := h.uc.Login(req.Email, req.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid email or password"})
	}

	admin, permissions, _ := h.uc.Me(user.GetID())

	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		HTTPOnly: true,
		SameSite: "Lax",
	})

	return SuccessResponse(c, fiber.Map{
		"user": fiber.Map{
			"id":          admin.ID,
			"email":       admin.Email,
			"name":        admin.Name,
			"last_name":   admin.LastName,
			"role":        admin.Role,
			"permissions": permissions,
		},
		"token":        token,
		"refreshToken": refreshToken,
	})
}

func (h *AuthHandler) Refresh(c fiber.Ctx) error {
	var req struct {
		RefreshToken string `json:"refreshToken"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	token, refreshToken, user, err := h.uc.RefreshToken(req.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	admin, permissions, _ := h.uc.Me(user.GetID())

	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		HTTPOnly: true,
		SameSite: "Lax",
	})

	return SuccessResponse(c, fiber.Map{
		"user": fiber.Map{
			"id":          admin.ID,
			"email":       admin.Email,
			"name":        admin.Name,
			"last_name":   admin.LastName,
			"role":        admin.Role,
			"permissions": permissions,
		},
		"token":        token,
		"refreshToken": refreshToken,
	})
}

func (h *AuthHandler) Logout(c fiber.Ctx) error {

	var req struct {
		RefreshToken string `json:"refreshToken"`
	}
	if err := c.Bind().JSON(&req); err == nil && req.RefreshToken != "" {
		_ = h.uc.Logout(req.RefreshToken)
	}

	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
	})
	return SuccessResponse(c, fiber.Map{"message": "Logged out successfully"})
}

func (h *AuthHandler) Me(c fiber.Ctx) error {
	claims, ok := c.Locals("user").(*jwtutil.CustomClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid session"})
	}

	admin, permissions, err := h.uc.Me(claims.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return SuccessResponse(c, fiber.Map{
		"user": fiber.Map{
			"id":          admin.ID,
			"email":       admin.Email,
			"name":        admin.Name,
			"last_name":   admin.LastName,
			"role":        admin.Role,
			"permissions": permissions,
		},
	})
}

func (h *AuthHandler) ForgotPassword(c fiber.Ctx) error {
	var req domain.ForgotPasswordRequest
	if err := c.Bind().JSON(&req); err != nil || req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "กรุณาระบุอีเมลให้ถูกต้อง"})
	}

	if err := h.uc.ForgotPassword(req.Email); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return SuccessResponse(c, fiber.Map{
		"message": "ระบบได้ส่งลิงก์สำหรับรีเซ็ตรหัสผ่านไปยังอีเมลของคุณเรียบร้อยแล้ว",
	})
}

func (h *AuthHandler) VerifyToken(c fiber.Ctx) error {
	token := c.Query("token")
	tokenType := c.Query("type", "reset")

	if token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Token is required"})
	}

	admin, err := h.uc.VerifyToken(token, tokenType)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return SuccessResponse(c, fiber.Map{
		"email":     admin.Email,
		"name":      admin.Name,
		"last_name": admin.LastName,
		"valid":     true,
	})
}

func (h *AuthHandler) ResetPassword(c fiber.Ctx) error {
	var req domain.ResetPasswordRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ข้อมูลไม่ถูกต้อง"})
	}

	if req.Token == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "กรุณาระบุ Token และรหัสผ่านใหม่"})
	}

	if len(req.Password) < 8 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "รหัสผ่านต้องมีความยาวอย่างน้อย 8 ตัวอักษร"})
	}

	if err := h.uc.ResetPassword(req.Token, req.Password, req.TokenType); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return SuccessResponse(c, fiber.Map{
		"message": "ตั้งรหัสผ่านใหม่สำเร็จเรียบร้อย สามารถเข้าสู่ระบบได้ทันที",
	})
}
