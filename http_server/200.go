package http_server

import (
	"fmt"
	"go-rustdesk-server/common"
	"net/http"
)

func Always200Server() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "0")
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusOK)
		fmt.Printf("[HTTP] %s → 200 OK\n", r.RemoteAddr)
	})
	fmt.Printf("[HTTP] Listening %s\n", common.PortHTTP)
	if err := http.ListenAndServe("0.0.0.0"+common.PortHTTP, nil); err != nil {
		fmt.Printf("[HTTP] Server failed: %v\n", err)
	}
}
