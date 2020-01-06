Quick Start
==================

To add the library to your go modules dependencies:
```bash
go get github.com/npxcomplete/http-rate-limit@v0.1.0
```

A minimal example for an http server:
```go
package examples

import (
	"io"
	"log"
	"net/http"
	"time"

	ratelimit "github.com/npxcomplete/http-rate-limit/src"
)

func main() {
	rateLimiter := ratelimit.NewSlidingWindowRateLimiter(ratelimit.SlidingWindowConfig{
		RequestLimit:     100,
		SubIntervalCount: 10,
		CapacityBound:    5_000,
		IntervalLength:   1 * time.Hour,
	})

	log.Fatal(http.ListenAndServe(
		":8080",
		rateLimiter.Middleware(http.HandlerFunc(HelloWorld)),
	))
}

func HelloWorld(resp http.ResponseWriter, req *http.Request) {
	resp.WriteHeader(200)
	io.WriteString(resp, "Hello World")
}
```