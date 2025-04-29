package request

import (
	"bytes"
	"io"
	"net/http"
)

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
