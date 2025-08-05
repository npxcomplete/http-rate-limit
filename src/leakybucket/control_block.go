package leakybucket

import (
	ratelimit "github.com/npxcomplete/http-rate-limit/src"
	"math"
	"sync"
	"time"
)

type controlBlock struct {
	mutex             sync.Mutex
	availableCapacity Capacity
	timeOfLastAccess  time.Time
	spentCapacity     AccessCost
}

func (cb *controlBlock) accessAttempt(
	limits *TenantLimit,
	clock ratelimit.Clock,
	accessCost AccessCost,
) bool {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := clock.Now()
	tdiff := now.Sub(cb.timeOfLastAccess)

	microsRefill := AccessCost(math.Floor(float64(int64(limits.Rate)*tdiff.Microseconds()) / 1_000_000.0))
	cb.availableCapacity = Capacity(math.Min(float64(cb.availableCapacity+int64(microsRefill)), float64(limits.Burst)))
	cb.timeOfLastAccess = now


	var quotaAvailable bool = int64(accessCost) <= cb.availableCapacity

	if quotaAvailable {
		cb.availableCapacity = cb.availableCapacity - int64(accessCost)
		cb.spentCapacity = cb.spentCapacity + accessCost
	}

	return quotaAvailable
}

func (cb *controlBlock) getAndClearSpentCapacity() uint32 {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	result := cb.spentCapacity
	cb.spentCapacity = 0
	return result
}

func (cb *controlBlock) decreaseAvailableCapacity(accessCost AccessCost) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.availableCapacity = Capacity(cb.availableCapacity - int64(accessCost))
}
