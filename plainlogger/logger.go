package plainlogger

import (
	"github.com/sirupsen/logrus"

	"github.com/tarmalonchik/golibs/trace"
)

type Config struct {
	LogLevel string `envconfig:"LOG_LEVEL" default:"error"`
}

type Logger struct {
	log *logrus.Logger
}

func NewLogger(conf Config) *Logger {
	l := &Logger{
		log: logrus.New(),
	}
	lvl, err := logrus.ParseLevel(conf.LogLevel)
	if err != nil {
		lvl = logrus.ErrorLevel
	}
	l.log.SetLevel(lvl)
	return l
}

func (l *Logger) GetLevel() logrus.Level {
	return l.log.Level
}

func (l *Logger) Info(args ...any) {
	box := make([]any, 0, len(args)+1)
	box = append(box, trace.FuncNameAndLineLogger())
	box = append(box, args...)
	l.log.Info(box)
}

func (l *Logger) Warn(args ...any) {
	box := make([]any, 0, len(args)+1)
	box = append(box, trace.FuncNameAndLineLogger())
	box = append(box, args...)
	l.log.Warn(box)
}

func (l *Logger) Error(args ...any) {
	box := make([]any, 0, len(args)+1)
	box = append(box, trace.FuncNameAndLineLogger())
	box = append(box, args...)
	l.log.Error(box)
}

func (l *Logger) Debug(args ...any) {
	box := make([]any, 0, len(args)+1)
	box = append(box, trace.FuncNameAndLineLogger())
	box = append(box, args...)
	l.log.Debug(box)
}

func (l *Logger) Trace(args ...any) {
	box := make([]any, 0, len(args)+1)
	box = append(box, trace.FuncNameAndLineLogger())
	box = append(box, args...)
	l.log.Trace(box)
}

func (l *Logger) Fatal(args ...any) {
	box := make([]any, 0, len(args)+1)
	box = append(box, trace.FuncNameAndLineLogger())
	box = append(box, args...)
	l.log.Fatal(box)
}

func (l *Logger) Errorf(format string, args ...any) {
	box := make([]any, 0, len(args)+1)
	box = append(box, trace.FuncNameAndLineLogger())
	box = append(box, args...)
	l.log.Errorf("%v "+format, box...)
}

func (l *Logger) Warnf(format string, args ...any) {
	box := make([]any, 0, len(args)+1)
	box = append(box, trace.FuncNameAndLineLogger())
	box = append(box, args...)
	l.log.Warnf("%v "+format, box...)
}

func (l *Logger) Infof(format string, args ...any) {
	box := make([]any, 0, len(args)+1)
	box = append(box, trace.FuncNameAndLineLogger())
	box = append(box, args...)
	l.log.Infof("%v "+format, box...)
}

func (l *Logger) Debugf(format string, args ...any) {
	box := make([]any, 0, len(args)+1)
	box = append(box, trace.FuncNameAndLineLogger())
	box = append(box, args...)
	l.log.Debugf("%v "+format, box...)
}

func (l *Logger) Fatalf(format string, args ...any) {
	box := make([]any, 0, len(args)+1)
	box = append(box, trace.FuncNameAndLineLogger())
	box = append(box, args...)
	l.log.Fatalf("%v "+format, box...)
}

func (l *Logger) With(fields map[string]any) *Entry {
	return &Entry{
		entry: l.log.WithFields(fields),
	}
}

type Entry struct {
	entry *logrus.Entry
}

func (l *Entry) GetLevel() logrus.Level {
	return l.entry.Level
}

func (l *Entry) Info(args ...any) {
	box := make([]any, 0, len(args)+1)
	box = append(box, trace.FuncNameAndLineLogger())
	box = append(box, args...)
	l.entry.Info(box)
}

func (l *Entry) Warn(args ...any) {
	box := make([]any, 0, len(args)+1)
	box = append(box, trace.FuncNameAndLineLogger())
	box = append(box, args...)
	l.entry.Warn(box)
}

func (l *Entry) Error(args ...any) {
	box := make([]any, 0, len(args)+1)
	box = append(box, trace.FuncNameAndLineLogger())
	box = append(box, args...)
	l.entry.Error(box)
}

func (l *Entry) Debug(args ...any) {
	box := make([]any, 0, len(args)+1)
	box = append(box, trace.FuncNameAndLineLogger())
	box = append(box, args...)
	l.entry.Debug(box)
}

func (l *Entry) Fatal(args ...any) {
	box := make([]any, 0, len(args)+1)
	box = append(box, trace.FuncNameAndLineLogger())
	box = append(box, args...)
	l.entry.Fatal(box)
}

func (l *Entry) Errorf(format string, args ...any) {
	box := make([]any, 0, len(args)+1)
	box = append(box, trace.FuncNameAndLineLogger())
	box = append(box, args...)
	l.entry.Errorf("%v "+format, box...)
}

func (l *Entry) Warnf(format string, args ...any) {
	box := make([]any, 0, len(args)+1)
	box = append(box, trace.FuncNameAndLineLogger())
	box = append(box, args...)
	l.entry.Warnf("%v "+format, box...)
}

func (l *Entry) Infof(format string, args ...any) {
	box := make([]any, 0, len(args)+1)
	box = append(box, trace.FuncNameAndLineLogger())
	box = append(box, args...)
	l.entry.Infof("%v "+format, box...)
}

func (l *Entry) Debugf(format string, args ...any) {
	box := make([]any, 0, len(args)+1)
	box = append(box, trace.FuncNameAndLineLogger())
	box = append(box, args...)
	l.entry.Debugf("%v "+format, box...)
}

func (l *Entry) Fatalf(format string, args ...any) {
	box := make([]any, 0, len(args)+1)
	box = append(box, trace.FuncNameAndLineLogger())
	box = append(box, args...)
	l.entry.Fatalf("%v "+format, box...)
}

func (l *Entry) Trace(args ...any) {
	box := make([]any, 0, len(args)+1)
	box = append(box, trace.FuncNameAndLineLogger())
	box = append(box, args...)
	l.entry.Trace(box)
}

func (l *Entry) With(fields map[string]any) *Entry {
	return &Entry{
		entry: l.entry.WithFields(fields),
	}
}
