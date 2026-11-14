package handler

import (
	"auth/internal/dto"
	"auth/internal/errors"
	"auth/internal/service"
	"auth/internal/utils/http/response"
	stdErrors "errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RefreshTokenHandler struct {
	refreshTokenService service.RefreshTokenServiceInterface
}

func NewRefreshTokenHandler(refreshTokenService service.RefreshTokenServiceInterface) *RefreshTokenHandler {
	return &RefreshTokenHandler{refreshTokenService: refreshTokenService}
}

// @Summary Refresh the tokens
// @Description Exchange a valid refresh token for a new access token (JWT) and a new refresh token. The refresh token sent is revoked (rotation) and cannot be used again. Reusing an already revoked refresh token revokes every refresh token of the user.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} dto.RefreshTokenResponse "Tokens refreshed successfully"
// @Failure 400 {object} string "Bad request"
// @Failure 401 {object} string "Invalid, expired or revoked refresh token"
// @Failure 500 {object} string "Internal server error"
// @Router /api/v1/auth/refresh [post]
func (h *RefreshTokenHandler) Handle(c *gin.Context) {
	var req dto.RefreshTokenRequest

	// 1. Validation automatique via les tags binding du DTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResponseWithError(c.Writer, http.StatusBadRequest, err.Error())
		return
	}

	// 2. Appel du service métier
	refreshResponse, err := h.refreshTokenService.Refresh(c.Request.Context(), &req)
	if err != nil {
		if stdErrors.Is(err, errors.ErrorInvalidRefreshToken) {
			// 401 générique : on ne dit pas si le token est inconnu, expiré ou révoqué
			response.ResponseWithError(c.Writer, http.StatusUnauthorized, errors.ErrorInvalidRefreshToken.Error())
			return
		}
		log.Printf("refresh: unexpected error: %v", err)
		response.ResponseWithError(c.Writer, http.StatusInternalServerError, "internal server error")
		return
	}

	// 3. Réponse en cas de succès
	response.ResponseWithSuccess(c.Writer, http.StatusOK, "tokens refreshed successfully", refreshResponse)
}
