package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	DatabaseURL       string
	HTTPAddr          string
	JWTPrivateKeyPath string
	JWTKeyID          string
	JWTIssuer         string
	JWTAudience       string
	AccessTokenTTL    time.Duration
	RefreshTokenTTL   time.Duration
	FrontendURL       string
	SMTPHost          string
	SMTPPort          string
	SMTPUser          string
	SMTPPassword      string
	SMTPFrom          string
}

func Load() (Config, error) {
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = "127.0.0.1:8080"
	}
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		return Config{}, fmt.Errorf("invalid HTTP_ADDR %q: %w", addr, err)
	}
	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		return Config{}, fmt.Errorf("HTTP_ADDR port must be between 1 and 65535")
	}
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	cfg := Config{DatabaseURL: dbURL, HTTPAddr: addr,
		JWTPrivateKeyPath: os.Getenv("JWT_PRIVATE_KEY_PATH"),
		JWTKeyID:          os.Getenv("JWT_KEY_ID"),
		JWTIssuer:         os.Getenv("JWT_ISSUER"),
		JWTAudience:       os.Getenv("JWT_AUDIENCE"),
		FrontendURL:       os.Getenv("FRONTEND_URL"),
		SMTPHost:          os.Getenv("SMTP_HOST"),
		SMTPPort:          os.Getenv("SMTP_PORT"),
		SMTPUser:          os.Getenv("SMTP_USER"),
		SMTPPassword:      os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:          os.Getenv("SMTP_FROM"),
	}
	for _, entry := range []struct{ name, value string }{
		{"JWT_PRIVATE_KEY_PATH", cfg.JWTPrivateKeyPath},
		{"JWT_KEY_ID", cfg.JWTKeyID},
		{"JWT_ISSUER", cfg.JWTIssuer},
		{"JWT_AUDIENCE", cfg.JWTAudience},
		{"FRONTEND_URL", cfg.FrontendURL},
		{"SMTP_HOST", cfg.SMTPHost},
		{"SMTP_PORT", cfg.SMTPPort},
		{"SMTP_FROM", cfg.SMTPFrom},
	} {
		if strings.TrimSpace(entry.value) == "" {
			return Config{}, fmt.Errorf("%s is required", entry.name)
		}
	}

	cfg.AccessTokenTTL, err = loadDuration("ACCESS_TOKEN_TTL")
	if err != nil {
		return Config{}, err
	}
	cfg.RefreshTokenTTL, err = loadDuration("REFRESH_TOKEN_TTL")
	if err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// loadDuration lit une durée obligatoire et strictement positive (ex. "15m", "24h").
func loadDuration(name string) (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return 0, fmt.Errorf("%s is required", name)
	}
	d, err := time.ParseDuration(value)
	if err != nil || d <= 0 {
		return 0, fmt.Errorf("%s must be a valid positive duration", name)
	}
	return d, nil
}
