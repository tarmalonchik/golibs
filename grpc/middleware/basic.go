package middleware

import (
	"context"
	"encoding/base64"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	basicPrefix = "Basic "
)

type BasicMiddleware struct {
	interceptor    grpc.UnaryServerInterceptor
	middlewareType Type
}

func (m *BasicMiddleware) GetInterceptor() grpc.UnaryServerInterceptor {
	return m.interceptor
}

func (m *BasicMiddleware) GetType() Type {
	return m.middlewareType
}

func NewBasicMiddleware(user, password string) Middleware {
	g := gag{
		user:     user,
		password: password,
	}

	return &BasicMiddleware{
		interceptor:    g.basic,
		middlewareType: TypeBasic,
	}
}

type gag struct {
	user, password string
}

func (g *gag) basic(
	ctx context.Context,
	req interface{},
	_ *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "metadata not provided")
	}

	values := md[ContextKeyAuthorization.String()]
	if len(values) == 0 {
		return nil, status.Error(codes.Unauthenticated, "authorization header missing")
	}

	auth := values[0]
	if !strings.HasPrefix(auth, basicPrefix) {
		return nil, status.Error(codes.Unauthenticated, "invalid auth scheme")
	}

	payload, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(auth, basicPrefix))
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "malformed base64")
	}

	parts := strings.SplitN(string(payload), ":", 2)
	if len(parts) != 2 {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials format")
	}
	if len(parts[0]) == 0 || len(parts[1]) == 0 {
		return nil, status.Error(codes.Unauthenticated, "authorization header missing")
	}

	if cryptCompare(g.user, parts[0]) && cryptCompare(g.password, parts[1]) {
		ctx = context.WithValue(ctx, ContextKeyUsername, g.user)
		return handler(ctx, req)
	}
	return nil, status.Error(codes.PermissionDenied, "invalid username or password")
}

func InjectBasic(ctx context.Context, user, password string) context.Context {
	token := base64.StdEncoding.EncodeToString([]byte(user + ":" + password))
	return metadata.AppendToOutgoingContext(ctx, ContextKeyAuthorization.String(), basicPrefix+token)
}
