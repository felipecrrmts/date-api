package mongoclient

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

// GetDatabase connects and returns a new mongodb database (service name is used by default but this can
// be overridden by MongoClientConfig.MongoDatabase)
func GetDatabase(opts ...Option) (*mongo.Database, error) {
	cfg, err := getConfig(opts...)
	if err != nil {
		return nil, err
	}
	client, err := getClient(cfg.conf)
	if err != nil {
		return nil, err
	}

	return client.Database(cfg.conf.MongoDBDatabase), nil
}

func getClient(conf *Config) (*mongo.Client, error) {
	if conf == nil {
		return nil, ErrEmptyConfig
	}
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, conf.MongoDBConnectTimeout)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions(conf))
	if err != nil {
		return nil, err
	}

	// Check to make sure connection is usable
	ctx, cancel = context.WithTimeout(ctx, conf.MongoDBConnectTimeout)
	defer cancel()
	if err = client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return client, err
}
