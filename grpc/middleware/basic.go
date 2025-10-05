package middleware

import (
	"context"
	"encoding/base64"
	"strings"

	proto "github.com/tarmalonchik/golibs/proto/gen/go/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

type Middleware struct {
	interceptor    grpc.UnaryServerInterceptor
	middlewareType MiddlewareType
}

func (m *Middleware) GetInterceptor() grpc.UnaryServerInterceptor {
	return m.interceptor
}

func (m *Middleware) GetType() MiddlewareType {
	return m.middlewareType
}

func NewBasicMiddleware() *Middleware {
	return &Middleware{
		interceptor:    basic,
		middlewareType: MiddlewareTypeBasic,
	}
}

func basic(
	ctx context.Context,
	req interface{},
	_ *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "metadata not provided")
	}

	values := md["authorization"]
	if len(values) == 0 {
		return nil, status.Error(codes.Unauthenticated, "authorization header missing")
	}

	auth := values[0]
	const prefix = "Basic "
	if !strings.HasPrefix(auth, prefix) {
		return nil, status.Error(codes.Unauthenticated, "invalid auth scheme")
	}

	payload, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(auth, prefix))
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "malformed base64")
	}

	parts := strings.SplitN(string(payload), ":", 2)
	if len(parts) != 2 {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials format")
	}

	user, pass := parts[0], parts[1]
	if !validateUser(user, pass) {
		return nil, status.Error(codes.PermissionDenied, "invalid username or password")
	}

	// всё ок, добавляем пользователя в контекст
	ctx = context.WithValue(ctx, "user", user)
	return handler(ctx, req)
}
