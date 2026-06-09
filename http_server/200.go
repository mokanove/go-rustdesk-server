package http_server

import (
	logs "github.com/danbai225/go-logs"
	"go-rustdesk-server/common"
	"net/http"
)

func Always200Server() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "0")
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusOK)
		logs.Debug("HTTP", r.RemoteAddr, "-> 200")
	})
	logs.Info("HTTP listening", common.PortHTTP)
	if err := http.ListenAndServe("0.0.0.0"+common.PortHTTP, nil); err != nil {
		logs.Err("HTTP server failed", err)
	}
}
