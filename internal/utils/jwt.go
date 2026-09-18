package utils

import (
	"time"

	"github.com/google/uuid"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
	jwt.RegisteredClaims
}

func GenerateJWT(id uuid.UUID, email string, secretKey []byte) (string, error) {
	claims := Claims{
		ID:    id,
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}
