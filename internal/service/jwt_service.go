package service

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/wmp/auth-service/internal/config"
)

type JWTService struct {
	secret       []byte
	expirationMs int64
}

func NewJWTService(cfg *config.Config) *JWTService {
	return &JWTService{
		secret:       []byte(cfg.JWTSecret),
		expirationMs: cfg.JWTExpirationMs,
	}
}

func (s *JWTService) GenerateToken(userID uuid.UUID, email string) (string, error) {
	now := time.Now()
	expiry := now.Add(time.Duration(s.expirationMs) * time.Millisecond)

	claims := jwt.MapClaims{
		"sub":   userID.String(),
		"email": email,
		"iat":   jwt.NewNumericDate(now),
		"exp":   jwt.NewNumericDate(expiry),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}
