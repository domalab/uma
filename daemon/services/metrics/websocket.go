package metrics

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/domalab/uma/daemon/logger"
)

// WebSocketStreamer manages WebSocket connections for real-time metrics streaming
type WebSocketStreamer struct {
	collector *MetricsCollector
	clients   map[*Client]bool
	upgrader  websocket.Upgrader
	mutex     sync.RWMutex
}

// Client represents a WebSocket client connection
type Client struct {
	conn         *websocket.Conn
	send         chan []byte
	subscriptions map[string]*Subscription
	id           string
	mutex        sync.RWMutex
}

// Subscription represents a client's subscription to specific metrics
type Subscription struct {
	Metric   string        `json:"metric"`
	Interval time.Duration `json:"interval"`
	Filters  map[string]interface{} `json:"filters,omitempty"`
	LastSent time.Time     `json:"-"`
}

// StreamMessage represents a message sent to WebSocket clients
type StreamMessage struct {
	Timestamp string      `json:"timestamp"`
	Metric    string      `json:"metric"`
	Data      interface{} `json:"data"`
	ClientID  string      `json:"client_id,omitempty"`
}

// SubscriptionRequest represents a client subscription request
type SubscriptionRequest struct {
	Action   string                 `json:"action"` // "subscribe", "unsubscribe", "list"
	Metrics  []string               `json:"metrics,omitempty"`
	Interval int                    `json:"interval,omitempty"` // seconds
	Filters  map[string]interface{} `json:"filters,omitempty"`
}

// NewWebSocketStreamer creates a new WebSocket streamer
func NewWebSocketStreamer(collector *MetricsCollector) *WebSocketStreamer {
	return &WebSocketStreamer{
		collector: collector,
		clients:   make(map[*Client]bool),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Allow all origins for now - should be configurable in production
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

// HandleWebSocket handles WebSocket upgrade and client management
func (ws *WebSocketStreamer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Red("WebSocket upgrade failed: %v", err)
		return
	}

	client := &Client{
		conn:          conn,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]*Subscription),
		id:            fmt.Sprintf("client_%d", time.Now().UnixNano()),
	}

	ws.mutex.Lock()
	ws.clients[client] = true
	ws.mutex.Unlock()

	logger.Blue("WebSocket client connected: %s", client.id)

	// Start client goroutines
	go ws.clientWriter(client)
	go ws.clientReader(client)
}

// clientReader handles incoming messages from a WebSocket client
func (ws *WebSocketStreamer) clientReader(client *Client) {
	defer func() {
		ws.removeClient(client)
		client.conn.Close()
	}()

	client.conn.SetReadLimit(512)
	client.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.conn.SetPongHandler(func(string) error {
		client.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Yellow("WebSocket error for client %s: %v", client.id, err)
			}
			break
		}

		ws.handleClientMessage(client, message)
	}
}

// clientWriter handles outgoing messages to a WebSocket client
func (ws *WebSocketStreamer) clientWriter(client *Client) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		client.conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.send:
			client.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := client.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				logger.Yellow("WebSocket write error for client %s: %v", client.id, err)
				return
			}

		case <-ticker.C:
			client.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleClientMessage processes incoming client messages
func (ws *WebSocketStreamer) handleClientMessage(client *Client, message []byte) {
	var req SubscriptionRequest
	if err := json.Unmarshal(message, &req); err != nil {
		logger.Yellow("Invalid WebSocket message from client %s: %v", client.id, err)
		return
	}

	switch req.Action {
	case "subscribe":
		ws.handleSubscribe(client, &req)
	case "unsubscribe":
		ws.handleUnsubscribe(client, &req)
	case "list":
		ws.handleListSubscriptions(client)
	default:
		logger.Yellow("Unknown action from client %s: %s", client.id, req.Action)
	}
}

// handleSubscribe processes subscription requests
func (ws *WebSocketStreamer) handleSubscribe(client *Client, req *SubscriptionRequest) {
	interval := time.Duration(req.Interval) * time.Second
	if interval < 1*time.Second {
		interval = 5 * time.Second // Default interval
	}

	client.mutex.Lock()
	for _, metric := range req.Metrics {
		client.subscriptions[metric] = &Subscription{
			Metric:   metric,
			Interval: interval,
			Filters:  req.Filters,
			LastSent: time.Time{}, // Force immediate send
		}
	}
	client.mutex.Unlock()

	logger.Blue("Client %s subscribed to %d metrics", client.id, len(req.Metrics))

	// Send confirmation
	response := map[string]interface{}{
		"action": "subscribed",
		"metrics": req.Metrics,
		"interval": req.Interval,
	}
	ws.sendToClient(client, response)
}

// handleUnsubscribe processes unsubscription requests
func (ws *WebSocketStreamer) handleUnsubscribe(client *Client, req *SubscriptionRequest) {
	client.mutex.Lock()
	for _, metric := range req.Metrics {
		delete(client.subscriptions, metric)
	}
	client.mutex.Unlock()

	logger.Blue("Client %s unsubscribed from %d metrics", client.id, len(req.Metrics))
}

// handleListSubscriptions sends current subscriptions to client
func (ws *WebSocketStreamer) handleListSubscriptions(client *Client) {
	client.mutex.RLock()
	subscriptions := make([]string, 0, len(client.subscriptions))
	for metric := range client.subscriptions {
		subscriptions = append(subscriptions, metric)
	}
	client.mutex.RUnlock()

	response := map[string]interface{}{
		"action": "subscriptions",
		"metrics": subscriptions,
	}
	ws.sendToClient(client, response)
}

// sendToClient sends a message to a specific client
func (ws *WebSocketStreamer) sendToClient(client *Client, data interface{}) {
	message, err := json.Marshal(data)
	if err != nil {
		logger.Red("Failed to marshal message for client %s: %v", client.id, err)
		return
	}

	select {
	case client.send <- message:
	default:
		close(client.send)
		ws.removeClient(client)
	}
}

// removeClient removes a client from the streamer
func (ws *WebSocketStreamer) removeClient(client *Client) {
	ws.mutex.Lock()
	if _, ok := ws.clients[client]; ok {
		delete(ws.clients, client)
		close(client.send)
		logger.Blue("WebSocket client disconnected: %s", client.id)
	}
	ws.mutex.Unlock()
}

// StartStreaming begins streaming metrics to connected clients
func (ws *WebSocketStreamer) StartStreaming() {
	go ws.streamingLoop()
	logger.Green("Started WebSocket metrics streaming")
}

// streamingLoop continuously streams metrics to subscribed clients
func (ws *WebSocketStreamer) streamingLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		ws.mutex.RLock()
		clients := make([]*Client, 0, len(ws.clients))
		for client := range ws.clients {
			clients = append(clients, client)
		}
		ws.mutex.RUnlock()

		for _, client := range clients {
			ws.processClientSubscriptions(client)
		}
	}
}

// processClientSubscriptions processes all subscriptions for a client
func (ws *WebSocketStreamer) processClientSubscriptions(client *Client) {
	client.mutex.RLock()
	subscriptions := make(map[string]*Subscription)
	for k, v := range client.subscriptions {
		subscriptions[k] = v
	}
	client.mutex.RUnlock()

	for metric, subscription := range subscriptions {
		if time.Since(subscription.LastSent) >= subscription.Interval {
			if data, found := ws.collector.GetMetric(metric); found {
				message := StreamMessage{
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					Metric:    metric,
					Data:      data,
				}

				ws.sendToClient(client, message)

				// Update last sent time
				client.mutex.Lock()
				if sub, exists := client.subscriptions[metric]; exists {
					sub.LastSent = time.Now()
				}
				client.mutex.Unlock()
			}
		}
	}
}

// GetStats returns WebSocket streamer statistics
func (ws *WebSocketStreamer) GetStats() map[string]interface{} {
	ws.mutex.RLock()
	defer ws.mutex.RUnlock()

	totalSubscriptions := 0
	for client := range ws.clients {
		client.mutex.RLock()
		totalSubscriptions += len(client.subscriptions)
		client.mutex.RUnlock()
	}

	return map[string]interface{}{
		"connected_clients":    len(ws.clients),
		"total_subscriptions": totalSubscriptions,
	}
}
