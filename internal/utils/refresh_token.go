package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

// refreshTokenBytes est la quantité d'aléa d'un refresh token (256 bits).
const refreshTokenBytes = 32

// GenerateRefreshToken renvoie un refresh token opaque : 32 octets aléatoires encodés en base64url.
// C'est la valeur envoyée au client ; elle ne doit jamais être stockée telle quelle.
func GenerateRefreshToken() (string, error) {
	b := make([]byte, refreshTokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate refresh token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// HashRefreshToken renvoie le SHA-256 du token en hexadécimal.
// C'est la valeur stockée en base (colonne token_hash).
func HashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
