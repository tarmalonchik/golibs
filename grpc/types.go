//go:generate go-enum -f=$GOFILE --nocase --values
package grpc

import (
	"errors"

	"github.com/sirupsen/logrus"
	"github.com/tarmalonchik/golibs/trace"
)

// LogLevel
// ENUM(
// panic
// fatal
// error
// warn
// info
// debug
// trace
// )
type LogLevel string

func (l LogLevel) LogrusLevel() logrus.Level {
	lvl, err := logrus.ParseLevel(l.String())
	if err != nil {
		panic(trace.FuncNameWithError(errors.New("invalid log level")))
	}
	return lvl
}
