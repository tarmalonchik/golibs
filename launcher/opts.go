package launcher

import (
	"context"
	"time"
)

type Opt func(v *launcher)

type LFunc func(ctx context.Context) error

func WithLogger(logger Logger) Opt {
	return func(v *launcher) {
		v.logger = logger
	}
}

func WithRepeaterPeriod(duration time.Duration) Opt {
	return func(v *launcher) {
		v.repeaterPeriod = 1 * time.Second
	}
}

func WithTimeout(timeout time.Duration) Opt {
	return func(v *launcher) {
		v.timeout = timeout
	}
}

func WithRunnersCount(count int64) Opt {
	return func(v *launcher) {
		v.parallelCount = count
	}
}
