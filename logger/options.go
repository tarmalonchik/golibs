package logger

type options struct {
	level Level
}
type Opt func(opt *options)

func WithLevel(level Level) Opt {
	return func(o *options) {
		if !level.IsValid() {
			panic("invalid level")
		}
		o.level = level
	}
}
