package middleware

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCodeToStatus(t *testing.T) {
	table := map[codes.Code]int{
		codes.OK:                 http.StatusOK,
		codes.Canceled:           http.StatusGone,
		codes.Unknown:            http.StatusInternalServerError,
		codes.InvalidArgument:    http.StatusBadRequest,
		codes.DeadlineExceeded:   http.StatusGatewayTimeout,
		codes.NotFound:           http.StatusNotFound,
		codes.AlreadyExists:      http.StatusConflict,
		codes.PermissionDenied:   http.StatusForbidden,
		codes.ResourceExhausted:  http.StatusTooManyRequests,
		codes.FailedPrecondition: http.StatusBadRequest,
		codes.Aborted:            http.StatusConflict,
		codes.OutOfRange:         http.StatusBadRequest,
		codes.Unimplemented:      http.StatusNotImplemented,
		codes.Internal:           http.StatusInternalServerError,
		codes.Unavailable:        http.StatusServiceUnavailable,
		codes.DataLoss:           http.StatusInternalServerError,
		codes.Unauthenticated:    http.StatusUnauthorized,
		codes.Code(99999):        http.StatusInternalServerError,
	}
	for value, excepted := range table {
		t.Run(fmt.Sprintf("%s_Test_Code", value.String()), func(t *testing.T) {
			assert.Equal(t, excepted, rpgCodeToHttpStatus(value))
		})
	}
}
func TestErrorHandler(t *testing.T) {
	table := []struct {
		Name          string
		Endpoint      func(*gin.Context)
		ExceptedCode  int
		ExceptedError error
	}{
		{
			Name: "Default error thrown",
			Endpoint: func(ctx *gin.Context) {
				ctx.Error(errors.New("something wrong happened"))
			},
			ExceptedCode: http.StatusInternalServerError,
			ExceptedError: &CustomError{
				Code:      int32(codes.Internal),
				NamedCode: codes.Internal.String(),
				Message:   "something wrong happened",
				Details:   nil,
			},
		},
		{
			Name: "Custom not found error thrown",
			Endpoint: func(ctx *gin.Context) {
				ctx.Error(status.Error(codes.NotFound, "not found"))
			},
			ExceptedCode: http.StatusNotFound,
			ExceptedError: &CustomError{
				Code:      int32(codes.NotFound),
				NamedCode: codes.NotFound.String(),
				Message:   "not found",
				Details:   nil,
			},
		},
		{
			Name: "Wrong error code received",
			Endpoint: func(ctx *gin.Context) {
				ctx.Error(status.Error(codes.Code(99999), "wrong code"))
			},
			ExceptedCode: http.StatusInternalServerError,
			ExceptedError: &CustomError{
				Code:      int32(codes.Internal),
				NamedCode: codes.Internal.String(),
				Message:   "wrong code",
				Details:   nil,
			},
		},
	}
	for _, v := range table {
		t.Run(v.Name, func(t *testing.T) {
			gin.SetMode(gin.ReleaseMode)
			engine := gin.Default()
			engine.Use(ErrorHandler)
			engine.GET("/", v.Endpoint)

			r := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			engine.ServeHTTP(w, r)
			assert.Equal(t, v.ExceptedCode, w.Result().StatusCode)
			var err *CustomError = &CustomError{}
			errs := json.NewDecoder(w.Result().Body).Decode(err)
			assert.NoError(t, errs)
			assert.EqualError(t, err, v.ExceptedError.Error())
		})
	}
}
