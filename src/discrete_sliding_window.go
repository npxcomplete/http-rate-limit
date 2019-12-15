package ratelimit

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	caches "github.com/npxcomplete/caches/src"
)

///////// IMPLEMENTATION /////////
func errorMessage(requestLimit int, intervalLength time.Duration) string {
	return fmt.Sprintf(
		"A maximum of %d requests may be attempted each %d hours",
		requestLimit,
		intervalLength/time.Hour,
	)
}

type SlidingWindowConfig struct {
	RequestLimit      int
	SubIntervalCount  int
	CapacityBound     int
	IntervalLength    time.Duration
	subIntervalLength time.Duration
}

func NewSlidingWindowRateLimiter(
	config SlidingWindowConfig,
) slidingWindowRateLimiter {

	config.subIntervalLength = time.Duration(float64(config.IntervalLength) / float64(config.SubIntervalCount))
	return slidingWindowRateLimiter{
		cache:  NewLRUStringSWCBCache(config.CapacityBound),
		clock:  HardwareClock{},
		log:    StdOutLogger{},
		config: &config,
	}

}

func (limiter slidingWindowRateLimiter) Middleware(servlet http.Handler) http.HandlerFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		userId := uniqueUserIdentifier(req)

		if limiter.AttemptAccess(userId, costToServiceRequest(req)) {
			servlet.ServeHTTP(resp, req)
			return
		} // else access attempt failed

		resp.WriteHeader(http.StatusTooManyRequests)
		io.WriteString(resp, errorMessage(limiter.config.RequestLimit, limiter.config.IntervalLength))
		return
	}
}

func (limiter slidingWindowRateLimiter) AttemptAccess(userId string, accessCost accessCounter) bool {
	controlBlock, err := limiter.cache.Get(userId)
	if err == caches.MissingValueError {
		// this should be the only place where the rate limiter performs heap allocations
		// Since we are dealing with a fixed size cache, If GC becomes a problem we can
		// look at creating a static object pool to compensate, recycling the evicted element.
		controlBlock = &swcb{
			StartTimeOfIntervalLastAccessed: limiter.clock.Now(),
			IndexOfLastAccess:               0,
			IntervalBuckets:                 make([]accessCounter, limiter.config.SubIntervalCount),
			AccessesInCurrentWindow:         0,
		}

		evicted := limiter.cache.Put(userId, controlBlock)
		if evicted != nil && evicted.wouldBeLimited(limiter.config.IntervalLength, limiter.clock) {
			// If you see this error message in production it's
			// time to increase the rate limiters cache capacity or
			// change to a more memory efficient implementation.
			limiter.log.Error("Rate limiter forced to evict a client susceptible rate limiting.")
		}
	}

	return controlBlock.accessAttempt(limiter.clock, limiter.config, accessCost)
}

// Currently we need to count no higher than 100.
type accessCounter byte

// A sliding window control block
// This uses a ring buffer to track accesses in sub-intervals over the full interval giving us fixed memory
// usage in O(SUB_INTERVAL_COUNT). An alternative would be to store N `time.Time` values
// using O(MAX_REQUESTS_IN_FULL_INTERVAL) memory, however that makes the core parameter extremely expensive to increase
//
// The consequence of this approach is that the rate limiter has only sub interval precision.
// Suppose for 100 requests per hour limit, with a 5 minute sub interval that a batch job submits 100 requests at time
// 0:04:59, then the same batch job will gain permission to submit another 100 requests at time 1:00:00.
// Effectively shortening the full window to 55m 1s. Suppose we quietly increase the full interval by one
// sub interval to compensate? Now in the above example the client wouldn't be reset till 1:05:00. HOWEVER,
// if the client had submitted their 100 requests at time 0:00:01 they STILL wouldn't get to submit a second batch
// until 1:05:00 effectively lengthening the full interval!
//
// Since we advertise in API an interval of 1h, it creates an unpredictable experience to
// users to silently increase the interval. Therefore, we accept the edge case with a shortened full interval.
// 64 bytes + sizeof(IntervalBuckets) + 48 bytes for the cache entry overhead => 122 bytes per entry.
// 1GB stores ~8.2M entries
type swcb struct {
	StartTimeOfIntervalLastAccessed time.Time
	IndexOfLastAccess               int
	IntervalBuckets                 []accessCounter
	AccessesInCurrentWindow         int
}

func (cb *swcb) wouldBeLimited(intervalLength time.Duration, clock Clock) bool {
	now := clock.Now()
	return cb.StartTimeOfIntervalLastAccessed.Add(intervalLength).After(now)
}

// Updates the control block and returns true if the request is allowed to proceed.
func (cb *swcb) accessAttempt(clock Clock, config *SlidingWindowConfig, accessCost accessCounter) bool {
	cb.reconcileSubWindows(config, clock.Now())

	// The astute reader will notice that we are still incrementing the request count even when the limit has
	// already been reached! Why? Suppose one of your clients is thoughtless / lazy. They might just be
	// running an infinite retry attempt when they hit rate limiting!!! If we only increment for successfully
	// allowed requests they will retry until the rate limiter resets their window, run through their request quota,
	// then continue to DoS the rate limiter until the window resets. Great for them, their code works, bad for us
	// we have to deal with all their wasted traffic. Instead we continue to deny them until the end of time, OR until
	// they wise up and start rate limiting themselves.

	quotaAvailable := cb.AccessesInCurrentWindow < config.RequestLimit

	// increment guarding against overflow errors
	quotaUsed := cb.IntervalBuckets[cb.IndexOfLastAccess]
	if quotaUsed+accessCost > quotaUsed {
		cb.IntervalBuckets[cb.IndexOfLastAccess] = quotaUsed + accessCost
		cb.AccessesInCurrentWindow = cb.AccessesInCurrentWindow + int(accessCost)
	}
	return quotaAvailable
}

func maxInt(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}

func (cb *swcb) reconcileSubWindows(config *SlidingWindowConfig, now time.Time) {
	durationSinceLastAccess := now.Sub(cb.StartTimeOfIntervalLastAccessed)
	if durationSinceLastAccess > config.IntervalLength {
		// re-use existing buckets to prevent memory fragmentation
		for i := 0; i < len(cb.IntervalBuckets); i++ {
			cb.IntervalBuckets[i] = 0
		}

		*cb = swcb{
			IntervalBuckets:                 cb.IntervalBuckets,
			AccessesInCurrentWindow:         0,
			StartTimeOfIntervalLastAccessed: now,
			IndexOfLastAccess:               0,
		}
	} else {
		// sub interval iterator
		i := ModuloRel{value: cb.IndexOfLastAccess, mod: config.SubIntervalCount}
		// starting in interval 1 and adding 3.5 intervals should place us in interval 4
		intervalsSinceLastAccess := int(math.Floor(float64(durationSinceLastAccess) / float64(config.subIntervalLength)))
		// current sub interval index
		k := i.Add(intervalsSinceLastAccess)

		for ; i != k; {
			i = i.Increment()
			cb.AccessesInCurrentWindow -= int(cb.IntervalBuckets[i.value])
			cb.IntervalBuckets[i.value] = 0
		}
		cb.StartTimeOfIntervalLastAccessed =
			cb.StartTimeOfIntervalLastAccessed.Add(time.Duration(intervalsSinceLastAccess) * config.subIntervalLength)
		cb.IndexOfLastAccess = k.value
	}
}

type slidingWindowRateLimiter struct {
	// Use a fixed capacity cache to memory bound our rate limiter
	// Consequence: Only the the noisiest N clients will be rate limited.
	cache LRUStringSWCBCache

	// for testing algorithms involving time we need a mockable time source
	clock Clock

	log Logger

	config *SlidingWindowConfig
}

func uniqueUserIdentifier(req *http.Request) string {
	return req.RemoteAddr
}

// a small nod to the fact that requests are not all created equal.
// if some requests would take an order of magnitude more work to service,
// then we probably need to treat user resource usage more carefully.
func costToServiceRequest(req *http.Request) accessCounter {
	return 1
}
