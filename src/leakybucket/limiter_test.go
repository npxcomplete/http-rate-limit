package leakybucket

import (
	"bytes"
	"github.com/npxcomplete/http-rate-limit/src/test_clocks"
	"github.com/npxcomplete/http-rate-limit/src/test_logger"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var start = time.Date(2000, 3, 12, 10, 10, 10, 0, time.UTC)

func Test_smoke_test_happy_path_with_http(t *testing.T) {
	config := Config{
		Tenancy: uniformLimits,
		TenantCapacity: 1,
	}
	limiter := NewRateLimiter(config)
	limiter.clock = test_clocks.FixedClock{T: start}
	limiter.log = &test_logger.LineLogger{Lines: make([]string, 0, 8)}

	limitedServlet := StdMiddleware(limiter)(
		http.HandlerFunc(func(resp http.ResponseWriter, _ *http.Request) {
			resp.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest("GET", "/path/to/resource", &bytes.Buffer{})
	req.RemoteAddr = "255.255.255.255"

	for i := 0; i < int(config.Tenancy(req.RemoteAddr).Burst); i++ {
		resp := httptest.NewRecorder()
		limitedServlet.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
	}

	resp := httptest.NewRecorder()
	limitedServlet.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusTooManyRequests, resp.Code)
}
