package dto

// RefreshTokenRequest est le body de POST /api/v1/auth/refresh.
// Le refresh token est une chaîne opaque en base64url, pas un JWT.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"dGhpcy1pcy1hbi1leGFtcGxlLXJlZnJlc2gtdG9rZW4"`
}

// RefreshTokenResponse contient le nouvel access token (JWT) et le nouveau refresh token issu de la rotation.
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token" example:"dGhpcy1pcy1hbi1leGFtcGxlLXJlZnJlc2gtdG9rZW4"`
}
