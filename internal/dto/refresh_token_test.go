package dto

import (
	"testing"

	"github.com/gin-gonic/gin/binding"
	"github.com/stretchr/testify/assert"
)

func TestRefreshToken(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(r *RefreshTokenRequest)
		wantErr bool
	}{
		{"requête valide", func(r *RefreshTokenRequest) {}, false},

		// token
		{"token manquant", func(r *RefreshTokenRequest) { r.Token = "" }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := RefreshTokenRequest{
				Token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			}
			tt.modify(&r)

			err := binding.Validator.ValidateStruct(r)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}	