package main

import (
	"github.com/nogavadu/load_balancer/internal/config"
	"github.com/nogavadu/load_balancer/internal/lib/logger/sl"
	"github.com/nogavadu/load_balancer/pkg/pretty_slog"
	"log/slog"
	"os"
)

const (
	configPath = "./config/config.yaml"
)

func main() {
	logger := slog.New(pretty_slog.NewHandler(nil))
	logger.Info("logger initialized")

	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Error("failed to load config", sl.Err(err))
		os.Exit(1)
	}
	logger.Info("config initialized")

	_ = cfg
}
