package shutdown

import (
	"context"
	"log"
	"os"
	"syscall"
	"time"
)

type Logger interface {
	Println(v ...any)
	Fatalf(format string, v ...any)
}

type Option interface {
	init()
	apply(*funcOption)
	context() context.Context
	log() Logger
	timeout() time.Duration
	signals() []os.Signal
}

type funcOption struct {
	f    func(*funcOption)
	l    Logger
	ctx  context.Context
	expr time.Duration
	sigs []os.Signal
}

func (fo *funcOption) init() {
	if fo == nil {
		return
	}

	if fo.l == nil {
		fo.l = log.Default()
	}

	if fo.ctx == nil {
		fo.ctx = context.Background()
	}

	if fo.expr <= 0 {
		fo.expr = 0
	}

	if len(fo.sigs) == 0 {
		fo.sigs = append(fo.sigs, os.Interrupt, syscall.SIGTERM)
	}
}

func (fo *funcOption) apply(do *funcOption) {
	fo.f(do)
}

func (fo *funcOption) context() context.Context {
	return fo.ctx
}

func (fo *funcOption) log() Logger {
	return fo.l
}

func (fo *funcOption) timeout() time.Duration {
	return fo.expr
}

func (fo *funcOption) signals() []os.Signal {
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

// WithLogger injects a service implementing Logger.
func WithLogger(l Logger) Option {
	return newFuncOption(func(option *funcOption) {
		option.l = l
	})
}

// WithSignals listen to given os.Signal as shutdown trigger.
func WithSignals(sigs ...os.Signal) Option {
	return newFuncOption(func(option *funcOption) {
		option.sigs = sigs
	})
}
