package test_clocks

import "time"

type FixedClock struct {
	T         time.Time
}

func (c FixedClock) Now() time.Time {
	return c.T
}
