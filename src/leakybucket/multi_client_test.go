package leakybucket

import (
	"context"
	"github.com/npxcomplete/http-rate-limit/src/test_clocks"
	"github.com/npxcomplete/http-rate-limit/src/test_logger"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"testing"
)

type ServiceProxy struct {
	ref LeakyBucket
}

func (proxy ServiceProxy) Join(ctx context.Context, req *JoinRequest) (*JoinResponse, error) {
	return proxy.ref.Join(ctx, req)
}

func (proxy ServiceProxy) Sync(ctx context.Context, req *SyncRequest) (*SyncResponse, error) {
	return proxy.ref.Sync(ctx, req)
}

func (proxy ServiceProxy) NewClients() []LeakyBucket {
	if proxy.ref == nil {
		return []LeakyBucket{}
	}
	return []LeakyBucket{proxy}
}

func Test_multiple_clients_can_push_capacity_negative(t *testing.T) {
	fredClient := ServiceProxy{}
	georgeClient := ServiceProxy{}

	testClock := test_clocks.FixedClock{T: start}
	skipper := test_clocks.TimeSkips(testClock)
	config := Config{
		TenantCapacity: 4,
		Tenancy:        uniformLimits,
		InitPeerGroup:  nil,
		Port:           12345,
		backgroundSync: skipper,
		clock:          testClock,
	}

	configFred := config
	configFred.InitPeerGroup = georgeClient
	fred := NewRateLimiter(configFred)
	fredClient.ref = fred
	fred.log = &test_logger.LineLogger{Lines: make([]string, 0, 8)}

	configGeorge := config
	configGeorge.InitPeerGroup = fredClient
	george := NewRateLimiter(configGeorge)
	georgeClient.ref = george
	george.log = &test_logger.LineLogger{Lines: make([]string, 0, 8)}

	// WHEN a tenant consumes capacity in parallel they can succeed
	fred.AttemptAccess(100, "111.111.111.111")
	george.AttemptAccess(100, "111.111.111.111")

	// AND WHEN SYNC occurs
	skipper.Step()

	// THEN the available capacity goes negative on both peers
	assert.False(t, fred.AttemptAccess(1, "111.111.111.111"))
	assert.False(t, george.AttemptAccess(1, "111.111.111.111"))
}

func Test_multiple_clients_intialize_from_peer(t *testing.T) {
	fredClient := ServiceProxy{}
	georgeClient := ServiceProxy{}

	testClock := test_clocks.FixedClock{T: start}
	config := Config{
		TenantCapacity: 4,
		Tenancy:        uniformLimits,
		InitPeerGroup:  nil,
		Port:           0,
		backgroundSync: test_clocks.TimeSkips(testClock),
		clock:          testClock,
	}

	// GIVEN an existing peer group
	configFred := config
	configFred.InitPeerGroup = georgeClient
	fred := NewRateLimiter(configFred)
	fredClient.ref = fred
	fred.log = &test_logger.LineLogger{Lines: make([]string, 0, 8)}

	fred.AttemptAccess(55, "111.111.111.111")

	// WHEN a new peer joins
	configGeorge := config
	configGeorge.InitPeerGroup = fredClient
	george := NewRateLimiter(configGeorge)
	georgeClient.ref = george
	george.log = &test_logger.LineLogger{Lines: make([]string, 0, 8)}

	// THEN the new peer is aware of tenant capacity exhaustion
	assert.False(t, george.AttemptAccess(55, "111.111.111.111"))
}

func HelloWorld(resp http.ResponseWriter, req *http.Request) {
	resp.WriteHeader(200)
	io.WriteString(resp, "Hello World")
}
