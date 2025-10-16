package launcher

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"go.uber.org/zap"
)

type Logger interface {
	Error(msg string, fields ...zap.Field)
}

type Launcher interface {
	Launch(ctx context.Context)
	AddRunner(in RunnerFunc, opts ...RunnerOpt)
}

type launcher struct {
	runnableCount  int64
	finishersCount int64
	parallelCount  int64

	runnableStack  chan *runner // stack with jobs need to run
	finishersStack chan *runner // stack with jobs run when finish

	errorChan chan error         // stack with errors
	logger    Logger             // you can use this logger for custom logging
	cancel    context.CancelFunc // context.Cancel func
	timeout   time.Duration      // time when forced termination will happen after crushing
	jobsDone  chan interface{}   // the channel to signal when all work done
}

func NewLauncher(opts ...Opt) Launcher {
	l := &launcher{}

	for i := range opts {
		opts[i](l)
	}

	if l.logger == nil {
		l.logger = zap.NewNop()
	}

	if l.timeout == 0 {
		l.timeout = 5 * time.Second
	}

	if l.parallelCount == 0 {
		l.parallelCount = 100
	}

	l.runnableStack = make(chan *runner, l.parallelCount)
	l.finishersStack = make(chan *runner, l.parallelCount)
	l.errorChan = make(chan error, l.parallelCount)
	l.jobsDone = make(chan interface{})

	return l
}

func (c *launcher) AddRunner(in RunnerFunc, opts ...RunnerOpt) {
	r := &runner{}
	for i := range opts {
		opts[i](r)
	}

	if r.isFinisher && (r.repeatOnFinish || r.repeatOnPanic || r.repeatOnError) {
		panic("finisher should have no repeater")
	}

	if r.isFinisher {
		c.finishersStack <- newRunner(in, opts...)
		c.incFinishers()
	} else {
		c.runnableStack <- newRunner(in, opts...)
		c.inc()
	}
}

func (c *launcher) Launch(ctx context.Context) {
	originalContext := ctx
	ctx, cancel := context.WithCancel(ctx)
	c.cancel = cancel

	go c.waitForInterruption()
	go c.logErrors(ctx)
	go c.runRunners(originalContext, ctx)

	<-ctx.Done()

	c.waitGraceful()
}

func (c *launcher) runRunners(originalContext, ctx context.Context) {
	for item := range c.runnableStack {
		if ctx.Err() != nil {
			close(c.runnableStack)
			continue
		}
		go func(*runner) {
			err, wasPanic := item.run(ctx)
			if err != nil {
				c.errorChan <- err
			}

			if wasPanic && item.repeatOnPanic {
				time.Sleep(100 * time.Millisecond)
				c.runnableStack <- item
				return
			}

			if err != nil && !wasPanic && item.repeatOnError {
				time.Sleep(100 * time.Millisecond)
				c.runnableStack <- item
				return
			}

			if item.repeatOnFinish {
				time.Sleep(100 * time.Millisecond)
				c.runnableStack <- item
				return
			}

			if c.dec() <= 0 {
				close(c.runnableStack)
			}
		}(item)
	}

	if c.finishersCount <= 0 {
		close(c.errorChan)
		c.cancel()
		return
	}

	for item := range c.finishersStack {
		if originalContext.Err() != nil {
			close(c.finishersStack)
			continue
		}

		go func(*runner) {
			err, _ := item.run(originalContext)
			if err != nil {
				c.errorChan <- err
			}

			if c.decFinishers() <= 0 {
				close(c.finishersStack)
			}
		}(item)
	}
	close(c.errorChan)
	c.cancel()
}

func (c *launcher) logErrors(ctx context.Context) {
	for err := range c.errorChan {
		c.logErr(ctx, err)
	}
	c.jobsDone <- nil
}

func (c *launcher) logErr(ctx context.Context, err error) {
	if errors.Is(err, context.Canceled) {
		select {
		case <-ctx.Done():
			return
		default:
			c.logErrAddPanicPrefix(err)
			return
		}
	}
	c.logErrAddPanicPrefix(err)
}

func (c *launcher) logErrAddPanicPrefix(err error) {
	if strings.Contains(err.Error(), panicError.Error()) {
		c.logger.Error(fmt.Sprintf("panic happened: %s", err.Error()))
		return
	}
	c.logger.Error(fmt.Sprintf("error happened: %s", err.Error()))
}

func (c *launcher) waitGraceful() {
	timeout, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	select {
	case <-timeout.Done():
		fmt.Println("graceful timeout done")
		return
	case <-c.jobsDone:
		fmt.Println("jobs done")
		return
	}
}

func (c *launcher) waitForInterruption() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	c.cancel()
}

func (c *launcher) inc() {
	atomic.AddInt64(&c.runnableCount, 1)
}

func (c *launcher) dec() int64 {
	return atomic.AddInt64(&c.runnableCount, -1)
}

func (c *launcher) incFinishers() {
	atomic.AddInt64(&c.finishersCount, 1)
}
func (c *launcher) decFinishers() int64 {
	return atomic.AddInt64(&c.finishersCount, -1)
}
