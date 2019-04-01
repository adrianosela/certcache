package certcache

import "context"

// FirestoreCertCache is a Google Firestore implementation of autocert.Cache
type FirestoreCertCache struct {
}

// Get returns a certificate data for the specified key.
// If there's no such key, Get returns ErrCacheMiss.
func (fcc *FirestoreCertCache) Get(ctx context.Context, key string) ([]byte, error) {
	return nil, nil
}

// Put stores the data in the cache under the specified key.
// Underlying implementations may use any data storage format,
// as long as the reverse operation, Get, results in the original data.
func (fcc *FirestoreCertCache) Put(ctx context.Context, key string, data []byte) error {
	return nil
}

// Delete removes a certificate data from the cache under the specified key.
// If there's no such key in the cache, Delete returns nil.
func (fcc *FirestoreCertCache) Delete(ctx context.Context, key string) error {
	return nil

}
