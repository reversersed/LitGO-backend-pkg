package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRecoveryMiddleware(t *testing.T) {
	handler := func(*gin.Context) {
		panic("panic message")
	}

	router := gin.Default()
	router.Use(ErrorHandler)
	router.Use(gin.CustomRecovery(RecoveryMiddleware))
	router.GET("/", handler)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	router.ServeHTTP(w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Result().StatusCode)

	b, err := io.ReadAll(w.Result().Body)
	if assert.Nil(t, err) {
		assert.Equal(t, string(b), "{\"code\":13,\"type\":\"Internal\",\"message\":\"service recovered from panic status\",\"details\":[{\"description\":\"panic message\"}]}")
	}
}
