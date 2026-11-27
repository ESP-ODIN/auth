package handler

import (
	"auth/internal/dto"
	appErrors "auth/internal/errors"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakeRefreshTokenService struct {
	got  string
	resp *dto.RefreshTokenResponse
	err  error
}

func (f *fakeRefreshTokenService) Refresh(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.RefreshTokenResponse, error) {
	f.got = req.RefreshToken
	return f.resp, f.err
}

func TestRefreshTokenHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tokens := &dto.RefreshTokenResponse{AccessToken: "new-access", RefreshToken: "new-refresh"}
	for _, tc := range []struct {
		name, body string
		resp       *dto.RefreshTokenResponse
		err        error
		status     int
		message    string
	}{
		{"succès", `{"refresh_token":"old-refresh"}`, tokens, nil, http.StatusOK, "tokens refreshed successfully"},
		{"body vide", `{}`, nil, nil, http.StatusBadRequest, ""},
		{"token vide", `{"refresh_token":""}`, nil, nil, http.StatusBadRequest, ""},
		{"ancienne clé token", `{"token":"old-refresh"}`, nil, nil, http.StatusBadRequest, ""},
		{"json invalide", `{`, nil, nil, http.StatusBadRequest, ""},
		{"token refusé", `{"refresh_token":"old-refresh"}`, nil, appErrors.ErrorInvalidRefreshToken, http.StatusUnauthorized, "invalid refresh token"},
		{"erreur interne", `{"refresh_token":"old-refresh"}`, nil, errors.New("database unavailable: secret detail"), http.StatusInternalServerError, "internal server error"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc := &fakeRefreshTokenService{resp: tc.resp, err: tc.err}
			r := gin.New()
			r.POST("/refresh", NewRefreshTokenHandler(svc).Handle)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/refresh", strings.NewReader(tc.body)))

			if rec.Code != tc.status {
				t.Fatalf("status %d, expected %d (body %s)", rec.Code, tc.status, rec.Body)
			}
			var body struct {
				Message string `json:"message"`
				Data    *struct {
					AccessToken  string `json:"access_token"`
					RefreshToken string `json:"refresh_token"`
				} `json:"data"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("invalid JSON body: %v", err)
			}
			if tc.message != "" && body.Message != tc.message {
				t.Fatalf("message %q, expected %q", body.Message, tc.message)
			}
			if tc.status == http.StatusOK {
				if svc.got != "old-refresh" {
					t.Fatalf("service received %q", svc.got)
				}
				if body.Data == nil || body.Data.AccessToken != "new-access" || body.Data.RefreshToken != "new-refresh" {
					t.Fatalf("unexpected data: %s", rec.Body)
				}
				return
			}
			// En erreur : aucun token, et aucun détail interne ne fuit
			if body.Data != nil || strings.Contains(rec.Body.String(), "secret detail") {
				t.Fatalf("error response leaks data: %s", rec.Body)
			}
		})
	}
}
