package load_balancer

import (
	"github.com/nogavadu/load_balancer/internal/lib/response"
	"log/slog"
	"net/http"
)

func NewHandler(backendsPool *BackendsPool, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		backend := backendsPool.GetNext()
		if backend == nil {
			logger.Error("no available backend")
			response.Err(w, "no available backend", http.StatusServiceUnavailable)
			return
		}

		backend.ReverseProxy.ServeHTTP(w, r)
	}
}
