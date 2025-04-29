package load_balancer

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
			logger.Error(fmt.Sprintf("failed to handle request to %s", serverUrl.Host), sl.Err(e))

			alive := isBackendAlive(serverUrl)
			bp.changeBackendStatus(serverUrl, alive)
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

func (bp *BackendsPool) PingBackends(logger *slog.Logger) {
	for _, b := range bp.backends {
		alive := isBackendAlive(b.URL)
		b.SetAlive(alive)
		if !alive {
			logger.Warn(fmt.Sprintf("backend %s is not alive", b.URL))
		} else {
			logger.Info(fmt.Sprintf("backend %s is alive", b.URL))
		}
	}
}

func (bp *BackendsPool) changeBackendStatus(backendUrl *url.URL, status bool) {
	for _, b := range bp.backends {
		if b.URL.String() == backendUrl.String() {
			b.SetAlive(status)
			break
		}
	}
}
