package shutdown

import (
	"context"
	"os"
	"syscall"
	"time"
)

type Option interface {
	apply(*funcOption)
	context() context.Context
	timeout() time.Duration
	signals() []os.Signal
}

type funcOption struct {
	f    func(*funcOption)
	ctx  context.Context
	expr time.Duration
	sigs []os.Signal
}

func (fo *funcOption) apply(do *funcOption) {
	fo.f(do)
}

func (fo *funcOption) context() context.Context {
	if fo.ctx == nil {
		return context.Background()
	}

	return fo.ctx
}

func (fo *funcOption) timeout() time.Duration {
	if fo.expr <= 0 {
		return 0
	}

	return fo.expr
}

func (fo *funcOption) signals() []os.Signal {
	if len(fo.sigs) == 0 {
		return []os.Signal{os.Interrupt, syscall.SIGTERM}
	}

	return fo.sigs
}

func newFuncOption(f func(*funcOption)) *funcOption {
	return &funcOption{
		f: f,
	}
}

func parseOptions(opts ...Option) Option {
	o := new(funcOption)
	for _, opt := range opts {
		opt.apply(o)
	}
	return o
}

// WithTimeExpiration set a timeout for shutdown functions.
func WithTimeExpiration(expr time.Duration) Option {
	return newFuncOption(func(option *funcOption) {
		if expr > 0 {
			option.expr = expr
		}
	})
}

// WithContext allows to pass a context.
func WithContext(ctx context.Context) Option {
	return newFuncOption(func(options *funcOption) {
		options.ctx = ctx
	})
}

// WithSignals listen to given os.Signal as shutdown trigger.
func WithSignals(sigs ...os.Signal) Option {
	return newFuncOption(func(option *funcOption) {
		option.sigs = sigs
	})
}
