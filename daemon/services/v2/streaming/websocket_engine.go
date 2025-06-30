package streaming

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/domalab/uma/daemon/logger"
	"github.com/domalab/uma/daemon/services/v2/collectors"
	"github.com/gorilla/websocket"
)

// WebSocketEngine provides real-time streaming
type WebSocketEngine struct {
	collector *collectors.SystemCollector
	clients   map[string]*StreamingClient
	upgrader  websocket.Upgrader
	mutex     sync.RWMutex

	// Performance configuration
	maxClients       int
	maxMessageSize   int64
	compressionLevel int
	deltaCompression bool
}

// StreamingClient represents a connected WebSocket client
type StreamingClient struct {
	id            string
	conn          *websocket.Conn
	send          chan []byte
	subscriptions map[string]*Subscription
	lastData      map[string]interface{}
	clientType    ClientType
	capabilities  ClientCapabilities
	mutex         sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
}

// Subscription defines what metrics a client wants to receive
type Subscription struct {
	Channel   string                 `json:"channel"`
	Interval  time.Duration          `json:"interval"`
	Fields    []string               `json:"fields,omitempty"`
	Filters   map[string]interface{} `json:"filters,omitempty"`
	LastSent  time.Time              `json:"-"`
	DeltaOnly bool                   `json:"delta_only"`
}

// ClientType identifies the type of client
type ClientType string

const (
	ClientHomeAssistant ClientType = "home-assistant"
	ClientiOS           ClientType = "ios"
	ClientAndroid       ClientType = "android"
	ClientWeb           ClientType = "web"
	ClientGeneric       ClientType = "generic"
)

// ClientCapabilities defines what the client supports
type ClientCapabilities struct {
	Compression      bool `json:"compression"`
	BinaryProtocol   bool `json:"binary"`
	DeltaCompression bool `json:"delta"`
	BatchUpdates     bool `json:"batch"`
}

// StreamMessage represents a message sent to clients
type StreamMessage struct {
	Timestamp int64       `json:"timestamp"`
	Channel   string      `json:"channel"`
	Data      interface{} `json:"data"`
	Delta     bool        `json:"delta,omitempty"`
	Sequence  int64       `json:"seq,omitempty"`
}

// ConnectionMessage handles client connection setup
type ConnectionMessage struct {
	Type         string             `json:"type"`
	Version      string             `json:"version"`
	Client       ClientType         `json:"client"`
	Capabilities ClientCapabilities `json:"capabilities"`
}

// SubscriptionMessage handles client subscriptions
type SubscriptionMessage struct {
	Type     string         `json:"type"`
	Channels []Subscription `json:"channels"`
}

// NewWebSocketEngine creates a streaming engine
func NewWebSocketEngine(collector *collectors.SystemCollector) *WebSocketEngine {
	return &WebSocketEngine{
		collector: collector,
		clients:   make(map[string]*StreamingClient),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for internal network
			},
			ReadBufferSize:    4096,
			WriteBufferSize:   4096,
			EnableCompression: true,
		},
		maxClients:       100,
		maxMessageSize:   1024 * 1024, // 1MB
		compressionLevel: 6,
		deltaCompression: true,
	}
}

// HandleWebSocket handles WebSocket upgrade and client management
func (wse *WebSocketEngine) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Check client limit
	wse.mutex.RLock()
	clientCount := len(wse.clients)
	wse.mutex.RUnlock()

	if clientCount >= wse.maxClients {
		http.Error(w, "Maximum clients reached", http.StatusServiceUnavailable)
		return
	}

	// Upgrade connection
	conn, err := wse.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Red("WebSocket upgrade failed: %v", err)
		return
	}

	// Create client context
	ctx, cancel := context.WithCancel(context.Background())

	// Create streaming client
	client := &StreamingClient{
		id:            fmt.Sprintf("client_%d", time.Now().UnixNano()),
		conn:          conn,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]*Subscription),
		lastData:      make(map[string]interface{}),
		clientType:    ClientGeneric,
		ctx:           ctx,
		cancel:        cancel,
	}

	// Register client
	wse.mutex.Lock()
	wse.clients[client.id] = client
	wse.mutex.Unlock()

	logger.Blue("WebSocket client connected: %s", client.id)

	// Start client handlers
	go wse.clientWriter(client)
	go wse.clientReader(client)

	// Start streaming for this client
	go wse.streamToClient(client)
}

// clientReader handles incoming messages from client
func (wse *WebSocketEngine) clientReader(client *StreamingClient) {
	defer wse.removeClient(client)

	client.conn.SetReadLimit(wse.maxMessageSize)
	client.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.conn.SetPongHandler(func(string) error {
		client.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		select {
		case <-client.ctx.Done():
			return
		default:
			_, message, err := client.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					logger.Yellow("WebSocket read error for client %s: %v", client.id, err)
				}
				return
			}

			if err := wse.processClientMessage(client, message); err != nil {
				logger.Yellow("Error processing message from client %s: %v", client.id, err)
			}
		}
	}
}

// clientWriter handles outgoing messages to client
func (wse *WebSocketEngine) clientWriter(client *StreamingClient) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		client.conn.Close()
	}()

	for {
		select {
		case <-client.ctx.Done():
			return
		case message, ok := <-client.send:
			client.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Use compression for large messages
			messageType := websocket.TextMessage
			if client.capabilities.Compression && len(message) > 1024 {
				// Compression would be applied here
			}

			if err := client.conn.WriteMessage(messageType, message); err != nil {
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

// processClientMessage processes incoming client messages
func (wse *WebSocketEngine) processClientMessage(client *StreamingClient, message []byte) error {
	var baseMessage map[string]interface{}
	if err := json.Unmarshal(message, &baseMessage); err != nil {
		return err
	}

	msgType, ok := baseMessage["type"].(string)
	if !ok {
		return fmt.Errorf("missing message type")
	}

	switch msgType {
	case "connect":
		return wse.handleConnect(client, message)
	case "subscribe":
		return wse.handleSubscribe(client, message)
	case "unsubscribe":
		return wse.handleUnsubscribe(client, message)
	case "ping":
		return wse.handlePing(client)
	default:
		return fmt.Errorf("unknown message type: %s", msgType)
	}
}

// handleConnect processes client connection setup
func (wse *WebSocketEngine) handleConnect(client *StreamingClient, message []byte) error {
	var connMsg ConnectionMessage
	if err := json.Unmarshal(message, &connMsg); err != nil {
		return err
	}

	client.mutex.Lock()
	client.clientType = connMsg.Client
	client.capabilities = connMsg.Capabilities
	client.mutex.Unlock()

	// Send connection acknowledgment
	response := map[string]interface{}{
		"type":    "connected",
		"version": "2.0",
		"server":  "uma-v2",
		"features": map[string]bool{
			"compression":   true,
			"delta":         wse.deltaCompression,
			"binary":        false, // Not implemented yet
			"batch_updates": true,
		},
	}

	return wse.sendToClient(client, response)
}

// handleSubscribe processes subscription requests
func (wse *WebSocketEngine) handleSubscribe(client *StreamingClient, message []byte) error {
	var subMsg SubscriptionMessage
	if err := json.Unmarshal(message, &subMsg); err != nil {
		return err
	}

	client.mutex.Lock()
	for _, sub := range subMsg.Channels {
		client.subscriptions[sub.Channel] = &Subscription{
			Channel:   sub.Channel,
			Interval:  sub.Interval,
			Fields:    sub.Fields,
			Filters:   sub.Filters,
			LastSent:  time.Time{}, // Force immediate send
			DeltaOnly: sub.DeltaOnly,
		}
	}
	client.mutex.Unlock()

	logger.Blue("Client %s subscribed to %d channels", client.id, len(subMsg.Channels))

	// Send subscription confirmation
	response := map[string]interface{}{
		"type":     "subscribed",
		"channels": subMsg.Channels,
	}

	return wse.sendToClient(client, response)
}

// handleUnsubscribe processes unsubscription requests
func (wse *WebSocketEngine) handleUnsubscribe(client *StreamingClient, message []byte) error {
	var unsubMsg struct {
		Type     string   `json:"type"`
		Channels []string `json:"channels"`
	}

	if err := json.Unmarshal(message, &unsubMsg); err != nil {
		return err
	}

	client.mutex.Lock()
	for _, channel := range unsubMsg.Channels {
		delete(client.subscriptions, channel)
		delete(client.lastData, channel)
	}
	client.mutex.Unlock()

	logger.Blue("Client %s unsubscribed from %d channels", client.id, len(unsubMsg.Channels))
	return nil
}

// handlePing responds to ping messages
func (wse *WebSocketEngine) handlePing(client *StreamingClient) error {
	response := map[string]interface{}{
		"type":      "pong",
		"timestamp": time.Now().Unix(),
	}
	return wse.sendToClient(client, response)
}

// streamToClient continuously streams metrics to a client
func (wse *WebSocketEngine) streamToClient(client *StreamingClient) {
	ticker := time.NewTicker(100 * time.Millisecond) // Check every 100ms
	defer ticker.Stop()

	// Start streaming immediately
	go wse.processClientSubscriptions(client)

	for {
		select {
		case <-client.ctx.Done():
			return
		case <-ticker.C:
			wse.processClientSubscriptions(client)
		}
	}
}

// processClientSubscriptions processes all subscriptions for a client
func (wse *WebSocketEngine) processClientSubscriptions(client *StreamingClient) {
	client.mutex.RLock()
	subscriptions := make(map[string]*Subscription)
	for k, v := range client.subscriptions {
		subscriptions[k] = v
	}
	client.mutex.RUnlock()

	for channel, subscription := range subscriptions {
		if time.Since(subscription.LastSent) >= subscription.Interval {
			if data, found := wse.collector.GetMetric(channel); found {
				// Apply delta compression if enabled
				var deltaData interface{}
				var isDelta bool

				if wse.deltaCompression && subscription.DeltaOnly {
					deltaData, isDelta = wse.calculateDelta(client, channel, data)
					if !isDelta && deltaData == nil {
						continue // No changes, skip this update
					}
				} else {
					deltaData = data
				}

				message := StreamMessage{
					Timestamp: time.Now().Unix(),
					Channel:   channel,
					Data:      deltaData,
					Delta:     isDelta,
				}

				if err := wse.sendToClient(client, message); err != nil {
					logger.Yellow("Failed to send to client %s: %v", client.id, err)
					continue
				}

				// Update last sent time
				client.mutex.Lock()
				if sub, exists := client.subscriptions[channel]; exists {
					sub.LastSent = time.Now()
				}
				client.mutex.Unlock()
			}
		}
	}
}

// calculateDelta calculates delta between current and last data
func (wse *WebSocketEngine) calculateDelta(client *StreamingClient, channel string, currentData interface{}) (interface{}, bool) {
	client.mutex.Lock()
	lastData, exists := client.lastData[channel]
	client.lastData[channel] = currentData
	client.mutex.Unlock()

	if !exists {
		return currentData, false // First time, send full data
	}

	// Simple delta calculation (would be more sophisticated in real implementation)
	// For now, just compare JSON representations
	currentJSON, _ := json.Marshal(currentData)
	lastJSON, _ := json.Marshal(lastData)

	if string(currentJSON) == string(lastJSON) {
		return nil, false // No changes
	}

	return currentData, true // Has changes
}

// sendToClient sends a message to a specific client
func (wse *WebSocketEngine) sendToClient(client *StreamingClient, data interface{}) error {
	message, err := json.Marshal(data)
	if err != nil {
		return err
	}

	select {
	case client.send <- message:
		return nil
	default:
		// Channel full, client too slow
		return fmt.Errorf("client send channel full")
	}
}

// removeClient removes a client from the engine
func (wse *WebSocketEngine) removeClient(client *StreamingClient) {
	wse.mutex.Lock()
	if _, ok := wse.clients[client.id]; ok {
		delete(wse.clients, client.id)
		close(client.send)
		client.cancel()
		logger.Blue("WebSocket client disconnected: %s", client.id)
	}
	wse.mutex.Unlock()
}

// GetStats returns streaming engine statistics
func (wse *WebSocketEngine) GetStats() map[string]interface{} {
	wse.mutex.RLock()
	defer wse.mutex.RUnlock()

	totalSubscriptions := 0
	clientTypes := make(map[ClientType]int)

	for _, client := range wse.clients {
		client.mutex.RLock()
		totalSubscriptions += len(client.subscriptions)
		clientTypes[client.clientType]++
		client.mutex.RUnlock()
	}

	return map[string]interface{}{
		"connected_clients":   len(wse.clients),
		"total_subscriptions": totalSubscriptions,
		"client_types":        clientTypes,
		"max_clients":         wse.maxClients,
		"delta_compression":   wse.deltaCompression,
	}
}
