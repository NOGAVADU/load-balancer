package load_balancer

import (
	"net"
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

func isBackendAlive(u *url.URL) bool {
	conn, err := net.DialTimeout("tcp", u.Host, pingTimeout)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}
