package test_clocks

import (
	"time"
)

type neverTicks struct {
	c chan time.Time
}

func NeverTick() *neverTicks {
	return &neverTicks{
		c: make(chan time.Time),
	}
}

func (ticker *neverTicks) Channel() <-chan time.Time {
	return ticker.c
}

func (ticker *neverTicks) Stop() {
	close(ticker.c)
}
