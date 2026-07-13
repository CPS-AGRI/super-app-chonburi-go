package http

import (
	"errors"
	"super-app-chonburi-go/pkg/jwtutil"

	"github.com/gofiber/fiber/v3"
)

func GetAdminIDFromClaims(c fiber.Ctx) (string, error) {
	userClaims, ok := c.Locals("user").(*jwtutil.CustomClaims)
	if !ok {
		return "", errors.New("unauthorized")
	}
	return userClaims.ID, nil
}
