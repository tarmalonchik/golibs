package grpc

import "google.golang.org/grpc"

func NewGRPCServer(opts ...grpc.ServerOption) *grpc.Server {
	server := grpc.NewServer()
	return server
}
