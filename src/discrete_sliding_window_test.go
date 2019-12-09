package ratelimit

import (
	"bytes"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"unsafe"

	"github.com/stretchr/testify/assert"

	"github.com/npxcomplete/http-rate-limit/src/test_clocks"
	"github.com/npxcomplete/http-rate-limit/src/test_logger"
)

var start = time.Date(2000, 3, 12, 10, 10, 10, 0, time.UTC)
var bound = math.Pow(2, 8*float64(unsafe.Sizeof(accessCounter(0))))

func Test_smoke_test_happy_path_with_http(t *testing.T) {
	limiter := &slidingWindowRateLimiter{
		cache: NewLRUStringSWCBCache(10),
		clock: &test_clocks.CreepingClock{T: start, Increment: 1 * time.Millisecond},
		log:   &test_logger.LineLogger{make([]string, 0, 8)},
	}

	limitedServlet := limiter.Middleware(http.HandlerFunc(func(resp http.ResponseWriter, _ *http.Request) {
		resp.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/path/to/resource", &bytes.Buffer{})
	req.RemoteAddr = "255.255.255.255"

	for i := 0; i < MAX_REQUESTS_IN_FULL_INTERVAL; i++ {
		resp := httptest.NewRecorder()
		limitedServlet.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
	}

	resp := httptest.NewRecorder()
	limitedServlet.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusTooManyRequests, resp.Code)

}

func Test_exceeding_rate_limit_causes_rejection(t *testing.T) {
	limiter := &slidingWindowRateLimiter{
		cache: NewLRUStringSWCBCache(10),
		clock: &test_clocks.CreepingClock{T: start, Increment: 1 * time.Millisecond},
		log:   &test_logger.LineLogger{make([]string, 0, 8)},
	}

	for i := 0; i < MAX_REQUESTS_IN_FULL_INTERVAL; i++ {
		assert.True(t, limiter.AttemptAccess("Dave", 1), fmt.Sprintf("i = %d", i))
	}
	assert.False(t, limiter.AttemptAccess("Dave", 1), fmt.Sprintf("i = 101"))
}

func Test_reset_on_windowed_rotation(t *testing.T) {
	limiter := &slidingWindowRateLimiter{
		cache: NewLRUStringSWCBCache(10),
		clock: &test_clocks.CreepingClock{T: start, Increment: 1 * time.Minute},
		log:   &test_logger.LineLogger{make([]string, 0, 8)},
	}

	for i := 0; i < MAX_REQUESTS_IN_FULL_INTERVAL*2; i++ {
		assert.True(t, limiter.AttemptAccess("Dave", 1), fmt.Sprintf("i = %d", i))
	}
}

func Test_reset_on_giant_pause(t *testing.T) {
	limiter := &slidingWindowRateLimiter{
		cache: NewLRUStringSWCBCache(10),
		clock: &test_clocks.CreepingClock{T: start, Increment: 2 * time.Hour},
		log:   &test_logger.LineLogger{make([]string, 0, 8)},
	}

	for i := 0; i < MAX_REQUESTS_IN_FULL_INTERVAL*2; i++ {
		assert.True(t, limiter.AttemptAccess("Dave", 1), fmt.Sprintf("i = %d", i))
	}
}

func Test_cache_space_exhaustion_is_logged(t *testing.T) {
	logs := &test_logger.LineLogger{make([]string, 0, 8)}
	limiter := &slidingWindowRateLimiter{
		cache: NewLRUStringSWCBCache(1),
		clock: &test_clocks.CreepingClock{T: start, Increment: 1 * time.Second},
		log:   logs,
	}

	// put Dave in a limited state
	for i := 0; i < MAX_REQUESTS_IN_FULL_INTERVAL; i++ {
		limiter.AttemptAccess("Dave", 1)
	}
	assert.False(t, limiter.AttemptAccess("Dave", 1))
	// then evict Dave by adding gary
	assert.Equal(t, 0, len(logs.Lines))
	limiter.AttemptAccess("Gary", 1)
	assert.Equal(t, 1, len(logs.Lines))
	assert.Equal(t, "Rate limiter forced to evict a client susceptible rate limiting.", logs.Lines[0])
}

func Test_excess_cache_space_implies_no_eviction(t *testing.T) {
	logs := &test_logger.LineLogger{make([]string, 0, 8)}
	limiter := &slidingWindowRateLimiter{
		cache: NewLRUStringSWCBCache(2),
		clock: &test_clocks.CreepingClock{T: start, Increment: 1 * time.Second},
		log:   logs,
	}

	// put Dave in a limited state
	for i := 0; i < MAX_REQUESTS_IN_FULL_INTERVAL; i++ {
		limiter.AttemptAccess("Dave", 1)
	}
	assert.False(t, limiter.AttemptAccess("Dave", 1))

	limiter.AttemptAccess("Gary", 1)
	assert.Equal(t, 0, len(logs.Lines))
}

func Test_cache_eviction_of_potentially_rate_limited_clients_are_logged(t *testing.T) {
	logs := &test_logger.LineLogger{make([]string, 0, 8)}
	limiter := &slidingWindowRateLimiter{
		cache: NewLRUStringSWCBCache(1),
		clock: &test_clocks.CreepingClock{T: start, Increment: 1 * time.Second},
		log:   logs,
	}

	// put Dave in a limited state
	limiter.AttemptAccess("Dave", 1)
	limiter.AttemptAccess("Gary", 1)
	if !assert.Equal(t, 1, len(logs.Lines)) {
		return
	}
	assert.Equal(t, "Rate limiter forced to evict a client susceptible rate limiting.", logs.Lines[0])
}

func Test_byte_boundary_no_overflow(t *testing.T) {
	logs := &test_logger.LineLogger{make([]string, 0, 8)}
	limiter := &slidingWindowRateLimiter{
		cache: NewLRUStringSWCBCache(2),
		clock: &test_clocks.CreepingClock{T: start, Increment: 1 * time.Second},
		log:   logs,
	}

	// what happens on overflow
	bound := math.Pow(2, 8*float64(unsafe.Sizeof(accessCounter(0))))
	for i := 0; i < int(bound); i++ {
		limiter.AttemptAccess("Dave", 1)
	}
	assert.False(t, limiter.AttemptAccess("Dave", 1))

	limiter.AttemptAccess("Gary", 1)
	assert.Equal(t, 0, len(logs.Lines))
}

func Test_whitebox_forward_intervals_cleared(t *testing.T) {
	logs := &test_logger.LineLogger{make([]string, 0, 8)}

	limiter := &slidingWindowRateLimiter{
		cache: NewLRUStringSWCBCache(2),
		clock: &test_clocks.CreepingClock{T: start, Increment: (SUB_INTERVAL_LENGTH * 2) + 1},
		log:   logs,
	}

	limiter.AttemptAccess("Dave", 1)
	cb, _ := limiter.cache.Get("Dave")
	cb.AccessesInCurrentWindow = 100
	cb.IntervalBuckets = []accessCounter{10, 10, 10, 10, 10, 10, 10, 10, 10, 10,}
	cb.IndexOfLastAccess = 2

	assert.True(t, limiter.AttemptAccess("Dave", 1))
	assert.Equal(t, accessCounter(10), cb.IntervalBuckets[2])
	assert.Equal(t, accessCounter(0), cb.IntervalBuckets[3])
	assert.Equal(t, accessCounter(1), cb.IntervalBuckets[4])
	assert.Equal(t, 4, cb.IndexOfLastAccess)
}

func Test_whitebox_backwards_intervals_cleared(t *testing.T) {
	logs := &test_logger.LineLogger{make([]string, 0, 8)}

	limiter := &slidingWindowRateLimiter{
		cache: NewLRUStringSWCBCache(2),
		clock: &test_clocks.CreepingClock{T: start, Increment: (SUB_INTERVAL_LENGTH * 2) + 1},
		log:   logs,
	}

	limiter.AttemptAccess("Dave", 1)
	cb, _ := limiter.cache.Get("Dave")
	cb.AccessesInCurrentWindow = 100
	cb.IntervalBuckets = []accessCounter{10, 10, 10, 10, 10, 10, 10, 10, 10, 10,}
	cb.IndexOfLastAccess = len(cb.IntervalBuckets) - 2

	assert.True(t, limiter.AttemptAccess("Dave", 1))
	assert.Equal(t, accessCounter(10), cb.IntervalBuckets[len(cb.IntervalBuckets)-2])
	assert.Equal(t, accessCounter(0), cb.IntervalBuckets[len(cb.IntervalBuckets)-1])
	assert.Equal(t, accessCounter(1), cb.IntervalBuckets[0])
	assert.Equal(t, 0, cb.IndexOfLastAccess)
}

func Test_whitebox_same_interval_no_clear(t *testing.T) {
	logs := &test_logger.LineLogger{make([]string, 0, 8)}

	limiter := &slidingWindowRateLimiter{
		cache: NewLRUStringSWCBCache(2),
		clock: &test_clocks.CreepingClock{T: start, Increment: 1},
		log:   logs,
	}

	limiter.AttemptAccess("Dave", 1)
	cb, _ := limiter.cache.Get("Dave")
	cb.AccessesInCurrentWindow = 92
	cb.IntervalBuckets = []accessCounter{10, 10, 10, 10, 10, 10, 10, 10, 1, 10,}
	cb.IndexOfLastAccess = len(cb.IntervalBuckets) - 2

	assert.True(t, limiter.AttemptAccess("Dave", 1))
	assert.Equal(t, accessCounter(2), cb.IntervalBuckets[len(cb.IntervalBuckets)-2])
	assert.Equal(t, len(cb.IntervalBuckets)-2, cb.IndexOfLastAccess)
}

func Test_whitebox_full_interval_step_full_clear(t *testing.T) {
	logs := &test_logger.LineLogger{make([]string, 0, 8)}

	limiter := &slidingWindowRateLimiter{
		cache: NewLRUStringSWCBCache(2),
		clock: &test_clocks.CreepingClock{T: start, Increment: FULL_INTERVAL_LENGTH + 1},
		log:   logs,
	}

	limiter.AttemptAccess("Dave", 1)
	cb, _ := limiter.cache.Get("Dave")
	cb.AccessesInCurrentWindow = 92
	cb.IntervalBuckets = []accessCounter{10, 10, 10, 10, 10, 10, 10, 10, 1, 10,}
	cb.IndexOfLastAccess = len(cb.IntervalBuckets) - 2

	assert.True(t, limiter.AttemptAccess("Dave", 1))
	assert.Equal(t, 0, cb.IndexOfLastAccess)
	assert.Equal(t, 1, cb.AccessesInCurrentWindow)
	assert.Equal(t, []accessCounter{1,0,0,0,0,0,0,0,0,0,}, cb.IntervalBuckets)
}

func Test_whitebox_maximum_sub_interval_clear(t *testing.T) {
	logs := &test_logger.LineLogger{make([]string, 0, 8)}

	limiter := &slidingWindowRateLimiter{
		cache: NewLRUStringSWCBCache(2),
		clock: &test_clocks.CreepingClock{T: start, Increment: FULL_INTERVAL_LENGTH - SUB_INTERVAL_LENGTH + 1},
		log:   logs,
	}

	limiter.AttemptAccess("Dave", 1)
	cb, _ := limiter.cache.Get("Dave")
	cb.AccessesInCurrentWindow = 100
	cb.IntervalBuckets = []accessCounter{10, 10, 10, 10, 10, 10, 10, 10, 10, 10,}
	cb.IndexOfLastAccess = len(cb.IntervalBuckets) - 2

	limiter.AttemptAccess("Dave", 1)
	assert.Equal(t, len(cb.IntervalBuckets) - 3, cb.IndexOfLastAccess)
	assert.Equal(t, 11, cb.AccessesInCurrentWindow)
	assert.Equal(t, []accessCounter{0,0,0,0,0,0,0,1,10,0,}, cb.IntervalBuckets)
}