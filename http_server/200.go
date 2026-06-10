package http_server

import (
	"net/http"
)

func Always200Server() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "0")
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusOK)
	})
	if err := http.ListenAndServe(":21114", nil); err != nil {
	}
}
