package request

import (
	"net"
	"net/http"
)

func GetClientIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
	}
	if ip == "" {
		ip = r.RemoteAddr
	}

	if host, _, err := net.SplitHostPort(ip); err == nil {
		return host
	}
	return ip
}
