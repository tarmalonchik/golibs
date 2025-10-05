package client

import (
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"github.com/tarmalonchik/golibs/grpc/interceptor"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewConnection(addr string, opts ...Opt) (*grpc.ClientConn, error) {
	conf := newDefaultOptions()
	for i := range opts {
		opts[i](conf)
	}

	return grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			interceptor.NewLoggingClientInterceptor(conf.logLevel),
			retry.UnaryClientInterceptor(conf.retry...),
		),
	)
}
