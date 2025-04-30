package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/nogavadu/load_balancer/internal/config"
	"github.com/nogavadu/load_balancer/internal/lib/logger/sl"
	lb "github.com/nogavadu/load_balancer/internal/load_balancer"
	"github.com/nogavadu/load_balancer/internal/token_bucket"
	"github.com/nogavadu/load_balancer/pkg/pretty_slog"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	logger := slog.New(pretty_slog.NewHandler(nil))
	logger.Info("logger initialized")

	cfg, err := config.Read(os.Getenv("CONFIG_PATH"))
	if err != nil {
		logger.Error("failed to load config", sl.Err(err))
		os.Exit(1)
	}
	if len(cfg.BackendsPool) == 0 {
		logger.Error("empty backends pool")
		os.Exit(1)
	}

	go cfg.WatchConfig(logger)

	logger.Info("config initialized")

	var backendsPool lb.BackendsPool

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

	go backendsPool.WatchBackends(logger)

	tokenBucketMiddleware := token_bucket.New()

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HTTPServer.Port),
		Handler: tokenBucketMiddleware(lb.NewHandler(&backendsPool, logger)),
	}

	go func() {
		logger.Info("load balancer started",
			slog.Int("port", cfg.HTTPServer.Port),
			slog.String("backends", fmt.Sprintf("%v", cfg.BackendsPool)),
		)

		if err = server.ListenAndServe(); err != nil {
			logger.Error("failed to start server", sl.Err(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit

	logger.Info(fmt.Sprintf("received signal: %s. Shutting down...", sig))

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err = server.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", sl.Err(err))
		if errors.Is(err, context.DeadlineExceeded) {
			log.Println("forcing shutdown due to timeout")
			server.Close()
		}
	}

	logger.Info("server turned down")
}
