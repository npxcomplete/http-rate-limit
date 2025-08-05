package ratelimit

import (
	"fmt"
	"net/http"
	"time"
)

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
	fmt.Printf("%s", msg)
}

func IPAddressOf(req *http.Request) string {
	return req.RemoteAddr
}

// a small nod to the fact that requests are not all created equal.
// if some requests would take an order of magnitude more work to service,
// then we probably need to treat user resource usage more carefully.
func FixedRequestCost(req *http.Request) uint32 {
	return 1
}
