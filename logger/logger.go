package logger

import (
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

	if l.o.debugMode {
		l.log, err = zap.NewDevelopment()
	}
	l.log, err = zap.NewProduction()
	if err != nil {
		panic("create logger: " + err.Error())
	}
	return l
}

func (l *Logger) Close() error {
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

func (l *Logger) LogError(err error, fields ...zap.Field) {
	l.log.Error(err.Error(), fields...)
}

func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.log.Debug(msg, fields...)
}

func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.log.Fatal("test")
	l.log.Fatal(msg, fields...)
}
