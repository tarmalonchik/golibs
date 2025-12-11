package logger

import (
	"context"

	"github.com/tarmalonchik/golibs/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	log *zap.Logger
	o   *options
}

func NewLogger(opts ...Opt) *Logger {
	l := &Logger{
		o: &options{},
	}
	for _, opt := range opts {
		opt(l.o)
	}

	var err error
	var cfg zap.Config

	if l.GetLevel() == LevelInfo || l.GetLevel() == LevelDebug || l.GetLevel() == LevelWarn {
		cfg = zap.NewDevelopmentConfig()
	} else {
		cfg = zap.NewProductionConfig()
	}

	zapLvl, err := zapcore.ParseLevel(l.o.level.String())
	if err != nil {
		panic("create logger lvl: " + err.Error())
	}

	cfg.Level = zap.NewAtomicLevel()
	cfg.Level.SetLevel(zapLvl)

	if l.log, err = cfg.Build(); err != nil {
		panic("build logger: " + err.Error())
	}

	l.log = l.log.WithOptions(zap.AddCallerSkip(1))

	return l
}

func (l *Logger) Close(_ context.Context) error {
	_ = l.log.Sync()
	return nil
}

func (l *Logger) GetLevel() Level {
	return l.o.level
}

func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.log.Info(msg, fields...)
	l.runSenders(LevelInfo, msg, fields...)
}

func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.log.Warn(msg, fields...)
	l.runSenders(LevelWarn, msg, fields...)
}

func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.log.Error(msg, fields...)
	fields = append(fields, zap.String("stacktrace", trace.FuncNameWithSkip(3).Error()))
	l.runSenders(LevelError, msg, fields...)
}

func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.log.Debug(msg, fields...)
	l.runSenders(LevelDebug, msg, fields...)
}

func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.log.Fatal(msg, fields...)
	l.runSenders(LevelFatal, msg, fields...)
}

func (l *Logger) runSenders(lvl Level, msg string, fields ...zap.Field) {
	go func() {
		for i := range l.o.senders {
			l.o.senders[i](lvl, msg, fields...)
		}
	}()
}
