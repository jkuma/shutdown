package shutdown_test

import (
	"context"
	"log"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/jkuma/shutdown"
)

type ServiceA struct {
	Closed bool
}

func (s *ServiceA) Close() error {
	s.Closed = true
	return nil
}

type ServiceB struct {
	Closed bool
}

func (s *ServiceB) Shutdown() error {
	time.Sleep(10 * time.Millisecond)
	s.Closed = true
	return nil
}

func TestShutdown_RunGraceful(t *testing.T) {
	sa := &ServiceA{}
	sb := &ServiceB{}

	shut := shutdown.
		New(shutdown.WithSignals(syscall.SIGIO)).
		Register(sa.Close, sb.Shutdown)

	p, err := os.FindProcess(syscall.Getpid())
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		for {
			_ = p.Signal(syscall.SIGIO)
		}
	}()

	shut.RunGraceful(func() {
		time.Sleep(200 * time.Millisecond)
	})

	if !sa.Closed {
		t.Errorf("Service A shall be closed")
	}

	if !sb.Closed {
		t.Errorf("Service B shall be closed")
	}
}

func TestShutdown_RunGracefulWithLogger(t *testing.T) {
	shutdown.
		New(shutdown.WithLogger(&log.Logger{})).
		RunGraceful(func() {
			time.Sleep(1 * time.Nanosecond)
		})
}

func TestShutdown_RunGracefulWithParentContext(t *testing.T) {
	var run bool

	ctx, cancel := context.WithCancel(context.Background())

	shut := shutdown.New(shutdown.WithContext(ctx))
	cancel()

	shut.RunGraceful(func() {
		run = true
	})

	if run {
		t.Errorf("Main function shall be ")
	}
}

func TestShutdown_RunGracefulWithExpiration(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		sa := &ServiceA{}
		sb := &ServiceB{}

		shut := shutdown.
			New(shutdown.WithSignals(syscall.SIGIO), shutdown.WithTimeExpiration(1*time.Nanosecond)).
			Register(sa.Close, sb.Shutdown)

		p, err := os.FindProcess(syscall.Getpid())
		if err != nil {
			t.Fatal(err)
		}

		go func() {
			for {
				_ = p.Signal(syscall.SIGIO)
			}
		}()

		shut.RunGraceful(func() {
			time.Sleep(300 * time.Millisecond)
		})

		return
	}

	// Invoke a subprocess to make sure code above breaks
	// and return an os.Exit(1) code.
	//
	// The os.Exit(1) code is due to gracefully shutdown timeout.
	//
	// https://go.dev/talks/2014/testing.slide#23
	cmd := exec.Command(os.Args[0], "-test.run=TestShutdown_RunGracefulWithExpiration")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}
