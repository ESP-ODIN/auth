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

type ResetPasswordHandler struct {
	resetPasswordService service.ResetPasswordServiceInterface
}

func NewResetPasswordHandler(resetPasswordService service.ResetPasswordServiceInterface) *ResetPasswordHandler {
	return &ResetPasswordHandler{
		resetPasswordService: resetPasswordService,
	}
}

// @Summary Reset user password
// @Description Validate a reset token and set a new password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.ResetPasswordRequest true "Reset token and new password"
// @Success 200 {object} string "Password reset successfully"
// @Failure 400 {object} string "Invalid or expired token"
// @Failure 500 {object} string "Internal server error"
// @Router /api/v1/auth/reset-password [post]
func (h *ResetPasswordHandler) Handle(c *gin.Context) {
	var req dto.ResetPasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResponseWithError(c.Writer, http.StatusBadRequest, err.Error())
		return
	}

	err := h.resetPasswordService.Execute(c.Request.Context(), &req)
	if err != nil {
		if stdErrors.Is(err, errors.ErrorInvalidOrExpiredToken) {
			response.ResponseWithError(c.Writer, http.StatusBadRequest, errors.ErrorInvalidOrExpiredToken.Error())
			return
		}
		log.Printf("reset-password: unexpected error: %v", err)
		response.ResponseWithError(c.Writer, http.StatusInternalServerError, "internal server error")
		return
	}

	response.ResponseWithSuccess(c.Writer, http.StatusOK, "password reset successfully", nil)
}
