package http

import "github.com/gofiber/fiber/v3"

type APIResponse[T any] struct {
	Success bool   `json:"success"`
	Data    T      `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

func SuccessResponse[T any](c fiber.Ctx, data T, status ...int) error {
	code := fiber.StatusOK
	if len(status) > 0 {
		code = status[0]
	}

	return c.Status(code).JSON(APIResponse[T]{
		Success: true,
		Data:    data,
	})
}

func ErrorResponse(c fiber.Ctx, message string, status ...int) error {
	code := fiber.StatusInternalServerError
	if len(status) > 0 {
		code = status[0]
	}

	return c.Status(code).JSON(APIResponse[struct{}]{
		Success: false,
		Error:   message,
	})
}
