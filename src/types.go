package ratelimit

import (
	"fmt"
	"net/http"
	"time"
)

///////// EXPORTS /////////

type RateLimiter interface {
	// Returns a servlet that rate limits access to the given servlet
	Middleware(servlet http.Handler) http.HandlerFunc
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
