package interceptor

import (
	"context"
	"path"

	"github.com/tarmalonchik/golibs/grpc/middleware"
	proto "github.com/tarmalonchik/golibs/proto/gen/go/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

type AuthOpt func(*authOptions)

type authOptions struct {
	middlewares map[middleware.MiddlewareType]grpc.UnaryServerInterceptor
}

func WithBasicAuth(middleware middleware.Middleware) AuthOpt {
	return func(v *authOptions) {
		v.middlewares[middleware.GetType()] = middleware.GetInterceptor()
	}
}

type Auth struct {
	authOptions authOptions
}

func NewAuth(opts ...AuthOpt) *Auth {
	auth := &authOptions{}

	for i := range opts {
		opts[i](auth)
	}

	return &Auth{authOptions: *auth}
}

func (a *Auth) Interceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	authType := getAuthTypeForMethod(path.Base(info.FullMethod))

	switch authType {
	case proto.AuthType_NONE:
		return handler(ctx, req)
	case proto.AuthType_BASIC:
		basic := a.authOptions.middlewares[middleware.MiddlewareTypeBasic]
		if basic == nil {
			return nil, status.Errorf(codes.Unauthenticated, "basic auth is unimplemented")
		}
		return basic(ctx, req, info, handler)
	default:
		return nil, status.Error(codes.PermissionDenied, "unknown auth type")
	}
}

func getAuthTypeForMethod(method string) proto.AuthType {
	svcDesc, err := protoregistry.GlobalFiles.FindDescriptorByName("hello.Greeter")
	if err != nil {
		return proto.AuthType_NONE
	}
	svc := svcDesc.(protoreflect.ServiceDescriptor)
	m := svc.Methods().ByName(protoreflect.Name(method))
	if m == nil {
		return proto.AuthType_NONE
	}

	opts := m.Options().(*descriptorpb.MethodOptions)
	ext := proto.GetExtension(opts, authpb.E_Auth).(*authpb.AuthOption)
	return ext.GetType()
}
