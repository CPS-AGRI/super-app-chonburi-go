package middleware

import (
	"log"
	"super-app-chonburi-go/internal/domain"
	"super-app-chonburi-go/pkg/jwtutil"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

func AuditLog(repo domain.AuditLogRepository) fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()

		var userID, roleID string
		var tokenString string

		authHeader := c.Get("Authorization")
		if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		}

		if tokenString == "" {
			tokenString = c.Cookies("auth_token")
		}

		if tokenString != "" {
			claims, parseErr := jwtutil.ParseToken(tokenString)
			if parseErr == nil {
				userID = claims.ID
				roleID = claims.Role
				c.Locals("user", claims)
			}
		}

		traceID := c.Get("X-Trace-ID")
		if traceID == "" {
			traceID = c.Get("X-Request-ID")
		}
		if traceID == "" {
			traceID = uuid.New().String()
		}

		err := c.Next()

		stop := time.Now()

		logRecord := &domain.AuditLog{
			ID:                 uuid.New().String(),
			TraceId:            traceID,
			UserId:             userID,
			RoleId:             roleID,
			Method:             c.Method(),
			Path:               c.Path(),
			ResponseStatusCode: c.Response().StatusCode(),
			IPAddress:          c.IP(),
			UserAgent:          c.Get("User-Agent"),
			RequestTime:        start,
			ResponseTime:       stop,
		}

		go func(record *domain.AuditLog) {
			if createErr := repo.Create(record); createErr != nil {
				log.Printf("⚠️ [AuditLog] Failed to record audit log: %v", createErr)
			}
		}(logRecord)

		return err
	}
}
