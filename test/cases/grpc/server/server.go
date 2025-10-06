package server

import (
	"context"
	"net"
	"net/netip"

	"github.com/sirupsen/logrus"
	"github.com/tarmalonchik/golibs/grpc"
	"github.com/tarmalonchik/golibs/grpc/interceptor"
	"github.com/tarmalonchik/golibs/grpc/middleware"
	"github.com/tarmalonchik/golibs/grpc/server"
	proto "github.com/tarmalonchik/golibs/proto/gen/go/test"
)

type modifier func(ctx context.Context, req *proto.EchoRequest) (*proto.EchoResponse, error)

type Server struct {
	proto.UnimplementedEchoServer
	ch       chan struct{}
	modifier modifier
}

func NewServer() *Server {
	return &Server{
		ch: make(chan struct{}, 1),
	}
}

func (s *Server) AddModifier(modifier modifier) {
	s.modifier = modifier
}

// nolint:noctx
func (s *Server) Run(listen netip.AddrPort, m middleware.Middleware) {
	lis, err := net.Listen("tcp", listen.String())
	if err != nil {
		logrus.Fatalf("failed to listen: %v", err)
	}

	opts := []server.Opt{
		server.WithLogLevel(grpc.LogLevelInfo),
	}
	if m != nil {
		opts = append(opts, server.WithAuth(
			interceptor.NewAuth(
				interceptor.WithBasicAuth(m),
			),
		))
	}

	srv := server.New(opts...)
	proto.RegisterEchoServer(srv, s)

	go func() {
		go func() {
			if err = srv.Serve(lis); err != nil {
				logrus.Fatalf("failed to serve: %v", err)
			}
		}()
		<-s.ch
		srv.Stop()
	}()
}

func (s *Server) Stop() {
	s.ch <- struct{}{}
}

func (s *Server) Echo(ctx context.Context, req *proto.EchoRequest) (*proto.EchoResponse, error) {
	return s.modifier(ctx, req)
}

func (s *Server) EchoAuth(ctx context.Context, req *proto.EchoRequest) (*proto.EchoResponse, error) {
	return s.modifier(ctx, req)
}

func (s *Server) EchoAuthNone(ctx context.Context, req *proto.EchoRequest) (*proto.EchoResponse, error) {
	return s.modifier(ctx, req)
}
