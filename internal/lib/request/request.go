package request

import (
	"bytes"
	"io"
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

func CloneRequest(r *http.Request) *http.Request {
	newReq := r.Clone(r.Context())

	if r.Body != nil {
		bodyBytes, _ := io.ReadAll(r.Body)
		r.Body.Close()

		r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

		newReq.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	return newReq
}
