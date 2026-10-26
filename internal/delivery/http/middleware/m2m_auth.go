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

// M2MOrJWT authenticates request via Either JWT Bearer (Web Admin) OR HMAC-SHA256 M2M signature (Pattaya Dashboard)
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

		// 2. Verify M2M HMAC Signature (Pattaya Dashboard / External Consumers)
		if err := verifyM2M(c, cfg); err == nil {
			c.Locals("auth_type", "m2m")
			return c.Next()
		} else {
			return err
		}
	}
}

func verifyM2M(c fiber.Ctx, cfg *config.Config) error {
	m2mKey := c.Get("X-M2M-Key")
	tsHeader := c.Get("X-Timestamp")
	nonce := c.Get("X-Nonce")
	signature := c.Get("X-Signature")

	if m2mKey == "" || tsHeader == "" || nonce == "" || signature == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Unauthorized: Missing required M2M security headers (X-M2M-Key, X-Timestamp, X-Nonce, X-Signature)",
		})
	}

	secretKey := cfg.M2MSecretKey
	if secretKey == "" {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"success": false,
			"error":   "Server Error: M2M authentication key not configured",
		})
	}

	// 1. Timestamp freshness check (Replay Attack Prevention - window ±300s)
	ts, err := strconv.ParseInt(tsHeader, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Unauthorized: Invalid X-Timestamp format",
		})
	}
	now := time.Now().Unix()
	if math.Abs(float64(now-ts)) > 300 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Unauthorized: Request expired or timestamp skew too large (> 5m)",
		})
	}

	// 2. Nonce single-use check (Prevent replay attacks within freshness window)
	if redis.Client != nil {
		nonceKey := fmt.Sprintf("m2m:nonce:%s:%s", m2mKey, nonce)
		set, err := redis.Client.SetNX(c.Context(), nonceKey, "1", 300*time.Second).Result()
		if err == nil && !set {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Unauthorized: Replay nonce detected",
			})
		}
	}

	// 3. Construct Canonical String
	method := strings.ToUpper(c.Method())
	path := c.Path()

	// Sort query parameters deterministically
	var sortedQuery string
	reqURI, err := url.ParseRequestURI(string(c.Request().RequestURI()))
	if err == nil {
		sortedQuery = reqURI.Query().Encode()
	}

	canonical := fmt.Sprintf("%s\n%s\n%s\n%s\n%s", method, path, sortedQuery, tsHeader, nonce)

	// 4. Compute HMAC-SHA256
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(canonical))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(strings.ToLower(signature)), []byte(strings.ToLower(expectedSig))) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Unauthorized: Invalid M2M signature",
		})
	}

	return nil
}
