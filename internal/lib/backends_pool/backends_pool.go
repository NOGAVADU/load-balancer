package backends_pool

import (
	"fmt"
	"github.com/nogavadu/load_balancer/internal/lib/logger/sl"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
)

type BackendsPool struct {
	backends []*Backend
	current  uint64
}

func (bp *BackendsPool) nextIdx() int {
	return int(atomic.AddUint64(&bp.current, 1) % uint64(len(bp.backends)))
}

func (bp *BackendsPool) GetNext() *Backend {
	next := bp.nextIdx()
	l := len(bp.backends) + next
	for i := next; i < l; i++ {
		idx := i % len(bp.backends)
		if bp.backends[idx].IsAlive() {
			if i != next {
				atomic.StoreUint64(&bp.current, uint64(idx))
			}
			return bp.backends[idx]
		}
	}
	return nil
}

func (bp *BackendsPool) AddBackend(uri string, logger *slog.Logger) error {
	serverUrl, err := url.Parse(uri)
	if err != nil {
		return err
	}

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = serverUrl.Scheme
			req.URL.Host = serverUrl.Host
			req.Host = serverUrl.Host
		},
		ErrorHandler: func(writer http.ResponseWriter, request *http.Request, e error) {
			logger.Error(fmt.Sprintf("failed to proxy request to %s", serverUrl.Host), sl.Err(e))
		},
	}

	bp.backends = append(bp.backends, &Backend{
		URL:          serverUrl,
		Alive:        true,
		Mux:          &sync.RWMutex{},
		ReverseProxy: proxy,
	})

	return nil
}
