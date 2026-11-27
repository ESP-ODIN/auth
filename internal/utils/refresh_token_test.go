package utils_test

import (
	"auth/internal/utils"
	"encoding/base64"
	"strings"
	"testing"
)

func TestGenerateRefreshToken(t *testing.T) {
	token, err := utils.GenerateRefreshToken()
	if err != nil {
		t.Fatal(err)
	}
	// base64url sans padding : pas de '+', '/' ni '='
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		t.Fatalf("token is not raw base64url: %v", err)
	}
	if len(raw) < 32 {
		t.Fatalf("token carries %d random bytes, want at least 32", len(raw))
	}
	if strings.ContainsAny(token, "+/=") {
		t.Fatalf("token %q is not URL-safe", token)
	}

	seen := map[string]bool{token: true}
	for i := 0; i < 100; i++ {
		other, err := utils.GenerateRefreshToken()
		if err != nil {
			t.Fatal(err)
		}
		if seen[other] {
			t.Fatal("generated the same token twice")
		}
		seen[other] = true
	}
}

func TestHashRefreshToken(t *testing.T) {
	// vecteur SHA-256 connu de "abc"
	const want = "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
	if got := utils.HashRefreshToken("abc"); got != want {
		t.Fatalf("hash = %s, want %s", got, want)
	}

	token, err := utils.GenerateRefreshToken()
	if err != nil {
		t.Fatal(err)
	}
	hash := utils.HashRefreshToken(token)
	if hash != utils.HashRefreshToken(token) {
		t.Fatal("hash is not deterministic")
	}
	if len(hash) != 64 {
		t.Fatalf("hash length = %d, want 64 hex characters", len(hash))
	}
	if strings.Contains(hash, token) || hash == token {
		t.Fatal("hash exposes the token")
	}
	other, err := utils.GenerateRefreshToken()
	if err != nil {
		t.Fatal(err)
	}
	if utils.HashRefreshToken(other) == hash {
		t.Fatal("different tokens share a hash")
	}
}
