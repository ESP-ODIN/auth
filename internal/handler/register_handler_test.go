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

type fakeRegisterService struct{}

func (f fakeRegisterService) Execute(ctx context.Context, req dto.RegisterRequest) (*dto.RegisterResponse, error) {
	return &dto.RegisterResponse{}, nil
}

func TestRegisterPasswordLength(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name, body string
		status     int
	}{
		{"72 chars", `{"email":"user@odin.test","password":"` + strings.Repeat("a", 72) + `"}`, http.StatusCreated},
		{"73 chars", `{"email":"user@odin.test","password":"` + strings.Repeat("a", 73) + `"}`, http.StatusBadRequest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()
			r.POST("/register", NewRegisterHandler(fakeRegisterService{}).Handle)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(tc.body)))
			if rec.Code != tc.status {
				t.Fatalf("status %d, expected %d", rec.Code, tc.status)
			}
		})
	}
}
