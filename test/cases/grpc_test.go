package cases

import (
	"context"
	"net/netip"
	"sync/atomic"
	"testing"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"github.com/stretchr/testify/suite"
	"github.com/tarmalonchik/golibs/grpc"
	"github.com/tarmalonchik/golibs/grpc/client"
	"github.com/tarmalonchik/golibs/grpc/middleware"
	proto "github.com/tarmalonchik/golibs/proto/gen/go/test"
	"github.com/tarmalonchik/golibs/test/basetest"
	"github.com/tarmalonchik/golibs/test/cases/grpc/server"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	defaultHost = "127.0.0.1:50051"
	defaultText = "text"
	defaultUser = "user"
	defaultPass = "pass"
)

type GRPCSuite struct {
	basetest.Suite
}

func TestGRPCSuite(t *testing.T) {
	suite.Run(t, new(GRPCSuite))
}

func (s *GRPCSuite) TestExponentialBackoffNoAuth() {
	ctx := context.Background()

	conn, err := client.NewConnection(
		defaultHost,
		client.WithLogLevel(grpc.LogLevelInfo),
		client.WithRetryMax(20),
		client.WithRetryCodes(codes.Unavailable, codes.ResourceExhausted, codes.DeadlineExceeded),
		client.WithRetryBackoff(retry.BackoffExponential(200*time.Millisecond)),
	)
	s.Require().NoError(err)

	defer func() {
		_ = conn.Close()
	}()

	cli := proto.NewEchoClient(conn)

	counter := atomic.Uint32{}

	srv := server.NewServer()
	srv.AddModifier(func(ctx context.Context, req *proto.EchoRequest) (*proto.EchoResponse, error) {
		if counter.Add(1) >= 5 {
			return &proto.EchoResponse{
				Text: req.GetText(),
			}, nil
		}
		return nil, status.Error(codes.Unavailable, "not implemented")

	})
	srv.Run(netip.MustParseAddrPort(defaultHost), nil)
	defer srv.Stop()

	ch := make(chan struct{})

	go func() {
		resp, err := cli.Echo(ctx, &proto.EchoRequest{
			Text: defaultText,
		})
		s.Require().NoError(err)
		s.Require().Equal(defaultText, resp.GetText())
		ch <- struct{}{}
	}()

	newCtx, c := context.WithTimeout(ctx, time.Second*10)
	defer c()

	select {
	case <-newCtx.Done():
		s.Fail(newCtx.Err().Error())
	case <-ch:
	}
}

func (s *GRPCSuite) TestProtoAuthAbsentServerAuthEnabled() {
	ctx := context.Background()

	conn, err := client.NewConnection(defaultHost, client.WithLogLevel(grpc.LogLevelInfo))
	s.Require().NoError(err)

	defer func() {
		_ = conn.Close()
	}()

	cli := proto.NewEchoClient(conn)

	srv := server.NewServer()
	srv.AddModifier(func(ctx context.Context, req *proto.EchoRequest) (*proto.EchoResponse, error) {
		return &proto.EchoResponse{Text: req.GetText()}, nil

	})
	srv.Run(netip.MustParseAddrPort(defaultHost), middleware.NewBasicMiddleware(defaultUser, defaultPass))
	defer srv.Stop()

	ch := make(chan struct{})

	go func() {
		resp, err := cli.Echo(ctx, &proto.EchoRequest{
			Text: defaultText,
		})
		s.Require().NoError(err)
		s.Require().Equal(defaultText, resp.GetText())
		ch <- struct{}{}
	}()

	newCtx, c := context.WithTimeout(ctx, time.Second*10)
	defer c()

	select {
	case <-newCtx.Done():
		s.Fail(newCtx.Err().Error())
	case <-ch:
	}
}

func (s *GRPCSuite) TestProtoAuthNoneServerAuthEnabled() {
	ctx := context.Background()

	conn, err := client.NewConnection(defaultHost, client.WithLogLevel(grpc.LogLevelInfo))
	s.Require().NoError(err)

	defer func() {
		_ = conn.Close()
	}()

	cli := proto.NewEchoClient(conn)

	srv := server.NewServer()
	srv.AddModifier(func(ctx context.Context, req *proto.EchoRequest) (*proto.EchoResponse, error) {
		return &proto.EchoResponse{Text: req.GetText()}, nil

	})
	srv.Run(netip.MustParseAddrPort(defaultHost), middleware.NewBasicMiddleware(defaultUser, defaultPass))
	defer srv.Stop()

	ch := make(chan struct{})

	go func() {
		resp, err := cli.EchoAuthNone(ctx, &proto.EchoRequest{
			Text: defaultText,
		})
		s.Require().NoError(err)
		s.Require().Equal(defaultText, resp.GetText())
		ch <- struct{}{}
	}()

	newCtx, c := context.WithTimeout(ctx, time.Second*10)
	defer c()

	select {
	case <-newCtx.Done():
		s.Fail(newCtx.Err().Error())
	case <-ch:
	}
}

func (s *GRPCSuite) TestProtoAuthBasicServerAuthEnabled() {
	ctx := context.Background()

	conn, err := client.NewConnection(defaultHost, client.WithLogLevel(grpc.LogLevelInfo))
	s.Require().NoError(err)

	defer func() {
		_ = conn.Close()
	}()

	cli := proto.NewEchoClient(conn)

	srv := server.NewServer()
	srv.AddModifier(func(ctx context.Context, req *proto.EchoRequest) (*proto.EchoResponse, error) {
		return &proto.EchoResponse{Text: req.GetText()}, nil
	})
	srv.Run(netip.MustParseAddrPort(defaultHost), middleware.NewBasicMiddleware(defaultUser, defaultPass))
	defer srv.Stop()

	ch := make(chan struct{})

	go func() {
		_, err = cli.EchoAuth(ctx, &proto.EchoRequest{
			Text: defaultText,
		})
		s.Require().Error(err)
		rs, ok := status.FromError(err)
		s.Require().True(ok)
		s.Require().Equal(codes.Unauthenticated, rs.Code())
		ch <- struct{}{}
	}()

	newCtx, c := context.WithTimeout(ctx, time.Second*10)
	defer c()

	select {
	case <-newCtx.Done():
		s.Fail(newCtx.Err().Error())
	case <-ch:
	}
}

func (s *GRPCSuite) TestProtoAuthBasicServerAuthEnabledSuccess() {
	ctx := context.Background()

	conn, err := client.NewConnection(defaultHost, client.WithLogLevel(grpc.LogLevelInfo))
	s.Require().NoError(err)

	defer func() {
		_ = conn.Close()
	}()

	cli := proto.NewEchoClient(conn)

	srv := server.NewServer()
	srv.AddModifier(func(ctx context.Context, req *proto.EchoRequest) (*proto.EchoResponse, error) {
		return &proto.EchoResponse{Text: req.GetText()}, nil
	})
	srv.Run(netip.MustParseAddrPort(defaultHost), middleware.NewBasicMiddleware(defaultUser, defaultPass))
	defer srv.Stop()

	ch := make(chan struct{})

	go func() {
		_, err = cli.EchoAuth(
			middleware.InjectBasic(ctx, defaultUser, defaultPass),
			&proto.EchoRequest{
				Text: defaultText,
			})
		s.Require().NoError(err)
		ch <- struct{}{}
	}()

	newCtx, c := context.WithTimeout(ctx, time.Second*10)
	defer c()

	select {
	case <-newCtx.Done():
		s.Fail(newCtx.Err().Error())
	case <-ch:
	}
}
