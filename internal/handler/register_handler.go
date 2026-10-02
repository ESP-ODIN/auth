package handler

import (
	"errors"
	"net/http"

	"auth/internal/dto"
	"auth/internal/service"
	"auth/internal/utils/http/response"
	"github.com/gin-gonic/gin"
)

type RegisterHandler struct {
	registerService service.RegisterServiceInterface
}

func NewRegisterHandler(registerService service.RegisterServiceInterface) *RegisterHandler {
	return &RegisterHandler{
		registerService: registerService,
	}
}

// @Summary Register a new user
// @Description Register a new user
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "User registration data"
// @Success 201 {object} string "User registered successfully"
// @Failure 400 {object} string "Bad request"
// @Failure 409 {object} string "User already exists"
// @Failure 500 {object} string "Internal server error"
// @Router /api/v1/register [post]
func (h *RegisterHandler) Handle(c *gin.Context) {
	var req dto.RegisterRequest

	// 1. Validation automatique via les tags binding du DTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResponseWithError(c.Writer, http.StatusBadRequest, err.Error())
		return
	}

	// 2. Appel du service métier
	registerResponse, err := h.registerService.Execute(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserAlreadyExists):
			response.ResponseWithError(c.Writer, http.StatusConflict, err.Error())
		default:
			response.ResponseWithError(c.Writer, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	// 3. Réponse en cas de succès
	response.ResponseWithSuccess(c.Writer, http.StatusCreated, "user registered successfully", registerResponse)
}
