package load_balancer

import (
	"log/slog"
	"net/http"
)

func NewHandler(backendsPool *BackendsPool, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		backend := backendsPool.GetNext()
		if backend == nil {
			logger.Error("no available backend")
			http.Error(w, "service not available", http.StatusServiceUnavailable)
			return
		}

		backend.ReverseProxy.ServeHTTP(w, r)
	}
}
