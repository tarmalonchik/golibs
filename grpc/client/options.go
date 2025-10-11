package client

import (
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"github.com/sirupsen/logrus"
	"github.com/tarmalonchik/golibs/grpc"
	"google.golang.org/grpc/codes"
)

type Opt func(*options)

type options struct {
	logLevel logrus.Level
	retry    []retry.CallOption
	timeout  time.Duration
}

func newDefaultOptions() *options {
	return &options{
		logLevel: logrus.ErrorLevel,
		timeout:  time.Second * 20,
	}
}

func WithLogLevel(lvl grpc.LogLevel) Opt {
	return func(v *options) {
		v.logLevel = lvl.LogrusLevel()
	}
}

func WithRetryMax(maxRetries uint) Opt {
	return func(v *options) {
		v.retry = append(v.retry, retry.WithMax(maxRetries))
	}
}

func WithRetryBackoff(bf retry.BackoffFunc) Opt {
	return func(v *options) {
		v.retry = append(v.retry, retry.WithBackoff(bf))
	}
}

func WithRetryCodes(codes ...codes.Code) Opt {
	return func(v *options) {
		v.retry = append(v.retry, retry.WithCodes(codes...))
	}
}

func WithTimeout(timeout time.Duration) Opt {
	return func(v *options) {
		v.timeout = timeout
	}
}
