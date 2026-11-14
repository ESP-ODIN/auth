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
		{"token manquant", func(r *RefreshTokenRequest) { r.RefreshToken = "" }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := RefreshTokenRequest{
				RefreshToken: "dGhpcy1pcy1hbi1leGFtcGxlLXJlZnJlc2gtdG9rZW4",
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
