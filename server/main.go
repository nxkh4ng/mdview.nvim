package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

//go:embed static/*
var staticFiles embed.FS

var version = "dev"

func main() {
	host := flag.String("host", "127.0.0.1", "listen host")
	port := flag.Int("port", 0, "listen port (0 = random)")
	browser := flag.String("browser", "", "browser command")
	maxBodySize := flag.Int("max-body-size", 10, "max request body size in MB")
	renderTimeout := flag.Int("render-timeout", 10, "max render time in seconds")
	versionFlag := flag.Bool("version", false, "print current version")

	flag.Parse()

	if *versionFlag {
		fmt.Printf("mdview %s\n", version)
		os.Exit(0)
	}

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

	srv := &http.Server{Handler: mux}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-quit
		fmt.Println("shutting down gracefully...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()

	if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
