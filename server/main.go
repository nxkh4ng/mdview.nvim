package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
)

//go:embed static/*
var staticFiles embed.FS

func main() {
	// Listen TCP on port 0
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal(err)
	}

	// Take port from listener
	port := listener.Addr().(*net.TCPAddr).Port
	fmt.Println(port)

	// Open browser
	url := fmt.Sprintf("http://localhost:%d", port)
	openBrowser(url)

	// Create mux + broker
	// and setup routes
	mux := http.NewServeMux()
	broker := NewBroker()
	setupRoutes(mux, broker)

	// Route for /ping
	// and / (static)
	mux.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok"}`)
	})

	subDir, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatalf("cannot find static folder: %v", err)
	}
	mux.Handle("GET /", http.FileServer(http.FS(subDir)))

	// Serve HTTP on listener
	log.Fatal(http.Serve(listener, mux))
}
