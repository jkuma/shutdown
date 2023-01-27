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

// Shutdown package aims to gracefully shut down a given function.
//
// Any shutdown Function shall be registered through Shutdown.Register().
// They are triggered if system receives at least one os.Signal configured
// in shutdown.New().
//
// By default, a timeout of 10 seconds is set in order to ensure that
// shutdown Function won’t hang up the system.
//
// Example of usage:
//
//	type ServiceA func() error
//	type ServiceB func() error
//
//	shut := New().
//		SetExpiration(3 * time.Second).
//		Register(ServiceA{}, ServiceB{})
//
//	shut.Process(func() {
//		http.ListenAndServe(":8080", nil)
//	})
type Shutdown struct {
	timeout   time.Duration
	signals   []os.Signal
	functions []Function
}

type Function func() error

// New instance of a new Shutdown service.
func New(signals ...os.Signal) *Shutdown {
	if len(signals) == 0 {
		signals = append(signals, os.Interrupt, syscall.SIGTERM)
	}

	return &Shutdown{
		timeout: 10 * time.Second,
		signals: signals,
	}
}

// SetExpiration timeout.
func (s *Shutdown) SetExpiration(t time.Duration) *Shutdown {
	s.timeout = t

	return s
}

// Register a collection of Function that will be gracefully closed.
func (s *Shutdown) Register(functions ...Function) *Shutdown {
	s.functions = append(s.functions, functions...)

	return s
}

// Process function with gracefully shutdown.
func (s *Shutdown) Process(fn func()) {
	signalReceived := make(chan os.Signal, 1)
	signal.Notify(signalReceived, s.signals...)

	done := make(chan struct{}, 1)

	go func() {
		fn()
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
			go func(fn Function) {
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
