package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
)

func main() {
	// 1. Listen TCP on port 0
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal(err)
	}

	// 2. Take port from listener
	port := listener.Addr().(*net.TCPAddr).Port
	fmt.Println(port)

	// 3. Create mux and add route `/ping` -> JSON
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok"}`)
	})

	// 4. Serve HTTP on listener
	log.Fatal(http.Serve(listener, mux))
}
