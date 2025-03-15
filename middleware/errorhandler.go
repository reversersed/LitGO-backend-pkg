package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/reversersed/LitGO-proto/gen/go/shared"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// @Description General error object. This structure always returns when error occurred
type CustomError struct {
	Code      int32  `json:"code" example:"3"`                     // Internal gRPC error code (e.g. 3)
	NamedCode string `json:"type" example:"InvalidArgument"`       // Error code in string (e.g. InvalidArgument)
	Message   string `json:"message" example:"Bad token provided"` // Error message. Can be shown to users
	Details   []any  `json:"details"`                              // Error details. Check 'ErrorDetail' structure for more information
}

func (c *CustomError) Error() string {
	return c.Message
}
func ErrorHandler(c *gin.Context) {
	c.SetSameSite(http.SameSiteNoneMode)
	c.Next()

	if lastError := c.Errors.Last(); lastError != nil {
		err, valid := status.FromError(lastError.Unwrap())
		if !valid {
			custom := CustomError{
				Code:      int32(codes.Internal),
				NamedCode: codes.Internal.String(),
				Message:   lastError.Error(),
				Details:   nil,
			}
			c.JSON(http.StatusInternalServerError, custom)
		} else {
			custom := CustomError{
				Code:      err.Proto().GetCode(),
				NamedCode: err.Code().String(),
				Message:   err.Message(),
				Details:   err.Details(),
			}
			c.JSON(rpgCodeToHttpStatus(err.Code()), custom)
		}
	}
}

func rpgCodeToHttpStatus(code codes.Code) int {
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
	}
	status, ok := table[code]
	if !ok {
		return http.StatusInternalServerError
	}
	return status
}
