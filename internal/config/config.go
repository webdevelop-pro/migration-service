package config

import (
	"sync"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Database     database
	Http         http
	LogLevel     string `envconfig:"LOG_LEVEL" default:"debug"`
	LogConsole   bool   `envconfig:"LOG_CONSOLE" default:"false"`
	GitHash      string
	MigrationDir string `split_words:"true" required:"true"`
	ForceApply   bool   `split_words:"true"`
}

type database struct {
	Database       string `envconfig:"DB_DATABASE" required:"true"`
	User           string `envconfig:"DB_USER" required:"true"`
	Password       string `envconfig:"DB_PASSWORD" required:"true"`
	Host           string `envconfig:"DB_HOST" default:"localhost"`
	Port           uint16 `envconfig:"DB_PORT" default:"5432"`
	MaxConnections int    `envconfig:"DB_MAX_CONNECTIONS" default:"100"`
}

type http struct {
	Host string `envconfig:"HOST"`
	Port string `envconfig:"PORT"`
}

var cfg *Config
var mu sync.Mutex

// NewConfig return config from environment
func NewConfig() (*Config, error) {
	c := Config{}
	err := envconfig.Process("", &c)
	return &c, err
}

// GetConfig create or return existing config
func GetConfig() *Config {
	var err error
	mu.Lock()
	defer mu.Unlock()
	if cfg == nil {
		cfg, err = NewConfig()
		if err != nil {
			panic(err)
		}
	}
	return cfg
}
