package leakybucket

import "time"

type LBTicker interface {
	Channel() <-chan time.Time
	Stop()
}

type ticker struct {
	t *time.Ticker
}

func NewTicker(d time.Duration) ticker {
	return ticker { time.NewTicker(d) }
}

func (t ticker) Channel() <-chan time.Time {
	return t.t.C
}

func (t ticker) Stop() {
	t.t.Stop()
}