package main

import (
	"github.com/nogavadu/load_balancer/pkg/pretty_slog"
	"log/slog"
)

func main() {
	opts := &slog.HandlerOptions{Level: slog.LevelDebug}
	logger := slog.New(pretty_slog.NewHandler(opts))
	logger.Debug("logger initialized")
}
