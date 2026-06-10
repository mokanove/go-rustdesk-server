package http_server

import (
	"net/http"
	"go-rustdesk-server/cmd"
)

func Always200Server() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		cmd.Info("HTTP %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		w.Header().Set("Content-Length", "0")
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusOK)
	})
	if err := http.ListenAndServe(":21114", nil); err != nil {
		cmd.Fatal("HTTP server error: %s", err)
	}
}
