package handler

import (
	"auth/internal/dto"
	"auth/internal/service"
	"auth/internal/utils/http/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ForgotPasswordHandler struct {
	forgotPasswordService service.ForgotPasswordServiceInterface
}

func NewForgotPasswordHandler(forgotPasswordService service.ForgotPasswordServiceInterface) *ForgotPasswordHandler {
	return &ForgotPasswordHandler{
		forgotPasswordService: forgotPasswordService,
	}
}

// @Summary Request a password reset
// @Description Send a password reset link to the given email if it exists
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.ForgotPasswordRequest true "Email address"
// @Success 202 {object} string "Reset link sent if email is registered"
// @Failure 400 {object} string "Bad request"
// @Router /api/v1/auth/forgot-password [post]
func (h *ForgotPasswordHandler) Handle(c *gin.Context) {
	var req dto.ForgotPasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResponseWithError(c.Writer, http.StatusBadRequest, err.Error())
		return
	}

	_ = h.forgotPasswordService.Execute(c.Request.Context(), &req)

	response.ResponseWithSuccess(c.Writer, http.StatusAccepted, "if this email is registered, a reset link has been sent", nil)
}
