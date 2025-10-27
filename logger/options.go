package logger

import "go.uber.org/zap"

type options struct {
	level   Level
	senders []Sender
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

type Sender func(lvl Level, msg string, fields ...zap.Field)

func WithSender(sender Sender) Opt {
	return func(o *options) {
		o.senders = append(o.senders, sender)
	}
}
