package load_balancer

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

const pingTimeout = 3 * time.Second

type Backend struct {
	URL          *url.URL
	Alive        bool
	Mux          *sync.RWMutex
	ReverseProxy *httputil.ReverseProxy
}

func (b *Backend) SetAlive(alive bool) {
	b.Mux.Lock()
	defer b.Mux.Unlock()

	b.Alive = alive
}

func (b *Backend) IsAlive() bool {
	b.Mux.RLock()
	defer b.Mux.RUnlock()

	return b.Alive
}

func PingBackend(u *url.URL) error {
	client := http.Client{
		Timeout: pingTimeout,
	}

	resp, err := client.Get(u.String())
	if err != nil {
		return fmt.Errorf("backend: %s is not alive: %w", u.String(), err)
	}
	defer resp.Body.Close()

	return nil
}
