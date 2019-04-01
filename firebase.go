package certcache

// Implementation of the autocert.Cache interface as per
// https://godoc.org/golang.org/x/crypto/acme/autocert#Cache

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
	"golang.org/x/crypto/acme/autocert"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// FirestoreCertCache is a Google Firestore implementation of autocert.Cache
type FirestoreCertCache struct {
	collectionName string // firestore has "collections" with "documents"
	ctxt           context.Context
	client         *firestore.Client
}

const (
	defaultCertCacheCollectionName = "certcache"
)

// NewFirestoreCertCache is the default constructor for a FirestoreCertCache
func NewFirestoreCertCache(credsPath, projectID string) *FirestoreCertCache {
	return NewFirestoreCertCacheInCollection(credsPath, projectID, defaultCertCacheCollectionName)
}

// NewFirestoreCertCacheInCollection is a constructor for a FirestoreCertCache
// with a custom Firestore Collection name
func NewFirestoreCertCacheInCollection(credsPath, projectID, certsCollectionName string) *FirestoreCertCache {
	cntxt := context.Background()
	creds := option.WithCredentialsFile(credsPath)
	cl, err := firestore.NewClient(cntxt, projectID, creds)
	if err != nil {
		log.Fatalf("[FIRESTORE] failed to initialize firestore client: %s", err)
	}
	return &FirestoreCertCache{
		ctxt:           cntxt,
		collectionName: certsCollectionName,
		client:         cl,
	}
}

type format struct {
	data string `firestore:"data"`
}

// Get returns a certificate data for the specified key.
// If there's no such key, Get returns ErrCacheMiss.
func (fcc *FirestoreCertCache) Get(ctx context.Context, key string) ([]byte, error) {
	docSnapshot, err := fcc.client.Collection(fcc.collectionName).Doc(key).Get(fcc.ctxt)
	if err != nil {
		if grpc.Code(err) == codes.NotFound {
			return nil, autocert.ErrCacheMiss
		}
		return nil, err
	}
	var doc format
	if err := docSnapshot.DataTo(&doc); err != nil {
		return nil, err
	}
	return []byte(doc.data), nil
}

// Put stores the data in the cache under the specified key.
// Underlying implementations may use any data storage format,
// as long as the reverse operation, Get, results in the original data.
func (fcc *FirestoreCertCache) Put(ctx context.Context, key string, data []byte) error {
	newDocRef := fcc.client.Collection(fcc.collectionName).Doc(key)
	_, err := newDocRef.Create(fcc.ctxt, &format{data: string(data)})
	return err
}

// Delete removes a certificate data from the cache under the specified key.
// If there's no such key in the cache, Delete returns nil.
func (fcc *FirestoreCertCache) Delete(ctx context.Context, key string) error {
	_, err := fcc.client.Collection(fcc.collectionName).Doc(key).Delete(fcc.ctxt)
	return err
}
