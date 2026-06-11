package main

import "net/http"

func setupRoutes(mux *http.ServeMux, broker *Broker) {
	mux.HandleFunc("GET /events", handleSSE(broker))
	mux.HandleFunc("POST /content", handleContent(broker))
	mux.HandleFunc("POST /scroll", handleScroll(broker))
	mux.HandleFunc("GET /chroma.css", handleChromaCSS())
}
