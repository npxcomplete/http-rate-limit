package ratelimit

import (
	"github.com/npxcomplete/caches/src"
)

func NewLRUStringSWCBCache(capacity int) LRUStringSWCBCache {
	return swcbCache{caches.NewKeyTypeValueTypeLRU(capacity)}
}

type LRUStringSWCBCache interface {
	// add the control block to the cache
	// return whatever control block was evicted if any
	Put(key string, value *swcb) *swcb
	Get(key string) (*swcb, error)
}

type swcbCache struct {
	generic caches.KeyTypeValueTypeCache
}

func (cache swcbCache) Put(key string, value *swcb) *swcb {
	swcb, _ := cache.generic.Put(key, value).(*swcb)
	return swcb
}

func (cache swcbCache) Get(key string) (ret *swcb, err error) {
	value, err := cache.generic.Get(key)
	ret, _ = value.(*swcb)
	return
}
