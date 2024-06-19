package config

import (
	"github.com/kelseyhightower/envconfig"
)

func Load(config interface{}) error {
	return envconfig.Process("", config)
}
