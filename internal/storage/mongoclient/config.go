package mongoclient

import (
	"fmt"
	"time"

	"github.com/muzzapp/date-api/internal/config"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Config struct {
	MongoDBUri            string        `envconfig:"MONGODB_URI" default:"mongodb://localhost:27017"`
	MongoDBDatabase       string        `envconfig:"MONGODB_DATABASE" default:"date"`
	MongoDBConnectTimeout time.Duration `envconfig:"MONGODB_CONNECT_TIMEOUT" default:"5s"`
}

type Options struct {
	conf *Config
}

func (o Options) validate() error {
	if o.conf.MongoDBDatabase == "" {
		return ErrDbNotSet
	}
	if o.conf.MongoDBUri == "" {
		return ErrUriNotSet
	}
	return nil
}

type Option func(options *Options)

func WithDatabaseName(databaseName string) Option {
	return func(options *Options) {
		options.conf.MongoDBDatabase = databaseName
	}
}

func defaultConfig() (*Options, error) {
	o := &Options{}
	c := &Config{}
	if err := config.Load(c); err != nil {
		return nil, err
	}
	o.conf = c
	return o, nil
}

func getConfig(opts ...Option) (*Options, error) {
	conf, err := defaultConfig()
	if err != nil {
		return nil, fmt.Errorf("couldn't get redis config from environment %w", err)
	}
	for _, opt := range opts {
		opt(conf)
	}
	if err = conf.validate(); err != nil {
		return nil, fmt.Errorf("config was invalid %w", err)
	}
	return conf, nil
}

// clientOptions configures the options for connecting to mongo
func clientOptions(config *Config) *options.ClientOptions {
	opts := options.Client()

	opts.ApplyURI(config.MongoDBUri).SetServerAPIOptions(options.ServerAPI(options.ServerAPIVersion1))
	return opts
}
