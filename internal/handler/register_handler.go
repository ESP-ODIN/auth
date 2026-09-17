package handler

import (
	"errors"
	"net/http"

	"auth/internal/dto"
	"auth/internal/service"
	"auth/internal/utils"
	"github.com/gin-gonic/gin"
)

type RegisterHandler struct {
	registerService *service.RegisterService
}

func NewRegisterHandler(registerService *service.RegisterService) *RegisterHandler {
	return &RegisterHandler{
		registerService: registerService,
	}
}

func (h *RegisterHandler) Handle(c *gin.Context) {
	var req dto.RegisterRequest

	// 1. Validation automatique via les tags binding du DTO
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseWithError(c.Writer, http.StatusBadRequest, err.Error())
		return
	}

	// 2. Appel du service métier
	userResponse, err := h.registerService.Execute(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserAlreadyExists):
			utils.ResponseWithError(c.Writer, http.StatusConflict, err.Error())
		default:
			utils.ResponseWithError(c.Writer, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	// 3. Réponse en cas de succès
	utils.ResponseWithSuccess(c.Writer, http.StatusCreated, "user registered successfully", userResponse)
}
