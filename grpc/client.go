package grpc

import (
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"

	"time"
)

func NewGRPCConnection(addr string) (*grpc.ClientConn, error) {
	logger, _ := zap.NewProduction()

	retryOpts := []retry.CallOption{
		retry.WithMax(20),
		retry.WithBackoff(retry.BackoffExponential(200 * time.Millisecond)),
		retry.WithCodes(codes.Unavailable, codes.ResourceExhausted, codes.DeadlineExceeded),
	}

	return grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			logging.UnaryClientInterceptor(InterceptorLogger(logger)),
			retry.UnaryClientInterceptor(retryOpts...),
		),
	)
}
