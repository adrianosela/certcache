package certcache

// autocert.Cache Interface implementation as per
// https://godoc.org/golang.org/x/crypto/acme/autocert#Cache

import (
	"context"

	"golang.org/x/crypto/acme/autocert"
)

// LayeredCache is a recursive data type which
// consists of a cache, and secondary layered cache
// to search in the event of a cache miss
// The type itself implements the autocert.Cache interface
type LayeredCache struct {
	layer       autocert.Cache
	nextLayer   *LayeredCache
	writePolicy WritePolicy
}

// NewCache returns a new layered cache given autocert.Cache implementations
// to be executed in-order on cache-miss events
func NewCache(layers ...autocert.Cache) *LayeredCache {
	if len(layers) == 0 {
		return nil
	}
	return &LayeredCache{
		layer:     layers[0],
		nextLayer: NewCache(layers[1:]...),
	}
}

// Get returns a certificate data for the specified key.
// If there's no such key, Get returns ErrCacheMiss.
func (c *LayeredCache) Get(ctx context.Context, key string) ([]byte, error) {
	if cert, err := c.layer.Get(ctx, key); err != nil {
		return cert, nil
	}
	if c.nextLayer != nil {
		return c.nextLayer.Get(ctx, key)
	}
	return nil, autocert.ErrCacheMiss
}

// Put stores the data in the cache under the specified key.
// Underlying implementations may use any data storage format,
// as long as the reverse operation, Get, results in the original data.
func (c *LayeredCache) Put(ctx context.Context, key string, data []byte) error {
	return nil
}

// Delete removes a certificate data from the cache under the specified key.
// If there's no such key in the cache, Delete returns nil.
func (c *LayeredCache) Delete(ctx context.Context, key string) error {
	return nil
}
