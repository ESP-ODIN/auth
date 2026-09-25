package utils

import (
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
	jwt.RegisteredClaims
}

// JWTSigner owns Auth's private key. Only public parameters are exported by JWKS.
type JWTSigner struct {
	privateKey *rsa.PrivateKey
	keyID      string
	issuer     string
	audience   string
}

type JWK struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type JWKS struct {
	Keys []JWK `json:"keys"`
}

// LoadJWTSigner accepts an unencrypted PKCS#1 or PKCS#8 RSA PEM file.
func LoadJWTSigner(path, keyID, issuer, audience string) (*JWTSigner, error) {
	if strings.TrimSpace(path) == "" || strings.TrimSpace(keyID) == "" || strings.TrimSpace(issuer) == "" || strings.TrimSpace(audience) == "" {
		return nil, fmt.Errorf("JWT private key path, key ID, issuer and audience are required")
	}
	pemBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read JWT private key: %w", err)
	}
	key, err := jwt.ParseRSAPrivateKeyFromPEM(pemBytes)
	if err != nil {
		return nil, fmt.Errorf("JWT private key must be a valid RSA PEM")
	}
	if err := key.Validate(); err != nil || key.N.BitLen() < 2048 {
		return nil, fmt.Errorf("JWT private key must be valid RSA with at least 2048 bits")
	}
	return &JWTSigner{privateKey: key, keyID: keyID, issuer: issuer, audience: audience}, nil
}

func (s *JWTSigner) JWKS() JWKS {
	return JWKS{Keys: []JWK{{
		Kty: "RSA", Use: "sig", Alg: "RS256", Kid: s.keyID,
		N: base64.RawURLEncoding.EncodeToString(s.privateKey.PublicKey.N.Bytes()),
		E: base64.RawURLEncoding.EncodeToString(big.NewInt(int64(s.privateKey.PublicKey.E)).Bytes()),
	}}}
}

func (s *JWTSigner) GenerateJWT(id uuid.UUID, email string) (string, error) {
	now := time.Now()
	claims := Claims{
		ID:    id,
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Audience:  jwt.ClaimStrings{s.audience},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour * 24)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = s.keyID
	return token.SignedString(s.privateKey)
}
