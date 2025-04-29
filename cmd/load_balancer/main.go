package main

import (
	"fmt"
	"github.com/nogavadu/load_balancer/internal/config"
	"github.com/nogavadu/load_balancer/internal/lib/logger/sl"
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

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HTTPServer.Port),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
	}

	logger.Info("load balancer started", slog.Int("port", cfg.HTTPServer.Port))
	if err = server.ListenAndServe(); err != nil {
		logger.Error("failed to start server", sl.Err(err))
	}
}
