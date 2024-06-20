package handlers

import (
	"fmt"
	"net/http"
)

// SSEManager manages the Server-Sent Events connections
type SSEManager struct {
	clients map[chan string]bool
}

// NewSSEManager creates a new SSEManager
func NewSSEManager() *SSEManager {
	return &SSEManager{
		clients: make(map[chan string]bool),
	}
}

// AddClient adds a new client to the SSEManager
func (m *SSEManager) AddClient(client chan string) {
	m.clients[client] = true
}

// RemoveClient removes a client from the SSEManager
func (m *SSEManager) RemoveClient(client chan string) {
	delete(m.clients, client)
	close(client)
}

// BroadcastMessage sends a message to all clients
func (m *SSEManager) BroadcastMessage(message string) {
	for client := range m.clients {
		client <- message
	}
}

// SSEHandler handles the SSE connection
func (m *SSEManager) SSEHandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	client := make(chan string)
	m.AddClient(client)

	defer m.RemoveClient(client)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	for {
		select {
		case msg := <-client:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

var SSE = NewSSEManager()
