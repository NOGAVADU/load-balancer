package response

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type Response struct {
	Status int    `json:"status"`
	Error  string `json:"error,omitempty"`
}

func Err(w http.ResponseWriter, msg string, status int) {
	response := Response{
		Status: status,
		Error:  msg,
	}

	JSON(w, response, status)
}

func JSON(w http.ResponseWriter, v any, status int) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(true)
	if err := enc.Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(buf.Bytes())
}
