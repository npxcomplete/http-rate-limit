package leakybucket

import (
	caches "github.com/npxcomplete/caches/src"
	"github.com/npxcomplete/http-rate-limit/src"
	"math"
	"sync"
	"time"
)

type TenantLimit struct {
	Rate           float64
	Burst          float64
}

type Config struct {
	TenantCapacity int
	Tenancy func(tenant string) *TenantLimit
}

func NewRateLimiter(
	config Config,
) *leakyBucketRateLimiter {
	return &leakyBucketRateLimiter{
		cache:  NewStringLBCBCache(config.TenantCapacity),
		clock:  ratelimit.HardwareClock{},
		log:    ratelimit.StdOutLogger{},
		config: &config,
	}
}

func (limiter *leakyBucketRateLimiter) AttemptAccess(tenantId string, accessCost uint64) bool {
	var err error
	var cb *lbcb

	cb, err = limiter.cache.Get(tenantId)
	if err == caches.MissingValueError {
		cb = &lbcb{
			mutex:             sync.Mutex{},
			availableCapacity: limiter.config.Tenancy(tenantId).Burst,
			timeOfLastAccess:  limiter.clock.Now(),
		}
		limiter.cache.Put(tenantId, cb)
	} else if err != nil {
		return false
	}

	return cb.accessAttempt(tenantId, limiter.clock, limiter.config, leakyBucketAccessCost(accessCost))
}

type leakyBucketAccessCost = float64

type leakyBucketRateLimiter struct {
	// Use a fixed capacity cache to memory bound our Rate limiter
	// Consequence: Only the noisiest N clients will be Rate limited.
	cache StringLBCache

	// for testing algorithms involving time we need a mockable time source
	clock ratelimit.Clock

	log ratelimit.Logger

	config *Config
}

type StringLBCache interface {
	Put(key string, value *lbcb) *lbcb
	Get(key string) (result *lbcb, err error)
}

type lbcb struct {
	mutex             sync.Mutex
	availableCapacity leakyBucketAccessCost
	timeOfLastAccess  time.Time
}

func (cb *lbcb) accessAttempt(tenantId string, clock ratelimit.Clock, config *Config, accessCost leakyBucketAccessCost) bool {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := clock.Now()
	tdiff := now.Sub(cb.timeOfLastAccess)

	tenancy := config.Tenancy(tenantId)
	microsRefill := tenancy.Rate * float64(tdiff.Microseconds()) / 1_000_000.0
	cb.availableCapacity = math.Min(
		cb.availableCapacity+microsRefill,
		tenancy.Burst,
	)

	quotaAvailable := accessCost <= cb.availableCapacity

	if quotaAvailable {
		cb.timeOfLastAccess = now
		cb.availableCapacity -= accessCost
	}

	return quotaAvailable
}
