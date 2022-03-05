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
	ratelimit "github.com/npxcomplete/http-rate-limit/src"
	"github.com/npxcomplete/http-rate-limit/src/leakybucket"
	"io"
	"log"
	"net/http"
)

var UniformLimit = leakybucket.TenantLimit{
	Rate:  100,
	Burst: 100,
}

func main() {
	uniformLimits := func(tenant string) *leakybucket.TenantLimit {
		return &UniformLimit
	}
	rateLimiter := leakybucket.NewRateLimiter(leakybucket.Config{
		Tenancy:        uniformLimits,
		TenantCapacity: 1,
	})

	log.Fatal(http.ListenAndServe(
		":8080",
		ratelimit.StdMiddleware(rateLimiter)(http.HandlerFunc(HelloWorld)),
	))
}

func HelloWorld(resp http.ResponseWriter, req *http.Request) {
	resp.WriteHeader(200)
	io.WriteString(resp, "Hello World")
}
```