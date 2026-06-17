package main

import (
	"testing"
	"time"
)

func TestNewBroker_Empty(t *testing.T) {
	b := NewBroker()

	if len(b.clients) != 0 {
		t.Errorf("expected 0 clients, got %d", len(b.clients))
	}
	if b.latestContent != "" {
		t.Errorf("expected empty latestContent, got %q", b.latestContent)
	}
	if b.latestScroll != "" {
		t.Errorf("expected empty latestScroll, got %q", b.latestScroll)
	}
}

func TestAdd_SingleClient(t *testing.T) {
	b := NewBroker()
	c := b.Add()

	if c == nil {
		t.Fatal("expected non-nil client")
	}
	if c.Events == nil {
		t.Fatal("expected non-nil channel")
	}
	if len(b.clients) != 1 {
		t.Errorf("expected 1 client, got %d", len(b.clients))
	}
}

func TestAdd_MultipleClients(t *testing.T) {
	b := NewBroker()
	c1 := b.Add()
	c2 := b.Add()
	c3 := b.Add()

	if len(b.clients) != 3 {
		t.Errorf("expected 3 clients, got %d", len(b.clients))
	}
	// Mỗi client phải khác nhau
	if c1 == c2 || c2 == c3 || c1 == c3 {
		t.Error("each Add should return a new Client instance")
	}
}

func TestRemove_ExistingClient(t *testing.T) {
	b := NewBroker()
	c := b.Add()
	b.Remove(c)

	if len(b.clients) != 0 {
		t.Errorf("expected 0 clients after remove, got %d", len(b.clients))
	}
	_, ok := <-c.Events
	if ok {
		t.Error("expected closed channel after Remove")
	}
}

func TestRemove_NonExistentClient(t *testing.T) {
	b := NewBroker()
	c := &Client{Events: make(chan string, 1)}

	b.Remove(c)
}

func TestRemove_Twice(t *testing.T) {
	b := NewBroker()
	c := b.Add()

	b.Remove(c)

	b.Remove(c)
}

func TestBroadcast_SingleClient(t *testing.T) {
	b := NewBroker()
	c := b.Add()

	b.Broadcast("msg1")
	b.Broadcast("msg2")

	select {
	case got := <-c.Events:
		if got != "msg1" {
			t.Errorf("expected msg1, got %q", got)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for msg1")
	}

	select {
	case got := <-c.Events:
		if got != "msg2" {
			t.Errorf("expected msg2, got %q", got)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for msg2")
	}
}

func TestBroadcast_MultipleClients(t *testing.T) {
	b := NewBroker()
	c1 := b.Add()
	c2 := b.Add()
	c3 := b.Add()

	b.Broadcast("hello")

	for i, c := range []*Client{c1, c2, c3} {
		select {
		case got := <-c.Events:
			if got != "hello" {
				t.Errorf("client %d: expected hello, got %q", i, got)
			}
		case <-time.After(time.Second):
			t.Fatalf("client %d: timeout", i)
		}
	}
}

func TestBroadcast_SlowClientNotBlock(t *testing.T) {
	b := NewBroker()
	_ = b.Add()

	for range 64 {
		b.Broadcast("fill")
	}

	done := make(chan struct{})
	go func() {
		b.Broadcast("overflow")
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Broadcast blocked on slow client")
	}
}

func TestBroadcast_ClientRemoved(t *testing.T) {
	b := NewBroker()
	c := b.Add()
	b.Remove(c)

	b.Broadcast("ignored")
}

func TestBroadcast_ZeroClients(t *testing.T) {
	b := NewBroker()
	// Broker không có client nào → không panic
	b.Broadcast("nobody")
}

func TestSetGetLatestContent(t *testing.T) {
	b := NewBroker()
	b.SetLatestContent("content-A")

	got := b.GetLatestContent()
	if got != "content-A" {
		t.Errorf("expected content-A, got %q", got)
	}
}

func TestSetGetLatestScroll(t *testing.T) {
	b := NewBroker()
	b.SetLatestScroll("scroll-42")

	got := b.GetLatestScroll()
	if got != "scroll-42" {
		t.Errorf("expected scroll-42, got %q", got)
	}
}

func TestLatestContent_Overwrite(t *testing.T) {
	b := NewBroker()
	b.SetLatestContent("first")
	b.SetLatestContent("second")

	if got := b.GetLatestContent(); got != "second" {
		t.Errorf("expected second, got %q", got)
	}
}

func TestLatestContent_AfterRemove(t *testing.T) {
	b := NewBroker()
	b.SetLatestContent("stored")

	c := b.Add()
	b.Remove(c)

	if got := b.GetLatestContent(); got != "stored" {
		t.Errorf("latest should survive remove, got %q", got)
	}
}

func TestSSEConsistency(t *testing.T) {
	b := NewBroker()
	b.SetLatestContent(`{"type":"content","html":"<p>hi</p>"}`)
	b.SetLatestScroll(`{"type":"scroll","cursor_line":5}`)

	c := b.Add()

	if got := b.GetLatestContent(); got == "" {
		t.Error("expected non-empty latest content")
	}
	if got := b.GetLatestScroll(); got == "" {
		t.Error("expected non-empty latest scroll")
	}

	b.Broadcast("update")

	select {
	case got := <-c.Events:
		if got != "update" {
			t.Errorf("expected update, got %q", got)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}
