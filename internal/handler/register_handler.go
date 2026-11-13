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
// @Success 201 {object} dto.RegisterResponse "User registered successfully"
// @Failure 400 {object} string "Bad request (invalid data or user under 18)"
// @Failure 409 {object} string "Email or username already exists"
// @Failure 500 {object} string "Internal server error"
// @Router /api/v1/auth/register [post]
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
		case stdErrors.Is(err, errors.ErrorEmailAlreadyExists),
			stdErrors.Is(err, errors.ErrorUsernameAlreadyExists):
			response.ResponseWithError(c.Writer, http.StatusConflict, err.Error())
		case stdErrors.Is(err, errors.ErrorUserUnderage):
			response.ResponseWithError(c.Writer, http.StatusBadRequest, err.Error())
		default:
			// L'erreur détaillée est loguée côté serveur, jamais renvoyée au client
			log.Printf("register: unexpected error: %v", err)
			response.ResponseWithError(c.Writer, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	// 3. Réponse en cas de succès
	response.ResponseWithSuccess(c.Writer, http.StatusCreated, "user registered successfully", registerResponse)
}
