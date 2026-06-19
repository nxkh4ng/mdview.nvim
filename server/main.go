package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
)

//go:embed static/*
var staticFiles embed.FS

func main() {
	host := flag.String("host", "127.0.0.1", "listen host")
	port := flag.Int("port", 0, "listen port (0 = random)")
	browser := flag.String("browser", "", "browser command")
	maxBodySize := flag.Int("max-body-size", 10, "max request body size in MB")
	renderTimeout := flag.Int("render-timeout", 10, "max render time in seconds")
	flag.Parse()

	addr := fmt.Sprintf("%s:%d", *host, *port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	if *port == 0 {
		tcpAddr, ok := listener.Addr().(*net.TCPAddr)
		if !ok {
			log.Fatal("unexpected address type")
		}
		*port = tcpAddr.Port
	}
	fmt.Printf("PORT:%d\n", *port)

	// Open browser
	url := fmt.Sprintf("http://%s:%d", *host, *port)
	openBrowser(url, *browser)

	// Create mux + broker
	// and setup routes
	mux := http.NewServeMux()
	broker := NewBroker()
	setupRoutes(mux, broker, *maxBodySize, *renderTimeout)

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
