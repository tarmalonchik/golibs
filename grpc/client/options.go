package client

import (
	"crypto/tls"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"github.com/sirupsen/logrus"
	"github.com/tarmalonchik/golibs/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type Opt func(*options)

type options struct {
	logLevel        logrus.Level
	retry           []retry.CallOption
	timeout         time.Duration
	perRetryTimeout time.Duration
	credentials     credentials.TransportCredentials
}

func newDefaultOptions() *options {
	return &options{
		logLevel:        logrus.ErrorLevel,
		timeout:         time.Second * 20,
		perRetryTimeout: time.Second * 5,
		credentials:     insecure.NewCredentials(),
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

func WithPerRetryTimeout(timeout time.Duration) Opt {
	return func(v *options) {
		v.retry = append(v.retry, retry.WithPerRetryTimeout(timeout))
	}
}

func WithTLS() Opt {
	return func(v *options) {
		v.credentials = credentials.NewTLS(&tls.Config{
			MinVersion: tls.VersionTLS12,
		})
	}
}
