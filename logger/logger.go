package logger

import (
	"context"

	"go.uber.org/zap"
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
	if l.GetLevel() == LevelInfo || l.GetLevel() == LevelDebug {
		l.log, err = zap.NewDevelopment()
	} else {
		l.log, err = zap.NewProduction()
	}
	if err != nil {
		panic("create logger: " + err.Error())
	}

	l.log = l.log.WithOptions(zap.AddCallerSkip(1))

	return l
}

func (l *Logger) Close(_ context.Context) error {
	return l.log.Sync()
}

func (l *Logger) GetLevel() Level {
	return l.o.level
}

func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.log.Info(msg, fields...)
}

func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.log.Warn(msg, fields...)
}

func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.log.Error(msg, fields...)
}

func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.log.Debug(msg, fields...)
}

func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.log.Fatal("test")
	l.log.Fatal(msg, fields...)
}

func (l *Logger) AddCallerSkip(skip int) *Logger {
	return &Logger{
		log: l.log.WithOptions(zap.AddCallerSkip(skip + 1)),
		o:   l.o,
	}
}
