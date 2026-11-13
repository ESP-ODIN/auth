package handler

import (
	"auth/internal/dto"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakeLoginService struct{}

func (f fakeLoginService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	return &dto.LoginResponse{}, nil
}

func TestLoginPasswordLength(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name, body string
		status     int
	}{
		{"72 chars", `{"email":"user@odin.test","password":"` + strings.Repeat("a", 72) + `"}`, http.StatusOK},
		{"73 chars", `{"email":"user@odin.test","password":"` + strings.Repeat("a", 73) + `"}`, http.StatusBadRequest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()
			r.POST("/login", NewLoginHandler(fakeLoginService{}, nil).Handle)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(tc.body)))
			if rec.Code != tc.status {
				t.Fatalf("status %d, expected %d", rec.Code, tc.status)
			}
		})
	}
}
