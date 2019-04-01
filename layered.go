package certcache

// autocert.Cache Interface implementation as per
// https://godoc.org/golang.org/x/crypto/acme/autocert#Cache

import (
	"context"
	"errors"

	"golang.org/x/crypto/acme/autocert"
)

// LayeredCache is an implementation of the autocert.Cache interface
// The behavior of the cache consists in checking itself for hits, and having
// a fall back storage layer to search in the event of a cache miss
type LayeredCache struct {
	layer       autocert.Cache
	nextLayer   *LayeredCache
	writePolicy WritePolicy
}

// WritePolicy determines the order in which the layered cache executes Put
type WritePolicy string

const (
	// PolicyWriteDeepFirst will write to caches starting from the last layer
	// provided. This is the default behavior, as the more common use case
	// will be for the last layer to be the most persistent e.g. DB call
	PolicyWriteDeepFirst = WritePolicy("DEEP_FIRST")
	// PolicyWriteShallowFirst will write to caches starting from the top,
	// often least persistent, layer e.g. a struct in process heap
	PolicyWriteShallowFirst = WritePolicy("SHALLOW_FIRST")
)

// NewLayered returns a new layered cache given autocert.Cache implementations
// this is the default constructor
func NewLayered(layers ...autocert.Cache) *LayeredCache {
	if len(layers) == 0 {
		return nil
	}
	return &LayeredCache{
		layer:       layers[0],
		nextLayer:   NewLayered(layers[1:]...),
		writePolicy: PolicyWriteDeepFirst,
	}
}

// NewLayeredWithPolicy returns a new layered cache and allows the user
// to specify the write policy
func NewLayeredWithPolicy(wp WritePolicy, layers ...autocert.Cache) *LayeredCache {
	if len(layers) == 0 {
		return nil
	}
	return &LayeredCache{
		layer:       layers[0],
		nextLayer:   NewLayeredWithPolicy(wp, layers[1:]...),
		writePolicy: wp,
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
	switch c.writePolicy {
	case PolicyWriteDeepFirst:
		if c.nextLayer != nil {
			if err := c.nextLayer.Put(ctx, key, data); err != nil {
				return err
			}
		}
		return c.layer.Put(ctx, key, data)
	case PolicyWriteShallowFirst:
		if err := c.layer.Put(ctx, key, data); err != nil {
			return err
		}
		if c.nextLayer != nil {
			return c.nextLayer.Put(ctx, key, data)
		}
		return nil
	default:
		return errors.New("unrecognized write policy")
	}
}

// Delete removes a certificate data from the cache under the specified key.
// If there's no such key in the cache, Delete returns nil.
func (c *LayeredCache) Delete(ctx context.Context, key string) error {
	switch c.writePolicy {
	case PolicyWriteDeepFirst:
		if c.nextLayer != nil {
			if err := c.nextLayer.Delete(ctx, key); err != nil {
				return err
			}
		}
		return c.layer.Delete(ctx, key)
	case PolicyWriteShallowFirst:
		if err := c.layer.Delete(ctx, key); err != nil {
			return err
		}
		if c.nextLayer != nil {
			return c.nextLayer.Delete(ctx, key)
		}
		return nil
	default:
		return errors.New("unrecognized write policy")
	}
}
