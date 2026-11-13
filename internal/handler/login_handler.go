package handler

import (
	"auth/internal/dto"
	"auth/internal/errors"
	"auth/internal/service"
	"auth/internal/utils/http/response"
	stdErrors "errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type LoginHandler struct {
	LoginServiceImpl service.LoginServiceInterface
}

func NewLoginHandler(loginService service.LoginServiceInterface) *LoginHandler {
	return &LoginHandler{
		LoginServiceImpl: loginService,
	}
}

// @Summary Login a user
// @Description Login a user and return a JWT
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "User login data"
// @Success 200 {object} dto.LoginResponse "User logged in successfully"
// @Failure 400 {object} string "Bad request"
// @Failure 401 {object} string "Unauthorized"
// @Failure 500 {object} string "Internal server error"
// @Router /api/v1/auth/login [post]
func (h *LoginHandler) Handle(c *gin.Context) {
	var req dto.LoginRequest

	// 1. Validation automatique via les tags binding du DTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ResponseWithError(c.Writer, http.StatusBadRequest, err.Error())
		return
	}

	// 2. Appel du service métier
	// Utilisation de Login car LoginServiceInterface définit Login
	loginResponse, err := h.LoginServiceImpl.Login(c.Request.Context(), &req)
	if err != nil {
		switch {
		case stdErrors.Is(err, errors.ErrorUserNotFound):
			// Sécurité : on renvoie Unauthorized pour éviter d'énumérer les utilisateurs
			response.ResponseWithError(c.Writer, http.StatusUnauthorized, errors.ErrorInvalidCredentials.Error())
		case stdErrors.Is(err, errors.ErrorInvalidCredentials):
			response.ResponseWithError(c.Writer, http.StatusUnauthorized, err.Error())
		default:
			response.ResponseWithError(c.Writer, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	// 3. Réponse en cas de succès (200 OK pour login)
	response.ResponseWithSuccess(c.Writer, http.StatusOK, "user logged in successfully", loginResponse)
}
