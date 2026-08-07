package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func setupTestBaseDir(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()

	filePath := filepath.Join(dir, "test.md")
	if err := os.WriteFile(filePath, []byte("# hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	return filePath
}

func request(t *testing.T, url string) *httptest.ResponseRecorder {
	t.Helper()

	handler := handleLocalFiles()
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, url, nil)
	handler(w, r)
	return w
}

func TestHandleLocalFiles_ValidFile(t *testing.T) {
	filePath := setupTestBaseDir(t)

	w := request(t, "/local/"+filePath)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "# hello" {
		t.Errorf("expected raw content, got:\n%s", w.Body.String())
	}
}

func TestHandleLocalFiles_NoBaseDir(t *testing.T) {
	w := request(t, "/local/some/file.md")

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for non-existent file, got %d", w.Code)
	}
}

func TestHandleLocalFiles_MissingLocalPrefix(t *testing.T) {
	setupTestBaseDir(t)

	w := request(t, "/notlocal/file.md")

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for wrong prefix, got %d", w.Code)
	}
}

func TestHandleLocalFiles_NonExistentFile(t *testing.T) {
	setupTestBaseDir(t)

	w := request(t, "/local/nonexistent.md")

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for non-existent file, got %d", w.Code)
	}
}

// ── handleContent tests ──

func TestHandleContent_ValidContent(t *testing.T) {
	broker := NewBroker()
	handler := handleContent(broker, 10, 10)

	body := `{"content":"# hello","base_dir":""}`
	req := httptest.NewRequest(http.MethodPost, "/content", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
}

func TestHandleContent_InvalidJSON(t *testing.T) {
	broker := NewBroker()
	handler := handleContent(broker, 10, 10)

	req := httptest.NewRequest(http.MethodPost, "/content", strings.NewReader(`not json`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", w.Code)
	}
}

func TestHandleContent_OversizedBody(t *testing.T) {
	broker := NewBroker()
	handler := handleContent(broker, 1, 10) // 1MB limit

	// Send > 1MB
	req := httptest.NewRequest(http.MethodPost, "/content", strings.NewReader(strings.Repeat("a", 2*1024*1024)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for oversized body, got %d", w.Code)
	}
}

func TestHandleContent_WithBaseDir(t *testing.T) {
	broker := NewBroker()
	handler := handleContent(broker, 10, 10)

	body := `{"content":"![](image.png)","base_dir":"/home/user/docs"}`
	req := httptest.NewRequest(http.MethodPost, "/content", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
}

func TestHandleContent_BroadcastsEvent(t *testing.T) {
	broker := NewBroker()
	handler := handleContent(broker, 10, 10)

	// Register an SSE client to verify broadcast
	client := broker.Add()

	body := `{"content":"# hello","base_dir":""}`
	req := httptest.NewRequest(http.MethodPost, "/content", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}

	select {
	case event := <-client.Events:
		var ev ContentEvent
		if err := json.Unmarshal([]byte(event), &ev); err != nil {
			t.Fatalf("invalid event JSON: %v", err)
		}
		if ev.Type != "content" {
			t.Errorf("expected event type 'content', got %q", ev.Type)
		}
		if !strings.Contains(ev.HTML, "hello") {
			t.Errorf("expected 'hello' in rendered HTML, got:\n%s", ev.HTML)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for broadcast event")
	}

	// Also verify latest content was stored
	if broker.GetLatestContent() == "" {
		t.Error("expected latest content to be set")
	}
}

// ── handleScroll tests ──

func TestHandleScroll_Valid(t *testing.T) {
	t.Parallel()

	broker := NewBroker()
	handler := handleScroll(broker, 10)

	body := `{"cursor_line":42}`
	req := httptest.NewRequest(http.MethodPost, "/scroll", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}

	if broker.GetLatestScroll() == "" {
		t.Error("expected latest scroll to be set")
	}
}

func TestHandleScroll_InvalidJSON(t *testing.T) {
	t.Parallel()

	broker := NewBroker()
	handler := handleScroll(broker, 10)

	req := httptest.NewRequest(http.MethodPost, "/scroll", strings.NewReader(`not json`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", w.Code)
	}
}

func TestHandleScroll_OversizedBody(t *testing.T) {
	t.Parallel()

	broker := NewBroker()
	handler := handleScroll(broker, 1) // 1MB limit

	req := httptest.NewRequest(http.MethodPost, "/scroll", strings.NewReader(strings.Repeat("a", 2*1024*1024)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for oversized body, got %d", w.Code)
	}
}

func TestHandleScroll_BroadcastsEvent(t *testing.T) {
	t.Parallel()

	broker := NewBroker()
	handler := handleScroll(broker, 10)

	client := broker.Add()

	body := `{"cursor_line":7}`
	req := httptest.NewRequest(http.MethodPost, "/scroll", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}

	select {
	case event := <-client.Events:
		var ev ScrollEvent
		if err := json.Unmarshal([]byte(event), &ev); err != nil {
			t.Fatalf("invalid event JSON: %v", err)
		}
		if ev.Type != "scroll" {
			t.Errorf("expected event type 'scroll', got %q", ev.Type)
		}
		if ev.CursorLine != 7 {
			t.Errorf("expected cursor_line=7, got %d", ev.CursorLine)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for broadcast event")
	}
}

// ── handleSSE tests ──

type flushRecorder struct {
	*httptest.ResponseRecorder
	flushed chan struct{}
}

func newFlushRecorder() *flushRecorder {
	return &flushRecorder{
		ResponseRecorder: httptest.NewRecorder(),
		flushed:          make(chan struct{}, 10),
	}
}

func (f *flushRecorder) Flush() {
	select {
	case f.flushed <- struct{}{}:
	default:
	}
}

func TestHandleSSE_Headers(t *testing.T) {
	t.Parallel()

	broker := NewBroker()
	handler := handleSSE(broker)

	ctx, cancel := context.WithCancel(t.Context())
	req := httptest.NewRequest(http.MethodGet, "/events", nil).WithContext(ctx)
	w := newFlushRecorder()

	go handler(w, req)
	time.Sleep(50 * time.Millisecond)
	cancel()

	if w.Header().Get("Content-Type") != "text/event-stream" {
		t.Errorf("expected Content-Type text/event-stream, got %q", w.Header().Get("Content-Type"))
	}
	if w.Header().Get("Cache-Control") != "no-cache" {
		t.Errorf("expected Cache-Control no-cache, got %q", w.Header().Get("Cache-Control"))
	}
	if w.Header().Get("Connection") != "keep-alive" {
		t.Errorf("expected Connection keep-alive, got %q", w.Header().Get("Connection"))
	}
}

func TestHandleSSE_SendsLatestOnConnect(t *testing.T) {
	t.Parallel()

	broker := NewBroker()
	broker.SetLatestContent(`{"type":"content","html":"<p>hello</p>"}`)
	broker.SetLatestScroll(`{"type":"scroll","cursor_line":42}`)

	handler := handleSSE(broker)

	ctx, cancel := context.WithCancel(t.Context())
	req := httptest.NewRequest(http.MethodGet, "/events", nil).WithContext(ctx)
	w := newFlushRecorder()

	go handler(w, req)

	// Wait for initial flush
	select {
	case <-w.flushed:
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for initial flush")
	}
	cancel()

	body := w.Body.String()
	if !strings.Contains(body, "hello") {
		t.Errorf("expected latest content in SSE response, got:\n%s", body)
	}
	if !strings.Contains(body, "42") {
		t.Errorf("expected scroll line 42 in SSE response, got:\n%s", body)
	}
}

func TestHandleSSE_ExitsOnCancel(t *testing.T) {
	t.Parallel()

	broker := NewBroker()
	handler := handleSSE(broker)

	ctx, cancel := context.WithCancel(t.Context())
	req := httptest.NewRequest(http.MethodGet, "/events", nil).WithContext(ctx)
	w := newFlushRecorder()

	done := make(chan struct{})
	go func() {
		handler(w, req)
		close(done)
	}()

	// Wait for initial flush, then cancel
	select {
	case <-w.flushed:
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for initial flush")
	}
	cancel()

	select {
	case <-done:
		// handler exited — good
	case <-time.After(time.Second):
		t.Fatal("handler did not exit within 1s after context cancel")
	}
}

func TestHandleSSE_ForwardsEvents(t *testing.T) {
	t.Parallel()

	broker := NewBroker()
	handler := handleSSE(broker)

	ctx, cancel := context.WithCancel(t.Context())
	req := httptest.NewRequest(http.MethodGet, "/events", nil).WithContext(ctx)
	w := newFlushRecorder()

	go handler(w, req)

	// Wait for initial flush
	select {
	case <-w.flushed:
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for initial flush")
	}

	// Broadcast an event after SSE client is connected
	broker.Broadcast(`{"type":"scroll","cursor_line":99}`)

	// Give handler time to process and flush
	time.Sleep(200 * time.Millisecond)
	cancel()

	body := w.Body.String()
	if !strings.Contains(body, "99") {
		t.Errorf("expected broadcasted cursor_line 99 in SSE response, got:\n%s", body)
	}
}
