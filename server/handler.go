package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
)

type ContentRequest struct {
	Content string `json:"content"`
	BaseDir string `json:"base_dir"`
}
type ContentEvent struct {
	Type string `json:"type"`
	HTML string `json:"html"`
}

type ScrollRequest struct {
	CursorLine int `json:"cursor_line"`
}
type ScrollEvent struct {
	Type       string `json:"type"`
	CursorLine int    `json:"cursor_line"`
}

var (
	currentBaseDir string
	baseDirMu      sync.RWMutex
)

func handleContent(broker *Broker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Read all body from request
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "cannot read body", http.StatusBadRequest)
			return
		}

		// Parse JSON into ContentRequest
		var req ContentRequest
		if err = json.Unmarshal(body, &req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		baseDirMu.Lock()
		currentBaseDir = filepath.Clean(req.BaseDir)
		baseDirMu.Unlock()

		htmlData, err := markdownToHTML([]byte(req.Content), req.BaseDir)
		if err != nil {
			http.Error(w, "cannot render markdown", http.StatusInternalServerError)
			return
		}

		// Create event struct that send to browser
		event := ContentEvent{
			Type: "content",
			HTML: string(htmlData),
		}

		// Marshal this struct into JSON string
		eventJSON, err := json.Marshal(event)
		if err != nil {
			http.Error(w, "cannot marshal event", http.StatusInternalServerError)
			return
		}

		broker.SetLatestContent(string(eventJSON))

		// Send to all clients through broker
		broker.Broadcast(string(eventJSON))

		// Return 204
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleScroll(broker *Broker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "cannot read body", http.StatusBadRequest)
			return
		}

		var req ScrollRequest
		if err = json.Unmarshal(body, &req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		event := ScrollEvent{
			Type:       "scroll",
			CursorLine: req.CursorLine,
		}

		eventJSON, err := json.Marshal(event)
		if err != nil {
			http.Error(w, "cannot marshal event", http.StatusInternalServerError)
			return
		}

		broker.SetLatestScroll(string(eventJSON))

		broker.Broadcast(string(eventJSON))

		w.WriteHeader(http.StatusNoContent)
	}
}

func handleSSE(broker *Broker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// SSE required headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		// Use Flusher for push data --> browser instant
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}

		// Add new client to broker,
		// and done, remove it
		client := broker.Add()
		defer broker.Remove(client)

		if c := broker.GetLatestContent(); c != "" {
			fmt.Fprintln(w, "data:", c)
			fmt.Fprintln(w)
		}
		if s := broker.GetLatestScroll(); s != "" {
			fmt.Fprintln(w, "data:", s)
			fmt.Fprintln(w)
		}
		flusher.Flush()

		// Loop read event from client channel,
		// write to response
		for {
			select {
			case event := <-client.Events:
				fmt.Fprintln(w, "data:", event)
				fmt.Fprintln(w)
				flusher.Flush()
			case <-r.Context().Done():
				// Browser close tab or disconnect
				return
			}
		}
	}
}

func handleChromaCSS() http.HandlerFunc {
	css := chromaCSS()
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		w.Write([]byte(css))
	}
}

func handleLocalFiles() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filePath, ok := strings.CutPrefix(r.URL.Path, "/local/")
		if !ok {
			http.NotFound(w, r)
			return
		}
		filePath = filepath.FromSlash(filePath)

		if !filepath.IsAbs(filePath) {
			filePath = filepath.Join("/", filePath)
			filePath = filepath.Clean(filePath)
		}

		baseDirMu.RLock()
		baseDir := currentBaseDir
		baseDirMu.RUnlock()

		if baseDir == "" || !strings.HasPrefix(filePath, baseDir) {
			http.NotFound(w, r)
			return
		}

		http.ServeFile(w, r, filePath)
	}
}
