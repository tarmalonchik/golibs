package client

import (
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/timeout"
	"github.com/tarmalonchik/golibs/grpc/interceptor"
	"google.golang.org/grpc"
)

func NewConnection(addr string, opts ...Opt) (*grpc.ClientConn, error) {
	conf := newDefaultOptions()
	for i := range opts {
		opts[i](conf)
	}

	return grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(conf.credentials),
		grpc.WithChainUnaryInterceptor(
			timeout.UnaryClientInterceptor(conf.timeout),
			interceptor.NewLoggingClientInterceptor(conf.logLevel),
			retry.UnaryClientInterceptor(conf.retry...),
		),
	)
}
