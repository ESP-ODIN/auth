package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

func TestRateLimiter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Allow 1 request per second, burst 1
	rl := NewRateLimiter(rate.Every(1*time.Second), 1)

	router := gin.New()
	router.POST("/test", rl.Limit(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// First request should pass
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/test", nil)
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Second request immediately after should be rate limited
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/test", nil)
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
}
