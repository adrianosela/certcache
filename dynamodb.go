package certcache

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"golang.org/x/crypto/acme/autocert"
)

// DynamoDB represents a DynamoDB implementation of autocert.Cache
type DynamoDB struct {
	client     dynamodbiface.DynamoDBAPI
	table      string
	priKeyname string
}

type obj struct {
	Data string `json:"Data"`
}

const (
	defaultDynamoDBTableName  = "certcache"
	defaultDynamoDBRegion     = "us-west-2"
	defaultDynamoDBPriKeyName = "id"
	defaultDynamoDBTimeout    = 10 * time.Second
)

// NewDynamoDB returns a DynamoDB certificate cache
func NewDynamoDB(credentials *credentials.Credentials, region, table string) *DynamoDB {
	if table == "" {
		table = defaultDynamoDBTableName
	}
	if region == "" {
		region = defaultDynamoDBRegion
	}
	svc := dynamodb.New(session.New(), &aws.Config{
		Credentials: credentials,
		Region:      aws.String(region),
		HTTPClient:  &http.Client{Timeout: defaultS3CertCacheTimeout},
	})
	return &DynamoDB{
		client:     svc,
		table:      table,
		priKeyname: defaultDynamoDBPriKeyName,
	}
}

// Get returns a certificate data for the specified key.
// If there's no such key, Get returns ErrCacheMiss.
func (ddb *DynamoDB) Get(ctx context.Context, key string) ([]byte, error) {
	result, err := ddb.client.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(ddb.table),
		Key:       buildPrimaryKey(ddb.priKeyname, key),
	})
	if err != nil || result == nil {
		return nil, fmt.Errorf("could not fetch object %s: %s", key, err)
	}
	if _, ok := result.Item[ddb.priKeyname]; !ok {
		return nil, autocert.ErrCacheMiss
	}
	var o obj
	if err = dynamodbattribute.UnmarshalMap(result.Item, &o); err != nil {
		return nil, fmt.Errorf("could not unmarshal object %s: %s", key, err)
	}
	return []byte(o.Data), nil
}

// Put stores the data in the cache under the specified key.
// Underlying implementations may use any data storage format,
// as long as the reverse operation, Get, results in the original data.
func (ddb *DynamoDB) Put(ctx context.Context, key string, data []byte) error {
	if _, err := ddb.client.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(ddb.table),
		Item: map[string]*dynamodb.AttributeValue{
			ddb.priKeyname: {S: aws.String(string(data))},
		},
	}); err != nil {
		return fmt.Errorf("could not store object %s: %s", key, err)
	}
	return nil
}

// Delete removes a certificate data from the cache under the specified key.
// If there's no such key in the cache, Delete returns nil.
func (ddb *DynamoDB) Delete(ctx context.Context, key string) error {
	if _, err := ddb.client.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String(ddb.table),
		Key:       buildPrimaryKey(ddb.priKeyname, key),
	}); err != nil {
		return fmt.Errorf("could not delete object %s: %s", key, err)
	}
	return nil
}

func buildPrimaryKey(primaryKey, objectKey string) map[string]*dynamodb.AttributeValue {
	return map[string]*dynamodb.AttributeValue{
		primaryKey: {
			S: aws.String(objectKey),
		},
	}
}
