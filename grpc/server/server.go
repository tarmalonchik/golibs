package server

import (
	"github.com/tarmalonchik/golibs/grpc/interceptor"
	"google.golang.org/grpc"
)

func New(opts ...Opt) *grpc.Server {
	conf := newDefaultOptions()

	for i := range opts {
		opts[i](conf)
	}

	server := grpc.NewServer(
		//grpc.UnaryInterceptor(interceptor.),
		grpc.ChainUnaryInterceptor(
			interceptor.NewLoggingServerInterceptor(conf.logLevel),
		),
	)
	return server
}
