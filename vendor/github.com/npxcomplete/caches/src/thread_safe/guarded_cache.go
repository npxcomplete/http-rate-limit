package thread_safe

import (
	"sync"

	"github.com/npxcomplete/caches/src"
)

func NewGuardedCache(p caches.Interface) caches.Interface {
	return &guardedLRU{
		generic: p,
	}
}

// genny is case sensitive even though this has other meanings in go, so we prefix the intent.
type  guardedLRU struct {
	mut sync.Mutex
	generic caches.Interface
}

// see caches.Interface for contract
func (cache *guardedLRU) Put(key caches.Key, value caches.Value) caches.Value {
	cache.mut.Lock()
	defer cache.mut.Unlock()
	return cache.generic.Put(key, value)
}

// see caches.Interface for contract
func (cache *guardedLRU)  Get(key caches.Key) (result caches.Value, err error) {
	cache.mut.Lock()
	defer cache.mut.Unlock()
	return cache.generic.Get(key)
}

