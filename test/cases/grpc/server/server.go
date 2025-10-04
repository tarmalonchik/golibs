package server

import (
	"context"
	"net"
	"net/netip"

	"github.com/sirupsen/logrus"
	proto "github.com/tarmalonchik/golibs/proto/gen/go/test"
	"google.golang.org/grpc"
)

type modifier func(ctx context.Context, req *proto.EchoRequest) (*proto.EchoResponse, error)

type EchoServer struct {
	proto.UnimplementedEchoServer
	ch       chan struct{}
	modifier modifier
}

func NewGreeterServer() *EchoServer {
	return &EchoServer{
		ch: make(chan struct{}, 1),
	}
}

func (s *EchoServer) AddModifier(modifier modifier) {
	s.modifier = modifier
}

func (s *EchoServer) Run(listen netip.AddrPort) {
	lis, err := net.Listen("tcp", listen.String())
	if err != nil {
		logrus.Fatalf("failed to listen: %v", err)
	}

	srv := grpc.NewServer()
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

func (s *EchoServer) Stop() {
	s.ch <- struct{}{}
}

func (s *EchoServer) Echo(ctx context.Context, req *proto.EchoRequest) (*proto.EchoResponse, error) {
	return s.modifier(ctx, req)
}
