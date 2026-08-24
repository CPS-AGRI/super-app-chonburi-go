package jwtutil

import (
	"os"
	"strings"
	"time"

	"super-app-chonburi-go/internal/domain"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

type CustomClaims struct {
	ID          string   `json:"id"`
	Email       string   `json:"email"`
	Name        string   `json:"name"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

func secretKey() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "super-secret-enterprise-key-12345"
	}
	return []byte(secret)
}

func GenerateToken(user domain.User) (string, error) {

	name := ""
	permissions := []string{}
	if admin, ok := user.(*domain.Admin); ok {
		name = admin.Name + " " + admin.LastName

		if admin.Role != nil && (strings.EqualFold(admin.Role.Type, "superadmin") || strings.EqualFold(admin.Role.Type, "super_admin")) {
			permissions = append(permissions, "MANAGE_CITY", "MANAGE_ADMINS", "MANAGE_DEPARTMENTS", "VIEW_ALL_REPORTS")
		}

		uniqueKeys := make(map[string]bool)
		for _, dept := range admin.Departments {
			for _, module := range dept.Modules {
				if module.Key != nil && *module.Key != "" {
					uniqueKeys[*module.Key] = true
				}

				if module.ID == "d01b2ce5-34a9-498b-bba0-b1b8360f1ea9" ||
					module.NameTh == "ศูนย์ร้องทุกข์" ||
					module.NameEn == "Complaint Center" {
					uniqueKeys["ModuleComplaintCenter"] = true
				}
			}
		}

		for key := range uniqueKeys {
			permissions = append(permissions, key)
		}
	}

	claims := CustomClaims{
		ID:          user.GetID(),
		Email:       user.GetEmail(),
		Name:        name,
		Role:        user.GetRole(),
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey())
}

func ParseToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(t *jwt.Token) (interface{}, error) {
		return secretKey(), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return claims, nil
}

func RequireAuth() fiber.Handler {
	return func(c fiber.Ctx) error {
		var tokenString string

		authHeader := c.Get("Authorization")
		if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		}

		if tokenString == "" {
			tokenString = c.Cookies("auth_token")
		}

		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: Missing or invalid token",
			})
		}

		claims, err := ParseToken(tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: " + err.Error(),
			})
		}

		c.Locals("user", claims)
		return c.Next()
	}
}
