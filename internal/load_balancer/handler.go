package load_balancer

import (
	"log/slog"
	"net/http"
)

func New(backendsPool *BackendsPool, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		backend := backendsPool.GetNext()
		if backend == nil {
			logger.Error("no available backend")
			return
		}

		backend.ReverseProxy.ServeHTTP(w, r)
	}
}
