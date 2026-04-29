package middleware

import (
	"super-app-chonburi-go/internal/domain"
	"super-app-chonburi-go/pkg/jwtutil"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

func AuditLog(repo domain.ActivityLogRepository) fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()

		// 1. Try to get user from token and set context BEFORE calling next handler
		var adminID, adminName string
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
				adminID = claims.ID
				adminName = claims.Name
				// Set to Locals so Handlers can use it during their execution
				c.Locals("user", claims)
			}
		}

		// 2. Execute next handler (the controller logic)
		err := c.Next()

		// 3. After handler finished, calculate metrics and save log
		stop := time.Now()
		duration := stop.Sub(start).Milliseconds()

		// Create activity log record
		logRecord := &domain.ActivityLog{
			ID:                 uuid.New(),
			TraceID:            c.Get("X-Trace-ID"),
			AdminID:            adminID,
			AdminName:          adminName,
			Method:             c.Method(),
			Path:               c.Path(),
			ResponseStatusCode: c.Response().StatusCode(),
			IPAddress:          c.IP(),
			UserAgent:          c.Get("User-Agent"),
			RequestTime:        start,
			ResponseTime:       stop,
			DurationMs:         duration,
		}

		// Save log asynchronously
		go repo.Create(logRecord)

		return err
	}
}
