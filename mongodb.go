package certcache

// Implementation of the autocert.Cache interface as per
// https://godoc.org/golang.org/x/crypto/acme/autocert#Cache

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"gopkg.in/mgo.v2/bson"
)

// MongoDB represents a MongoDB implementation of autocert.Cache
type MongoDB struct {
	conn     *mongo.Database
	ctxt     context.Context
	collname string
	timeout  time.Duration
}

type doc struct {
	_id  string `bson:"_id"`
	data []byte `bson:"data"`
}

const (
	defaultMongoCertCacheCollectionName = "certcache"
	defaultMongoCertCacheDBName         = "certcache"
	defaultMongoCertCacheTimeout        = time.Second * 10
)

// NewMongoDB returns a Mongo cache given a mongodb connection string
// e.g. fmt.Sprintf("mongodb://%s:%s@%s/%s", username, password, host, db)
func NewMongoDB(uri string) *MongoDB {
	if uri == "" {
		log.Fatalf("failed to connect to Mongo. Must specify connection string")
	}
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("failed to create Mongo client: %s", err)
	}
	ctxt := context.Background()
	if err = client.Connect(ctxt); err != nil {
		log.Fatalf("failed to connect to Mongo server: %s", err)
	}
	if err = client.Ping(ctxt, readpref.Primary()); err != nil {
		log.Fatalf("failed to reach Mongo server: %s", err)
	}
	return &MongoDB{
		ctxt:     ctxt,
		conn:     client.Database(defaultMongoCertCacheDBName),
		collname: defaultMongoCertCacheCollectionName,
		timeout:  defaultMongoCertCacheTimeout,
	}
}

// Get returns a certificate data for the specified key.
// If there's no such key, Get returns ErrCacheMiss.
func (mgo *MongoDB) Get(ctx context.Context, key string) ([]byte, error) {
	var d doc
	err := mgo.conn.Collection(mgo.collname).FindOne(ctx, bson.M{"_id": key}).Decode(&d)
	if err != nil {
		return nil, fmt.Errorf("failed to get %s from Mongo: %s", key, err)
	}
	return d.data, nil
}

// Put stores the data in the cache under the specified key.
// Underlying implementations may use any data storage format,
// as long as the reverse operation, Get, results in the original data.
func (mgo *MongoDB) Put(ctx context.Context, key string, data []byte) error {
	doc := bson.D{
		bson.DocElem{Name: "_id", Value: key},
		bson.DocElem{Name: "data", Value: data},
	}
	if _, err := mgo.conn.Collection(mgo.collname).InsertOne(mgo.ctxt, doc); err != nil {
		return fmt.Errorf("failed to store %s in Mongo: %s", key, err)
	}
	return nil
}

// Delete removes a certificate data from the cache under the specified key.
// If there's no such key in the cache, Delete returns nil.
func (mgo *MongoDB) Delete(ctx context.Context, key string) error {
	doc := bson.D{bson.DocElem{Name: "_id", Value: key}}
	if _, err := mgo.conn.Collection(mgo.collname).DeleteOne(ctx, doc); err != nil {
		return fmt.Errorf("failed to delete %s from Mongo: %s", key, err)
	}
	return nil
}
