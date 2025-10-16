package logger

type options struct {
	debugMode bool
	level     Level
}
type Opt func(opt *options)

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
