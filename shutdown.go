package shutdown

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Shutdown struct {
	timeout   time.Duration
	signals   []os.Signal
	functions []Closable
}

type Closable func() error

func New(signals ...os.Signal) *Shutdown {
	if len(signals) == 0 {
		signals = append(signals, os.Interrupt, syscall.SIGTERM)
	}

	return &Shutdown{
		timeout: 10 * time.Second,
		signals: signals,
	}
}

func (s *Shutdown) SetExpiration(t time.Duration) *Shutdown {
	s.timeout = t

	return s
}

// Register a collection of Closable functions that will be gracefully
// closed.
func (s *Shutdown) Register(c ...Closable) *Shutdown {
	s.functions = append(s.functions, c...)

	return s
}

// Process function with gracefully shutdown.
func (s *Shutdown) Process(fn func()) {
	signalReceived := make(chan os.Signal, 1)
	signal.Notify(signalReceived, s.signals...)

	done := make(chan struct{}, 1)

	go func() {
		if fn != nil {
			fn()
		}

		done <- struct{}{}
	}()

	select {
	case <-done:
		return
	case <-signalReceived:
		fmt.Println("Shutdown gracefully...")

		var wg sync.WaitGroup

		timeoutFunc := time.AfterFunc(s.timeout, func() {
			log.Fatalf("timeout %d ms has been elapsed, force exit", s.timeout.Milliseconds())
		})
		defer timeoutFunc.Stop()

		for _, function := range s.functions {
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
