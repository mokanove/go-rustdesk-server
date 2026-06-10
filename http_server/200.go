package http_server

import (
	"go-rustdesk-server/common"
	"net/http"
)

func Always200Server() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "0")
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusOK)
	})
	if err := http.ListenAndServe("0.0.0.0"+common.PortHTTP, nil); err != nil {
	}
}
