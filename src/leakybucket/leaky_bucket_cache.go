package leakybucket

import (
	caches "github.com/npxcomplete/caches/src"
	"sync"
)

func NewStringLBCBCache(capacity int) privateStringLBCBCache {
	return WrapStringLBCBCache(caches.NewLRUCache(capacity))
}

func WrapStringLBCBCache(cache caches.Interface) privateStringLBCBCache {
	return privateStringLBCBCache{
		generic: cache,
		mutex:   &sync.RWMutex{},
	}
}

// genny is case-sensitive even though this has other meanings in go, so we prefix the intent.
type privateStringLBCBCache struct {
	mutex   *sync.RWMutex
	generic caches.Interface
}

// see caches.Interface for contract
func (cache privateStringLBCBCache) Put(key string, value *lbcb) *lbcb {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	result, _ := cache.generic.Put(key, value).(*lbcb)
	return result
}

// see caches.Interface for contract
func (cache privateStringLBCBCache) Get(key string) (result *lbcb, err error) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	value, err := cache.generic.Get(key)
	result, _ = value.(*lbcb)
	return
}
