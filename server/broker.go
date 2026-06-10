package main

import "sync"

type Client struct {
	Events chan string
}

func NewClient() *Client {
	return &Client{
		Events: make(chan string, 64),
	}
}

type Broker struct {
	mu      sync.Mutex
	clients map[*Client]struct{}

	latestContent string
	latestScroll  string
}

func NewBroker() *Broker {
	return &Broker{
		clients: make(map[*Client]struct{}),
	}
}

func (b *Broker) Add() *Client {
	b.mu.Lock()
	defer b.mu.Unlock()

	client := NewClient()
	b.clients[client] = struct{}{}
	return client
}

func (b *Broker) Remove(client *Client) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, exists := b.clients[client]; !exists {
		return
	}
	delete(b.clients, client)
	close(client.Events)
}

func (b *Broker) Broadcast(event string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for client := range b.clients {
		select {
		case client.Events <- event:
			// send success
		default:
			// client slow, skip (not block)
		}
	}
}

func (b *Broker) SetLatestContent(event string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.latestContent = event
}

func (b *Broker) GetLatestContent() string {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.latestContent
}

func (b *Broker) SetLatestScroll(event string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.latestScroll = event
}

func (b *Broker) GetLatestScroll() string {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.latestScroll
}
