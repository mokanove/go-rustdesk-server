package http_server

import (
	"go-rustdesk-server/cmd"
	"net/http"
)

func Always200Server() {
	handler := func(w http.ResponseWriter, r *http.Request) {
		cmd.Info("HTTP %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		w.Header().Set("Content-Length", "0")
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusOK)
	}
	ports := []string{":21114", ":21119"}
	for _, addr := range ports {
		addr := addr
		mux := http.NewServeMux()
		mux.HandleFunc("/", handler)
		go func() {
			cmd.Info("Fake HTTP server listening on %s", addr)
			if err := http.ListenAndServe(addr, mux); err != nil {
				cmd.Fatal("Fake HTTP server error on %s: %s", addr, err)
			}
		}()
	}
	select {}
}
