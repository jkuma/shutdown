# Shutdown [![Go Report Card](https://goreportcard.com/badge/github.com/jkuma/shutdown)](https://goreportcard.com/report/github.com/jkuma/shutdown)

Shutdown package aims to gracefully shut down a given function with configurable options.

## Options

`shutdown.New()` accepts a collection of `shutdown.Options` as listed bellow :

- `shutdown.WithContext()` pass a parent context.
- `shutdown.WithLogger()` pass a custom logger implementing `shutdown.Logger`.
- `shutdown.WithSignals()` configure `os.Signal` to be triggered as shutdown event
- `shutdown.WithTimeExpiration()` set up a timeout so shutdown functions won't hang up the system.

## Usage

Any function requiring graceful shutdown shall be registered through `Shutdown.Register()`. These 
functions must implement `shutdown.Closable` which is a `func() error` type that fit perfectly 
with `io.Closer` interface.

`os.Interrupt` and `syscall.SIGTERM` are listened by default if no signals configured with `shutdown.WithSignals()`.

No timeout is set by default.

`log.Default()` is set as default logger.

```go
package main

import (
	"context"
	"log"
	"net/http"

	"github.com/jkuma/shutdown"
)

func main() {
	ctx := context.Background()
	server := &http.Server{}

	httpshut := func() error {
		return server.Shutdown(ctx)
	}

	shut := shutdown.New(shutdown.WithContext(ctx)).Register(httpshut)

	shut.RunGraceful(func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	})
}
```
