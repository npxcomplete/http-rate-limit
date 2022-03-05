package ratelimit

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

///////// EXPORTS /////////

type RateLimiter interface {
	AttemptAccess(userId string, requestCost uint64) bool
}

func StdMiddleware(
	limiter RateLimiter,
) func(servlet http.Handler) http.HandlerFunc {
	return Middleware(limiter, UniqueTenantIdentifier, FixedRequestCost)
}

func Middleware(
	limiter RateLimiter,
	tenantIdentifier func(r *http.Request) string,
	costOfRequest func(req *http.Request) uint64,
) func(servlet http.Handler) http.HandlerFunc {
	return func(servlet http.Handler) http.HandlerFunc {
		return func(resp http.ResponseWriter, req *http.Request) {
			tenantId := tenantIdentifier(req)

			if limiter.AttemptAccess(tenantId, costOfRequest(req)) {
				servlet.ServeHTTP(resp, req)
				return
			} // else access attempt failed

			resp.WriteHeader(http.StatusTooManyRequests)
			io.WriteString(resp, "Throttle limit exceeded.")
			return
		}
	}
}

///////// DEPENDS ON /////////

// Abstracts over time.Now, see TimeStamp
type Clock interface {
	Now() time.Time
}

type Logger interface {
	Error(msg string)
}

///////// Simple Dependency Implementations /////////

var _ Clock = HardwareClock{}

type HardwareClock struct{}

func (_ HardwareClock) Now() time.Time {
	return time.Now()
}

type StdOutLogger struct{}

func (_ StdOutLogger) Error(msg string) {
	fmt.Printf(msg)
}

func UniqueTenantIdentifier(req *http.Request) string {
	return req.RemoteAddr
}

// a small nod to the fact that requests are not all created equal.
// if some requests would take an order of magnitude more work to service,
// then we probably need to treat user resource usage more carefully.
func FixedRequestCost(req *http.Request) uint64 {
	return 1
}