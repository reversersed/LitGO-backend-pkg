package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
	shared_pb "github.com/reversersed/LitGO-proto/gen/go/shared"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func RecoveryMiddleware(c *gin.Context, err any) {
	stat, er := status.New(codes.Internal, "service recovered from panic status").WithDetails(&shared_pb.ErrorDetail{
		Description: fmt.Sprintf("%v", err),
	})
	if er != nil {
		c.Error(status.Error(codes.Internal, fmt.Sprintf("service recovered from panic status: %v | %v", err, er)))
	} else {
		c.Error(stat.Err())
	}
}
