package test_clocks

import (
	ratelimit "github.com/npxcomplete/http-rate-limit/src"
	"time"
)

type TimeSkipper struct {
	c        chan time.Time
	clock    ratelimit.Clock
}

func TimeSkips(clock ratelimit.Clock) *TimeSkipper {
	return &TimeSkipper{
		c: make(chan time.Time),
		clock: clock,
	}
}

func (ticker *TimeSkipper) Channel() <-chan time.Time {
	return ticker.c
}

func (ticker *TimeSkipper) Stop() {
	close(ticker.c)
}

func (ticker *TimeSkipper) Step() {
	ticker.c <- ticker.clock.Now()
}
