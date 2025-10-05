package interceptor

import (
	"context"
	"errors"
	"strings"

	"github.com/tarmalonchik/golibs/grpc/middleware"
	authProto "github.com/tarmalonchik/golibs/proto/gen/go/auth"
	"github.com/tarmalonchik/golibs/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

type AuthOpt func(*authOptions)

type authOptions struct {
	middlewares map[middleware.Type]grpc.UnaryServerInterceptor
}

func WithBasicAuth(middleware middleware.Middleware) AuthOpt {
	return func(v *authOptions) {
		v.middlewares[middleware.GetType()] = middleware.GetInterceptor()
	}
}

type Auth interface {
	Interceptor(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error)
}

type auth struct {
	authOptions authOptions
}

func NewAuth(opts ...AuthOpt) Auth {
	authOpts := &authOptions{
		middlewares: make(map[middleware.Type]grpc.UnaryServerInterceptor),
	}

	for i := range opts {
		if opts[i] != nil {
			opts[i](authOpts)
		}
	}

	return &auth{authOptions: *authOpts}
}

func (a *auth) Interceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	authType, err := getAuthTypeForMethod(info.FullMethod)
	if err != nil {
		return nil, err
	}

	switch authType {
	case authProto.AuthType_NONE:
		return handler(ctx, req)
	case authProto.AuthType_BASIC:
		basic := a.authOptions.middlewares[middleware.TypeBasic]
		if basic == nil {
			return nil, status.Errorf(codes.Unauthenticated, "basic auth is unimplemented")
		}
		return basic(ctx, req, info, handler)
	default:
		return nil, status.Error(codes.PermissionDenied, "unknown auth type")
	}
}

func getAuthTypeForMethod(fullMethod string) (authProto.AuthType, error) {
	if len(fullMethod) == 0 {
		return authProto.AuthType_NONE, trace.FuncNameWithError(errors.New("empty method"))
	}

	if fullMethod[0] != '/' {
		return authProto.AuthType_NONE, trace.FuncNameWithError(errors.New("invalid method format, first symbol"))
	}

	fullMethod = fullMethod[1:]
	items := strings.Split(fullMethod, "/")
	if len(items) != 2 {
		return authProto.AuthType_NONE, trace.FuncNameWithError(errors.New("invalid method format"))
	}

	service, method := items[0], items[1]

	desc, err := protoregistry.GlobalFiles.FindDescriptorByName(protoreflect.FullName(service))
	if err != nil {
		return authProto.AuthType_NONE, trace.FuncNameWithError(errors.New("find descriptor by service name"))
	}

	svc, ok := desc.(protoreflect.ServiceDescriptor)
	if !ok {
		return authProto.AuthType_NONE, trace.FuncNameWithError(errors.New("conv svc to service descriptor"))
	}

	m := svc.Methods().ByName(protoreflect.Name(method))
	if m == nil {
		return authProto.AuthType_NONE, trace.FuncNameWithError(errors.New("method not found"))
	}

	opts, ok := m.Options().(*descriptorpb.MethodOptions)
	if !ok {
		return authProto.AuthType_NONE, trace.FuncNameWithError(errors.New("option not found"))
	}

	ext, ok := proto.GetExtension(opts, authProto.E_Auth).(*authProto.AuthOption)
	if !ok || ext == nil {
		return authProto.AuthType_NONE, nil
	}
	return ext.Type, nil
}
