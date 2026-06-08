package http_server

import (
	"fmt"
	"net/http"
)

func Always200Server() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "0")
		w.Header().Del("Content-Type")
		w.WriteHeader(http.StatusOK)
		
		fmt.Printf("[HTTP] Client connected from %s, returned a fake 200 OK\n", r.RemoteAddr)
	})

	fmt.Println("[HTTP] Started an Always 200 OK Server, listening on port 21114 from any addr")

	if err := http.ListenAndServe(":21114", nil); err != nil {
		fmt.Printf("[HTTP] Server failed to start: %v\n", err)
	}
}