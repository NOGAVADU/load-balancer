package config

import (
	"errors"
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"time"
)

var (
	errEmptyConfigPath    = errors.New("empty config path")
	errConfigFileNotFound = errors.New("config file not found")
	errInvalidConfig      = errors.New("invalid config")
)

type HTTPServer struct {
	Address     string        `yaml:"address"`
	Timeout     time.Duration `yaml:"timeout"`
	IdleTimeout time.Duration `yaml:"idle_timeout"`
}

type Config struct {
	HTTPServer   HTTPServer `yaml:"http_server"`
	BackendsPool []string   `yaml:"backends_pool"`
}

func Load(path string) (*Config, error) {
	if path == "" {
		return nil, errEmptyConfigPath
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, errConfigFileNotFound
	}

	var cfg Config
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		return nil, fmt.Errorf("%w: %w", errInvalidConfig, err)
	}

	return &cfg, nil
}
