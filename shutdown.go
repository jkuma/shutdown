package shutdown

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"
)

type Shutdown struct {
	Option
	fns []Closable
}

type Closable func() error

func New(opts ...Option) *Shutdown {
	option := parseOptions(opts...)

	return &Shutdown{
		Option: option,
	}
}

// Register a collection of Closable functions that will be
// gracefully closed.
func (s *Shutdown) Register(c ...Closable) *Shutdown {
	s.fns = c
	return s
}

// RunGraceful shutdown of given function.
func (s *Shutdown) RunGraceful(fn func()) {
	signalReceived := make(chan os.Signal, 1)
	signal.Notify(signalReceived, s.signals()...)

	ctx, cancel := context.WithCancel(s.context())

	go func() {
		select {
		case <-ctx.Done():
			return
		default:
			defer cancel()
			if fn != nil {
				fn()
			}
		}
	}()

	select {
	case <-ctx.Done():
		return
	case <-signalReceived:
		fmt.Println("Shutdown gracefully...")

		var wg sync.WaitGroup

		if s.timeout() > 0 {
			timeoutFunc := time.AfterFunc(s.timeout(), func() {
				log.Fatalf("Timeout %d ms has been elapsed, force exit", s.timeout().Milliseconds())
			})
			defer timeoutFunc.Stop()
		}

		for _, function := range s.fns {
			wg.Add(1)
			go func(fn Closable) {
				defer wg.Done()

				if err := fn(); err != nil {
					log.Printf("Service could not be closed: %s", err)
				}
			}(function)
		}

		wg.Wait()

		fmt.Println("Shutdown gracefully done")
	}
}
