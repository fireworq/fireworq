package web

import (
	"net/http"
	"strconv"
)

func writeJSON(w http.ResponseWriter, json []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(json)))
	w.WriteHeader(200)
	w.Write(json)
}
