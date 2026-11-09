package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"

	"super-app-chonburi-go/config"
	"super-app-chonburi-go/pkg/jwtutil"
	"super-app-chonburi-go/pkg/redis"

	"github.com/gofiber/fiber/v3"
)

// M2MOrJWT authenticates request via:
// 1. JWT Bearer token / Cookie (Chonburi Web Admin)
// 2. Internal Microservice Secret header: X-Internal-Secret (Alert Service, etc.)
// 3. HMAC-SHA256 M2M signature (Pattaya Dashboard / External Consumers)
func M2MOrJWT(cfg *config.Config) fiber.Handler {
	return func(c fiber.Ctx) error {
		// 1. Try JWT Bearer / Cookie first (Chonburi Web Admin)
		authHeader := c.Get("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if claims, err := jwtutil.ParseToken(tokenString); err == nil {
				c.Locals("user", claims)
				c.Locals("auth_type", "jwt")
				return c.Next()
			}
		}

		if cookieToken := c.Cookies("auth_token"); cookieToken != "" {
			if claims, err := jwtutil.ParseToken(cookieToken); err == nil {
				c.Locals("user", claims)
				c.Locals("auth_type", "jwt_cookie")
				return c.Next()
			}
		}

		// 2. Try Internal Microservice Secret (chon-alert-notification, internal workers)
		internalSecret := c.Get("X-Internal-Secret")
		if internalSecret != "" && cfg.M2MSecretKey != "" && internalSecret == cfg.M2MSecretKey {
			c.Locals("auth_type", "internal_service")
			return c.Next()
		}

		// 3. Verify M2M HMAC Signature (Pattaya Dashboard / External Consumers)
		m2mKey := c.Get("X-M2M-Key")
		if m2mKey != "" {
			if err := verifyM2M(c, cfg); err == nil {
				c.Locals("auth_type", "m2m")
				return c.Next()
			} else {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"success": false,
					"error":   err.Error(),
				})
			}
		}

		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Unauthorized: Invalid or missing authentication credentials (JWT Bearer, X-Internal-Secret, or M2M HMAC signature required)",
		})
	}
}

func verifyM2M(c fiber.Ctx, cfg *config.Config) error {
	m2mKey := c.Get("X-M2M-Key")
	tsHeader := c.Get("X-Timestamp")
	nonce := c.Get("X-Nonce")
	signature := c.Get("X-Signature")

	if m2mKey == "" || tsHeader == "" || nonce == "" || signature == "" {
		return fmt.Errorf("Missing required M2M security headers (X-M2M-Key, X-Timestamp, X-Nonce, X-Signature)")
	}

	secretKey := cfg.M2MSecretKey
	if secretKey == "" {
		return fmt.Errorf("Server Error: M2M authentication key not configured")
	}

	// 1. Timestamp freshness check (Replay Attack Prevention - window ±300s)
	ts, err := strconv.ParseInt(tsHeader, 10, 64)
	if err != nil {
		return fmt.Errorf("Invalid X-Timestamp format")
	}
	now := time.Now().Unix()
	if math.Abs(float64(now-ts)) > 300 {
		return fmt.Errorf("Request expired or timestamp skew too large (> 5m)")
	}

	// 2. Construct Canonical String
	method := strings.ToUpper(c.Method())
	path := c.Path()

	// Sort query parameters deterministically
	var sortedQuery string
	reqURI, err := url.ParseRequestURI(string(c.Request().RequestURI()))
	if err == nil {
		sortedQuery = reqURI.Query().Encode()
	}

	canonical := fmt.Sprintf("%s\n%s\n%s\n%s\n%s", method, path, sortedQuery, tsHeader, nonce)

	// 3. Compute and Verify HMAC-SHA256 FIRST (Prevent unauthenticated cache flooding / DoS)
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(canonical))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(strings.ToLower(signature)), []byte(strings.ToLower(expectedSig))) {
		return fmt.Errorf("Invalid M2M signature")
	}

	// 4. Nonce single-use check (Prevent replay attacks within freshness window)
	// Committed ONLY after signature has been authenticated
	if redis.Client != nil {
		nonceKey := fmt.Sprintf("m2m:nonce:%s:%s", m2mKey, nonce)
		set, err := redis.Client.SetNX(c.Context(), nonceKey, "1", 300*time.Second).Result()
		if err != nil {
			return fmt.Errorf("Service Unavailable: Security nonce store unreachable")
		}
		if !set {
			return fmt.Errorf("Replay nonce detected")
		}
	}

	return nil
}
