package config

import (
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	setJWTEnv(t)
	for _, tc := range []struct {
		value, want string
		invalid     bool
	}{
		{"", "127.0.0.1:8080", false},
		{"127.0.0.1:9090", "127.0.0.1:9090", false},
		{"[::1]:8080", "[::1]:8080", false},
		{"localhost", "", true},
		{"localhost:abc", "", true},
		{"localhost:0", "", true},
		{"localhost:65536", "", true},
	} {
		t.Run(tc.value, func(t *testing.T) {
			t.Setenv("HTTP_ADDR", tc.value)
			t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db")
			cfg, err := Load()
			if (err != nil) != tc.invalid {
				t.Fatalf("Load error = %v, invalid = %v", err, tc.invalid)
			}
			if !tc.invalid && cfg.HTTPAddr != tc.want {
				t.Fatalf("address = %q, want %q", cfg.HTTPAddr, tc.want)
			}
		})
	}
}

func TestLoadDatabaseURL(t *testing.T) {
	t.Setenv("HTTP_ADDR", "127.0.0.1:8080")
	t.Setenv("DATABASE_URL", "")
	if _, err := Load(); err == nil {
		t.Fatal("expected error when DATABASE_URL is missing")
	}
}

func setJWTEnv(t *testing.T) {
	t.Helper()
	t.Setenv("JWT_PRIVATE_KEY_PATH", "secrets/private.pem")
	t.Setenv("JWT_KEY_ID", "dev-key")
	t.Setenv("JWT_ISSUER", "https://auth.odin.test")
	t.Setenv("JWT_AUDIENCE", "odin-api")
	t.Setenv("ACCESS_TOKEN_TTL", "15m")
	t.Setenv("REFRESH_TOKEN_TTL", "24h")
}

func TestLoadJWTConfig(t *testing.T) {
	t.Setenv("HTTP_ADDR", "127.0.0.1:8080")
	t.Setenv("DATABASE_URL", "postgres://localhost/test")
	setJWTEnv(t)
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.JWTPrivateKeyPath != "secrets/private.pem" || cfg.JWTKeyID != "dev-key" || cfg.JWTIssuer != "https://auth.odin.test" || cfg.JWTAudience != "odin-api" {
		t.Fatal("JWT config mismatch")
	}
	for _, name := range []string{"JWT_PRIVATE_KEY_PATH", "JWT_KEY_ID", "JWT_ISSUER", "JWT_AUDIENCE"} {
		t.Run(name, func(t *testing.T) {
			t.Setenv(name, " ")
			if _, err := Load(); err == nil {
				t.Fatalf("accepted missing %s", name)
			}
		})
	}
}

func TestLoadTokenTTL(t *testing.T) {
	t.Setenv("HTTP_ADDR", "127.0.0.1:8080")
	t.Setenv("DATABASE_URL", "postgres://localhost/test")
	setJWTEnv(t)
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AccessTokenTTL != 15*time.Minute || cfg.RefreshTokenTTL != 24*time.Hour {
		t.Fatalf("TTL = %v / %v, want 15m / 24h", cfg.AccessTokenTTL, cfg.RefreshTokenTTL)
	}
	for _, name := range []string{"ACCESS_TOKEN_TTL", "REFRESH_TOKEN_TTL"} {
		for _, value := range []string{"", " ", "abc", "15", "0s", "-5m"} {
			t.Run(name+"="+value, func(t *testing.T) {
				t.Setenv(name, value)
				if _, err := Load(); err == nil {
					t.Fatalf("accepted %s=%q", name, value)
				}
			})
		}
	}
}
