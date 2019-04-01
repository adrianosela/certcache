package certcache

// Implementation of the autocert.Cache interface as per
// https://godoc.org/golang.org/x/crypto/acme/autocert#Cache

import (
	"context"
	"log"

	"golang.org/x/crypto/acme/autocert"
)

// Functional allows the user to use functions to define a cert cache.
// If we have the get function always return an autocert.ErrCacheMiss error,
// we can use this cert cache for testing next cache layer's preconditions,
// or simply logging events (see Newlogger() function)
type Functional struct {
	get func(context.Context, string) ([]byte, error)
	put func(context.Context, string, []byte) error
	del func(context.Context, string) error
}

// NewFunctional is the constructor for a functional Cert Cache
func NewFunctional(
	get func(context.Context, string) ([]byte, error),
	put func(context.Context, string, []byte) error,
	del func(context.Context, string) error,
) *Functional {
	return &Functional{
		get: get,
		put: put,
		del: del,
	}
}

// NewLogger is the constructor for a Functional cert cache implementation
// which does nothing other than log events
func NewLogger() *Functional {
	return NewFunctional(
		func(ctx context.Context, key string) ([]byte, error) {
			log.Printf("[CERT CACHE] getting key %s", key)
			return nil, autocert.ErrCacheMiss
		},
		func(ctx context.Context, key string, data []byte) error {
			log.Printf("[CERT CACHE] putting key %s", key)
			return nil
		},
		func(ctx context.Context, key string) error {
			log.Printf("[CERT CACHE] deleting key %s", key)
			return nil
		},
	)
}

// Get returns a certificate data for the specified key.
// If there's no such key, Get returns ErrCacheMiss.
func (f *Functional) Get(ctx context.Context, key string) ([]byte, error) {
	return f.get(ctx, key)
}

// Put stores the data in the cache under the specified key.
// Underlying implementations may use any data storage format,
// as long as the reverse operation, Get, results in the original data.
func (f *Functional) Put(ctx context.Context, key string, data []byte) error {
	return f.put(ctx, key, data)
}

// Delete removes a certificate data from the cache under the specified key.
// If there's no such key in the cache, Delete returns nil.
func (f *Functional) Delete(ctx context.Context, key string) error {
	return f.del(ctx, key)
}
