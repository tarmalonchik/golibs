package interceptor

import (
	"context"
	"fmt"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func NewLoggingServerInterceptor(level logrus.Level) grpc.UnaryServerInterceptor {
	return logging.UnaryServerInterceptor(wrapLogrus(logrusWithLevel(level)))
}

func NewLoggingClientInterceptor(level logrus.Level) grpc.UnaryClientInterceptor {
	return logging.UnaryClientInterceptor(wrapLogrus(logrusWithLevel(level)))
}

func logrusWithLevel(level logrus.Level) *logrus.Logger {
	l := logrus.New()
	l.SetLevel(level)
	return l
}

func wrapLogrus(l logrus.FieldLogger) logging.Logger {
	return logging.LoggerFunc(func(_ context.Context, lvl logging.Level, msg string, fields ...any) {
		f := make(map[string]any, len(fields)/2)
		i := logging.Fields(fields).Iterator()
		for i.Next() {
			k, v := i.At()
			f[k] = v
		}
		l := l.WithFields(f)

		switch lvl {
		case logging.LevelDebug:
			l.Debug(msg)
		case logging.LevelInfo:
			l.Info(msg)
		case logging.LevelWarn:
			l.Warn(msg)
		case logging.LevelError:
			l.Error(msg)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}
