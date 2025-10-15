package logger

type options struct {
	debugMode  bool
	level      Level
	ignoreFunc func(err error) bool
}
type Opt func(opt *options)

func WithIgnoreFunc(ignoreFunc func(err error) bool) Opt {
	return func(opt *options) {
		opt.ignoreFunc = ignoreFunc
	}
}

func WithLevel(level Level) Opt {
	return func(o *options) {
		o.level = level
	}
}

func WithDebugMode() Opt {
	return func(o *options) {
		o.debugMode = true
	}
}
