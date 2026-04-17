package jwtutil

import (
	"time"

	"super-app-chonburi-go/internal/domain"

	"github.com/golang-jwt/jwt/v5"
)

var SecretKey = []byte("super-secret-enterprise-key-12345") // In production, move to ENV

type CustomClaims struct {
	domain.User
	jwt.RegisteredClaims
}

func GenerateToken(user domain.User) (string, error) {
	claims := CustomClaims{
		User: user,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(SecretKey)
}
