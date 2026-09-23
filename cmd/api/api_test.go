package main

import (
	"auth/internal/testutil"
	"auth/internal/utils"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRoutes(t *testing.T) {
	signer, _ := testutil.Signer(t)
	r := routes(nil, signer)
	for _, tc := range []struct {
		method, path, body string
		status             int
	}{
		{http.MethodGet, "/.well-known/jwks.json", "", http.StatusOK},
		{http.MethodPost, "/.well-known/jwks.json", "", http.StatusNotFound},
		{http.MethodGet, "/unknown", "", http.StatusNotFound},
		{http.MethodPost, "/api/v1/auth/register", "{}", http.StatusBadRequest},
	} {
		t.Run(tc.method+tc.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body)))
			if rec.Code != tc.status {
				t.Fatalf("status %d, expected %d", rec.Code, tc.status)
			}
		})
	}
}

func TestJWKSVerifiesIssuedToken(t *testing.T) {
	signer, _ := testutil.Signer(t)
	rec := httptest.NewRecorder()
	routes(nil, signer).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/.well-known/jwks.json", nil))
	if rec.Code != http.StatusOK || !strings.HasPrefix(rec.Header().Get("Content-Type"), "application/json") {
		t.Fatal("invalid JWKS response")
	}
	var doc struct {
		Keys []map[string]string `json:"keys"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &doc); err != nil {
		t.Fatal(err)
	}
	if len(doc.Keys) != 1 {
		t.Fatal("expected single signing key")
	}
	k := doc.Keys[0]
	allowed := map[string]bool{"kty": true, "use": true, "alg": true, "kid": true, "n": true, "e": true}
	if len(k) != len(allowed) {
		t.Fatal("unexpected/private JWK fields")
	}
	for name := range k {
		if !allowed[name] {
			t.Fatalf("unexpected/private field %s", name)
		}
	}
	if k["kty"] != "RSA" || k["alg"] != "RS256" || k["use"] != "sig" {
		t.Fatal("incorrect JWK metadata")
	}
	n, err := base64.RawURLEncoding.DecodeString(k["n"])
	if err != nil {
		t.Fatal(err)
	}
	e, err := base64.RawURLEncoding.DecodeString(k["e"])
	if err != nil {
		t.Fatal(err)
	}
	publicKey := &rsa.PublicKey{N: new(big.Int).SetBytes(n), E: int(new(big.Int).SetBytes(e).Int64())}
	id := uuid.New()
	raw, err := signer.GenerateJWT(id, "user@odin.test")
	if err != nil {
		t.Fatal(err)
	}
	claims := &utils.Claims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		if token.Header["kid"] != k["kid"] {
			t.Fatal("JWT/JWKS kid mismatch")
		}
		return publicKey, nil
	}, jwt.WithValidMethods([]string{"RS256"}), jwt.WithIssuer("https://auth.odin.test"), jwt.WithAudience("odin-api"), jwt.WithExpirationRequired())
	if err != nil || !token.Valid {
		t.Fatalf("JWKS cannot verify token: %v", err)
	}
	if claims.ID != id || claims.Email != "user@odin.test" {
		t.Fatal("identity mismatch")
	}
}
