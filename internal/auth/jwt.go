package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

type TokenManager struct {
	secret []byte
	expiry time.Duration
}

func NewTokenManager(secret string, expiry time.Duration) *TokenManager {
	return &TokenManager{
		secret: []byte(secret),
		expiry: expiry,
	}
}

func (tm *TokenManager) GenerateToken(userID uuid.UUID) (string, error) {
	if userID == uuid.Nil {
		return "", fmt.Errorf("user ID cannot be nil")
	}

	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tm.expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(tm.secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

func (tm *TokenManager) ValidateToken(tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("token cannot be empty")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return tm.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
