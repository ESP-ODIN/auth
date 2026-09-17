package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRoutes(t *testing.T) {
	cases := []struct {
		method string
		path   string
		status int
	}{
		{method: http.MethodGet, path: "/name", status: http.StatusOK},
		{method: http.MethodPost, path: "/name", status: http.StatusMethodNotAllowed},
		{method: http.MethodGet, path: "/unknown", status: http.StatusNotFound},
	}

	for _, tc := range cases {
		t.Run(tc.method+tc.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.path, nil)
			routes().ServeHTTP(rec, req)

			// http verification
			if rec.Code != tc.status {
				t.Fatalf("status = %d, want %d", rec.Code, tc.status)
			}

			// JSON verification for successful GET /name request
			if tc.status == http.StatusOK {
				if got := rec.Header().Get("Content-Type"); got != "application/json" {
					t.Fatalf("Content-Type = %q, want %q", got, "application/json")
				}

				var body struct {
					Message string `json:"message"`
				}
				if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
					t.Fatalf("failed to decode JSON response: %v", err)
				}

				expectedMsg := "User name retrieved successfully"
				if body.Message != expectedMsg {
					t.Fatalf("message = %q, want %q", body.Message, expectedMsg)
				}
			}
		})
	}
}
