# Shutdown

Shutdown package aims to gracefully shut down a given function.

# Usage

Any function requiring graceful shutdown shall be registered with `Shutdown.Register()`.

These functions are triggered if system receives at least one `os.Signal` configured as
parameters of `shutdown.New()`. 

`os.Interrupt` and `syscall.SIGTERM` are listened by default if no signals configured.

A timeout of 10 seconds is set by default to make sure that shutdown functions won’t
hang up the system. This duration can be changed through `Shutdown.SetExpiration()`.

```go
package main

import (
	"github.com/jkuma/shutdown"
)

type ServiceA struct {
	Closed bool
}

func (s *ServiceA) Close() error {
	s.Closed = true
	return nil
}

func main() {
	sa := &ServiceA{}
	shut := New().Register(sa.Close)

	shut.Process(func() {
		// Do stuff...
	})
}
```
