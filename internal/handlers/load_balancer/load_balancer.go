package load_balancer

import (
	bp "github.com/nogavadu/load_balancer/internal/lib/backends_pool"
	"log/slog"
	"net/http"
)

func New(backendsPool *bp.BackendsPool, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		backend := backendsPool.GetNext()
		if backend == nil {
			logger.Error("no available backend")
			return
		}

		backend.ReverseProxy.ServeHTTP(w, r)
	}
}
