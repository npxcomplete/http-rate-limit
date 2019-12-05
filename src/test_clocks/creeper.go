package test_clocks

import "time"

type CreepingClock struct {
	T time.Time
	Increment time.Duration
}

func (c *CreepingClock) Now() time.Time {
	c.T = c.T.Add(c.Increment)
	return c.T
}
