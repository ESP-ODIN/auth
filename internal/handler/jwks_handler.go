package handler

import (
	"auth/internal/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

func NewJWKSHandler(signer *utils.JWTSigner) gin.HandlerFunc {
	publicKeys := signer.JWKS()
	return func(c *gin.Context) { c.JSON(http.StatusOK, publicKeys) }
}
