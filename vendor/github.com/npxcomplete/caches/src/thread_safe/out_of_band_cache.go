package thread_safe

import (
	"context"

	"github.com/npxcomplete/caches/src"
)


// This is idiomatic ish, but it's also four to ten times slower than the guarded implementation
func NewOutOfBandCache(ctx context.Context, inner caches.Interface) (safe caches.Interface) {
	universalConstruction := make(chan func())

	safe = &outOfBan{
		generic:               inner,
		universalConstruction: universalConstruction,
	}

	go func(ctx context.Context, inner caches.Interface) {
		for {
			select {
			case f := <-universalConstruction:
				f()
			case <-ctx.Done():
				return
			}
		}
	}(ctx, inner)

	return
}

// genny is case sensitive even though this has other meanings in go, so we prefix the intent.
type outOfBan struct {
	universalConstruction chan func()
	generic               caches.Interface
}

// see caches.Interface for contract
func (cache *outOfBan) Put(key caches.Key, value caches.Value) caches.Value {
	ret := make(chan caches.Value)
	cache.universalConstruction <- func() {
		ret <- cache.generic.Put(key, value)
	}
	return <-ret
}

// see caches.Interface for contract
func (cache *outOfBan) Get(key caches.Key) (result caches.Value, err error) {
	ret := make(chan func() (caches.Value, error))
	cache.universalConstruction <- func() {
		val, err := cache.generic.Get(key)
		ret <- func() (caches.Value, error) { return val, err }
	}
	return (<-ret)()
}
