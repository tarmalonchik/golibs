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
	proto "github.com/tarmalonchik/golibs/proto/gen/go/test"
	"github.com/tarmalonchik/golibs/test/basetest"
	"github.com/tarmalonchik/golibs/test/cases/grpc/server"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	defaultHost = "127.0.0.1:50051"
	defaultText = "text"
)

type GRPCSuite struct {
	basetest.Suite
}

func TestGRPCSuite(t *testing.T) {
	suite.Run(t, new(GRPCSuite))
}

func (s *GRPCSuite) TestExponentialBackoff() {
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

	srv := server.NewGreeterServer()
	srv.AddModifier(func(ctx context.Context, req *proto.EchoRequest) (*proto.EchoResponse, error) {
		if counter.Add(1) >= 5 {
			return &proto.EchoResponse{
				Text: req.GetText(),
			}, nil
		}
		return nil, status.Error(codes.Unavailable, "not implemented")

	})
	srv.Run(netip.MustParseAddrPort(defaultHost))
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
