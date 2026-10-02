package handler

import (
	"auth/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewJWKSHandler(signer *utils.JWTSigner) gin.HandlerFunc {
	publicKeys := signer.JWKS()
	return func(c *gin.Context) { c.JSON(http.StatusOK, publicKeys) }
}
