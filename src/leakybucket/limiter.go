package leakybucket

import (
	"context"
	caches "github.com/npxcomplete/caches/src"
	"github.com/npxcomplete/caches/src/thread_safe"
	"github.com/npxcomplete/http-rate-limit/src"
	"sync"
	"time"
)

type RateLimiter interface {
	LeakyBucket
	// AttemptAccess
	//
	// Given one or more identifiers find all
	// capacity allocations.
	//
	// If NO allocations are found the method returns false.
	//
	// If ALL allocations have sufficient capacity available,
	// then ALL allocations are decremented and the function
	// returns true.
	//
	// if ANY allocation has insufficient capacity available,
	// then ALL allocations are STILL decremented and the function
	// returns false.
	//
	// We decrement capacity even if the request will not be serviced
	// such that the client is forced to actively manage its own rate limiting.
	// Mindless clients can not DoS the service until capacity becomes available.
	//
	// Note that in a distributed environment we should not allow the locally
	// spent capacity to drop below (-1 * burst) for any given allocation.
	// by implication this will keep the global burst from dropping below
	// (-N * burst), where N is the number of hosts in the peer group.
	AttemptAccess(requestCost uint32, userIds ...string) bool
}

type TenantLimit struct {
	Rate  AccessCost
	Burst AccessCost
}

type Config struct {
	TenantCapacity int
	Tenancy        func(tenant string) *TenantLimit

	// method to establish client connections
	InitPeerGroup ClientFactory

	// defaults to port 49238
	Port uint16

	backgroundSync LBTicker
	clock          ratelimit.Clock
}

var _ LeakyBucket = &leakyBucketRateLimiter{}

func NewRateLimiter(config Config) *leakyBucketRateLimiter {
	var backgroundSync LBTicker
	if config.backgroundSync == nil {
		backgroundSync = NewTicker(10 * time.Second)
	} else {
		backgroundSync = config.backgroundSync
	}

	var clock ratelimit.Clock
	if config.clock == nil {
		clock = ratelimit.HardwareClock{}
	} else {
		clock = config.clock
	}

	cache := thread_safe.NewGuardedCache[string, *controlBlock](
		caches.NewLRUCache[string, *controlBlock](config.TenantCapacity),
	)
	limiter := &leakyBucketRateLimiter{
		// TODO https://github.com/Yiling-J/theine-go
		//cache:       caches.New2Q[string, *controlBlock](config.TenantCapacity),
		cache:  cache,
		clock:  clock,
		log:    ratelimit.StdOutLogger{},
		config: &config,
	}

	limiter.refreshPeers()

	for _, peer := range limiter.peerClients {
		join, err := peer.Join(context.Background(), &JoinRequest{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		})

		if err != nil {
			limiter.log.Error(err.Error())
		}

		for id, capa := range join.CapacityBook.AvailableCapacity {
			limiter.primeCache(id, capa)
		}
	}

	go limiter.periodicBlast(backgroundSync)

	return limiter

}

func (limiter *leakyBucketRateLimiter) periodicBlast(backgroundSync LBTicker) {
	for range backgroundSync.Channel() {
		// TODO every tick we call sync on each peer, careful not to call ourselves
		spentCapacity := make(map[string]uint32)

		limiter.cache.Range(func(key string, value *controlBlock) bool {
			spentCapacity[key] = value.getAndClearSpentCapacity()

			return true
		})

		for _, tClient := range limiter.peerClients {
			_, err := tClient.Sync(context.Background(), &SyncRequest{
				CapacityChange: &CapacityChange{
					SpentCapacity: spentCapacity,
				},
			})

			if err != nil {
				limiter.log.Error(err.Error())
			}
		}
	}
}

func (limiter *leakyBucketRateLimiter) decreaseAvailableCapacity(accessCost uint32, tenantId string) bool {
	tenancy := limiter.config.Tenancy(tenantId)
	if tenancy == nil {
		return false
	}

	cb, err := limiter.cache.Get(tenantId)
	if err == caches.MissingValueError {
		cb = &controlBlock{
			mutex:             sync.Mutex{},
			availableCapacity: Capacity(tenancy.Burst),
			timeOfLastAccess:  limiter.clock.Now(),
		}
		limiter.cache.Put(tenantId, cb)
	} else if err != nil {
		limiter.log.Error(err.Error())
		return false
	}

	// All identities get tested.
	cb.decreaseAvailableCapacity(accessCost)
	return true
}

func (limiter *leakyBucketRateLimiter) AttemptAccess(accessCost uint32, tenantIds ...string) bool {
	tenancies := make(map[string]*TenantLimit)

	for _, tenantId := range tenantIds {
		tenancy := limiter.config.Tenancy(tenantId)
		if tenancy != nil {
			tenancies[tenantId] = tenancy
		}
	}

	if len(tenancies) == 0 {
		return false
	}

	allowAccess := true

	for id, limits := range tenancies {
		cb, err := limiter.cache.Get(id)
		if err == caches.MissingValueError {
			cb = &controlBlock{
				mutex:             sync.Mutex{},
				availableCapacity: Capacity(limits.Burst),
				timeOfLastAccess:  limiter.clock.Now(),
			}
			limiter.cache.Put(id, cb)
		} else if err != nil {
			allowAccess = false
			limiter.log.Error(err.Error())
		}

		// All identities get tested.
		if !cb.accessAttempt(limits, limiter.clock, accessCost) {
			allowAccess = false
		}
	}

	return allowAccess
}

func (limiter *leakyBucketRateLimiter) Join(ctx context.Context, req *JoinRequest) (*JoinResponse, error) {
	availableCapacity := make(map[string]int64)
	resp := &JoinResponse{
		CapacityBook: &CapacityBook{
			AvailableCapacity: availableCapacity,
		},
	}

	limiter.cache.Range(func(id string, value *controlBlock) bool {
		availableCapacity[id] = value.availableCapacity
		// dont stop iterating early
		return true
	})

	go limiter.refreshPeers()

	return resp, nil
}

func (limiter *leakyBucketRateLimiter) Sync(ctx context.Context, req *SyncRequest) (*SyncResponse, error) {
	for tenantId, capacity := range req.CapacityChange.SpentCapacity {
		limiter.decreaseAvailableCapacity(capacity, tenantId)
	}

	resp := &SyncResponse{Timestamp: time.Now().UTC().Format(time.RFC3339)}
	return resp, nil
}

func (limiter *leakyBucketRateLimiter) refreshPeers() {
	if limiter.config.InitPeerGroup == nil {
		limiter.peerClients = []LeakyBucket{}
		return
	}

	limiter.peerClients = limiter.config.InitPeerGroup.NewClients()
}

func (limiter *leakyBucketRateLimiter) primeCache(id string, capa Capacity) {
	limiter.cache.Put(id, &controlBlock{
		mutex:             sync.Mutex{},
		availableCapacity: capa,
		timeOfLastAccess:  limiter.clock.Now(),
		spentCapacity:     0,
	})
}

// AccessCost is only ever a positive number.
// However, the available capacity can become negative.
type AccessCost = uint32
type Capacity = int64

type leakyBucketRateLimiter struct {
	// Use a fixed capacity cache to memory bound our Rate limiter
	// Consequence: Only the noisiest N clients will be Rate limited.
	cache caches.Interface[string, *controlBlock]

	// for testing algorithms involving time we need a mockable time source
	clock ratelimit.Clock

	log ratelimit.Logger

	config *Config

	peerClients []LeakyBucket
}
