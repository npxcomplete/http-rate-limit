package examples

import (
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

	middleware := leakybucket.StdMiddleware(rateLimiter)
	log.Fatal(http.ListenAndServe(
		":8080",
		middleware(http.HandlerFunc(HelloWorld)),
	))
}

func HelloWorld(resp http.ResponseWriter, req *http.Request) {
	resp.WriteHeader(200)
	io.WriteString(resp, "Hello World")
}
