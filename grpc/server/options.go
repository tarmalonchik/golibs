package server

import (
	"github.com/sirupsen/logrus"
	"github.com/tarmalonchik/golibs/grpc"
	"github.com/tarmalonchik/golibs/grpc/interceptor"
)

type Opt func(*options)

type options struct {
	logLevel logrus.Level
	auth     interceptor.Auth
}

func newDefaultOptions() *options {
	return &options{
		logLevel: logrus.ErrorLevel,
	}
}

func WithLogLevel(lvl grpc.LogLevel) Opt {
	return func(v *options) {
		v.logLevel = lvl.LogrusLevel()
	}
}

func WithAuth(auth *interceptor.Auth) Opt {
	return func(v *options) {
		v.auth = *auth
	}
}
