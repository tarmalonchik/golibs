package server

import (
	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/tarmalonchik/golibs/grpc/interceptor"
	"google.golang.org/grpc"
)

func New(opts ...Opt) *grpc.Server {
	conf := newDefaultOptions()

	for i := range opts {
		opts[i](conf)
	}

	serverOpts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			grpcmiddleware.ChainUnaryServer(conf.interceptors...),
			interceptor.NewLoggingServerInterceptor(conf.logLevel),
		),
	}

	if conf.auth != nil {
		serverOpts = append(serverOpts, grpc.UnaryInterceptor(conf.auth.Interceptor))
	}

	return grpc.NewServer(serverOpts...)
}
