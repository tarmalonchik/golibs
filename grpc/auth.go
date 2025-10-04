package grpc

import (
	"context"
	"path"

	proto "github.com/tarmalonchik/golibs/proto/gen/go/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

func authInterceptor(
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
		return handleBasicAuth(ctx, req, info, handler)
	case proto.AuthType_JWT:
		return handleJWTAuth(ctx, req, info, handler)
	case proto.AuthType_ADMIN_ONLY:
		return handleAdminAuth(ctx, req, info, handler)
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
