package ratelimit

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/npxcomplete/http-rate-limit/src/test_clocks"
	"github.com/npxcomplete/http-rate-limit/src/test_logger"
)

var start = time.Date(2000, 3, 12, 10, 10, 10, 0, time.UTC)

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
