package test_clocks

import (
	"time"
)

type CreepingClock struct {
	T         time.Time
	Increment time.Duration
}

func (c *CreepingClock) Now() time.Time {
	c.T = c.T.Add(c.Increment)
	return c.T
}

func NewMultiIncrementClock(
	T time.Time,
	Increments []time.Duration,
) *MultiIncrementClock {
	return &MultiIncrementClock{
		T:          T,
		Increments: Increments,
		Index:      0,
	}
}

type MultiIncrementClock struct {
	T          time.Time
	Increments []time.Duration
	Index      int
}

func (c *MultiIncrementClock) Now() time.Time {
	c.T = c.T.Add(c.Increments[c.Index])
	c.Index = c.Index + 1%len(c.Increments)
	return c.T
}
