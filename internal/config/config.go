package config

import (
	"errors"
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/nogavadu/load_balancer/internal/lib/logger/sl"
	"log/slog"
	"os"
	"sync"
	"time"
)

var (
	errEmptyConfigPath    = errors.New("empty config path")
	errConfigFileNotFound = errors.New("config file not found")
	errInvalidConfig      = errors.New("invalid config")
)

type Config struct {
	mux          *sync.RWMutex
	BackendsPool []string `yaml:"backends_pool"`
}

func Read() (*Config, error) {
	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		return nil, errEmptyConfigPath
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, errConfigFileNotFound
	}

	cfg := &Config{
		mux: &sync.RWMutex{},
	}
	if err := cleanenv.ReadConfig(path, cfg); err != nil {
		return nil, fmt.Errorf("%w: %w", errInvalidConfig, err)
	}

	return cfg, nil
}

func (c *Config) UpdateBackendsPool() error {
	c.mux.Lock()
	defer c.mux.Unlock()

	cfg, err := Read()
	if err != nil {
		return fmt.Errorf("%w: %w", errInvalidConfig, err)
	}

	c.BackendsPool = cfg.BackendsPool

	return nil
}

func (c *Config) WatchConfig(logger *slog.Logger) {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		err := c.UpdateBackendsPool()
		if err != nil {
			logger.Error("failed to update backends pool", sl.Err(err))
		} else {
			logger.Info("backends pool has been updated")
		}
	}
}
