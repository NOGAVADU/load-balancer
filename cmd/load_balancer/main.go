package main

import (
	"fmt"
	"github.com/nogavadu/load_balancer/internal/config"
	"github.com/nogavadu/load_balancer/internal/lib/logger/sl"
	lb "github.com/nogavadu/load_balancer/internal/load_balancer"
	"github.com/nogavadu/load_balancer/pkg/pretty_slog"
	"log/slog"
	"net/http"
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
	if len(cfg.BackendsPool) == 0 {
		logger.Error("empty backends pool")
		os.Exit(1)
	}
	logger.Info("config initialized")

	var backendsPool lb.BackendsPool

	logger.Info("backends configuration started")
	var errCounter int
	for _, backend := range cfg.BackendsPool {
		err = backendsPool.AddBackend(backend, logger)
		if err != nil {
			logger.Error(fmt.Sprintf("failed to parse backend URL: %s", backend), sl.Err(err))
			errCounter++
		}
	}
	if errCounter == len(cfg.BackendsPool) {
		logger.Error("failed to configure backends")
		os.Exit(1)
	} else if errCounter != 0 {
		logger.Error(fmt.Sprintf("backends configurated with %d errors", errCounter), sl.Err(err))
	} else {
		logger.Info("all backends configurated")
	}

	go lb.StartBackendsChecking(&backendsPool, logger)

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HTTPServer.Port),
		Handler: lb.New(&backendsPool, logger),
	}

	logger.Info("load balancer started", slog.Int("port", cfg.HTTPServer.Port))
	if err = server.ListenAndServe(); err != nil {
		logger.Error("failed to start server", sl.Err(err))
	}
}
