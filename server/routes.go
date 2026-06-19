package main

import "net/http"

func setupRoutes(mux *http.ServeMux, broker *Broker, maxBodySize, renderTimeout int) {
	mux.HandleFunc("GET /events", handleSSE(broker))
	mux.HandleFunc("POST /content", handleContent(broker, maxBodySize, renderTimeout))
	mux.HandleFunc("POST /scroll", handleScroll(broker, maxBodySize))
	mux.HandleFunc("GET /chroma.css", handleChromaCSS())
	mux.HandleFunc("GET /local/", handleLocalFiles())
}
