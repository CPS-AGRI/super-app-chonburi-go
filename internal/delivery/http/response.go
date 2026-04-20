package http

import "github.com/gofiber/fiber/v3"

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func SuccessResponse(c fiber.Ctx, data interface{}, status ...int) error {
	code := fiber.StatusOK
	if len(status) > 0 {
		code = status[0]
	}

	return c.Status(code).JSON(APIResponse{
		Success: true,
		Data:    data,
	})
}

func ErrorResponse(c fiber.Ctx, message string, status ...int) error {
	code := fiber.StatusInternalServerError
	if len(status) > 0 {
		code = status[0]
	}

	return c.Status(code).JSON(APIResponse{
		Success: false,
		Error:   message,
	})
}
