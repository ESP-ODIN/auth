package utils_test

import (
	"auth/internal/testutil"
	"auth/internal/utils"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGenerateJWT(t *testing.T) {
	signer, publicKey := testutil.Signer(t)
	id := uuid.New()
	raw, err := signer.GenerateJWT(id, "user@odin.test")
	if err != nil {
		t.Fatal(err)
	}
	claims := &utils.Claims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(*jwt.Token) (any, error) { return publicKey, nil }, jwt.WithValidMethods([]string{"RS256"}), jwt.WithIssuer("https://auth.odin.test"), jwt.WithAudience("odin-api"), jwt.WithExpirationRequired(), jwt.WithIssuedAt())
	if err != nil || !token.Valid {
		t.Fatalf("invalid token: %v", err)
	}
	if token.Method.Alg() != "RS256" || token.Header["kid"] != "test-key" {
		t.Fatal("incorrect algorithm/kid")
	}
	if claims.ID != id || claims.Email != "user@odin.test" {
		t.Fatal("identity changed")
	}
	if claims.IssuedAt == nil || claims.ExpiresAt == nil || claims.ExpiresAt.Sub(claims.IssuedAt.Time) != 24*time.Hour {
		t.Fatal("expected 24 hour lifetime")
	}
	if time.Since(claims.IssuedAt.Time) < 0 || time.Since(claims.IssuedAt.Time) > 5*time.Second {
		t.Fatal("incorrect iat")
	}
	_, wrongKey := testutil.Signer(t)
	if _, err := jwt.Parse(raw, func(*jwt.Token) (any, error) { return wrongKey, nil }, jwt.WithValidMethods([]string{"RS256"})); err == nil {
		t.Fatal("accepted wrong key")
	}
	for _, method := range []jwt.SigningMethod{jwt.SigningMethodHS256, jwt.SigningMethodNone} {
		var key any = []byte("test-only-secret")
		if method == jwt.SigningMethodNone {
			key = jwt.UnsafeAllowNoneSignatureType
		}
		forged, err := jwt.NewWithClaims(method, claims).SignedString(key)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := jwt.Parse(forged, func(*jwt.Token) (any, error) { return publicKey, nil }, jwt.WithValidMethods([]string{"RS256"})); err == nil {
			t.Fatal("accepted unexpected algorithm")
		}
	}
}

func TestLoadJWTSigner(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	pkcs8, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	weak, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name  string
		data  []byte
		valid bool
	}{
		{"pkcs8", pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: pkcs8}), true},
		{"malformed", []byte("not a key"), false},
		{"public-only", pem.EncodeToMemory(&pem.Block{Type: "RSA PUBLIC KEY", Bytes: x509.MarshalPKCS1PublicKey(&key.PublicKey)}), false},
		{"weak", pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(weak)}), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "key.pem")
			if err := os.WriteFile(path, tc.data, 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := utils.LoadJWTSigner(path, "key", "issuer", "audience"); (err == nil) != tc.valid {
				t.Fatalf("unexpected load result: %v", err)
			}
		})
	}
	if _, err := utils.LoadJWTSigner(filepath.Join(t.TempDir(), "missing"), "key", "issuer", "audience"); err == nil {
		t.Fatal("accepted missing file")
	}
}
