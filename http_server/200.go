package http_server

import (
	"go-rustdesk-server/cmd"
	"net/http"
)

func Always200Server() {
	addrs := []string{":21114", ":21119"}
	for _, addr := range addrs {
		go listenOn(addr)
	}
}

func listenOn(addr string) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		cmd.Info("HTTP %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		w.Header().Set("Content-Length", "0")
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusOK)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)
	cmd.Info("HTTP server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		cmd.Fatal("HTTP server on %s exited with error: %v", addr, err)
	}
}
