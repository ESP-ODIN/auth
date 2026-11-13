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

func registerBody(password string) string {
	return `{"email":"user@odin.test","password":"` + password + `","username":"janedoe","first_name":"Jane","last_name":"Doe","date_of_birth":"1990-01-31"}`
}

func TestRegisterPasswordLength(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name, body string
		status     int
	}{
		{"100 chars", registerBody(strings.Repeat("a", 100)), http.StatusCreated},
		{"101 chars", registerBody(strings.Repeat("a", 101)), http.StatusBadRequest},
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
