package middleware

import (
	"strings"
	"super-app-chonburi-go/internal/domain"
	"super-app-chonburi-go/pkg/jwtutil"

	"github.com/gofiber/fiber/v3"
)

// DenyExecutivesMutation blocks Executives role from POST, PUT, PATCH, DELETE operations on operational endpoints
func DenyExecutivesMutation() fiber.Handler {
	return func(c fiber.Ctx) error {
		// Allow auth actions and KPI threshold configuration
		if strings.HasPrefix(c.Path(), "/api/v1/auth") || strings.HasPrefix(c.Path(), "/api/v1/kpi/thresholds") {
			return c.Next()
		}

		method := c.Method()
		if method == fiber.MethodPost || method == fiber.MethodPut || method == fiber.MethodPatch || method == fiber.MethodDelete {
			userClaims, _ := c.Locals("user").(*jwtutil.CustomClaims)
			if userClaims == nil {
				authHeader := c.Get("Authorization")
				if strings.HasPrefix(authHeader, "Bearer ") {
					tokenString := strings.TrimPrefix(authHeader, "Bearer ")
					if claims, err := jwtutil.ParseToken(tokenString); err == nil {
						userClaims = claims
						c.Locals("user", claims)
					}
				} else if cookieToken := c.Cookies("auth_token"); cookieToken != "" {
					if claims, err := jwtutil.ParseToken(cookieToken); err == nil {
						userClaims = claims
						c.Locals("user", claims)
					}
				}
			}

			if userClaims != nil {
				if strings.EqualFold(userClaims.Role, domain.RoleExecutives) || strings.EqualFold(userClaims.Role, "executive") {
					return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
						"success": false,
						"error":   "ผู้บริหารมีสิทธิ์เข้าถึงเฉพาะข้อมูลสรุปภาพรวม (Dashboard) เท่านั้น ไม่สามารถแก้ไขหรือบันทึกข้อมูลปฏิบัติการได้",
					})
				}
			}
		}
		return c.Next()
	}
}
