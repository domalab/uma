package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/cskr/pubsub"
	"github.com/domalab/uma/daemon/logger"
	"github.com/domalab/uma/daemon/services/api/utils"
	"github.com/gorilla/websocket"
)

// EventType represents different types of events
type EventType string

const (
	// System Events
	EventSystemStats  EventType = "system.stats"
	EventSystemHealth EventType = "system.health"
	EventSystemLoad   EventType = "system.load"
	EventCPUStats     EventType = "cpu.stats"
	EventMemoryStats  EventType = "memory.stats"
	EventNetworkStats EventType = "network.stats"

	// Storage Events
	EventStorageStatus    EventType = "storage.status"
	EventDiskStats        EventType = "disk.stats"
	EventArrayStatus      EventType = "array.status"
	EventParityStatus     EventType = "parity.status"
	EventDiskSMARTWarning EventType = "disk.smart.warning"
	EventCacheStatus      EventType = "cache.status"

	// Container Events
	EventDockerEvents    EventType = "docker.events"
	EventContainerStats  EventType = "container.stats"
	EventContainerHealth EventType = "container.health"
	EventImageEvents     EventType = "image.events"

	// VM Events
	EventVMEvents EventType = "vm.events"
	EventVMStats  EventType = "vm.stats"
	EventVMHealth EventType = "vm.health"

	// Alert Events
	EventTemperatureAlert EventType = "temperature.alert"
	EventResourceAlert    EventType = "resource.alert"
	EventSecurityAlert    EventType = "security.alert"
	EventSystemAlert      EventType = "system.alert"

	// Infrastructure Events
	EventUPSStatus   EventType = "ups.status"
	EventFanStatus   EventType = "fan.status"
	EventPowerStatus EventType = "power.status"

	// Operational Events
	EventTaskProgress EventType = "task.progress"
	EventBackupStatus EventType = "backup.status"
	EventUpdateStatus EventType = "update.status"
)

// WebSocketConnection represents an active WebSocket connection
type WebSocketConnection struct {
	conn          *websocket.Conn
	subscriptions map[EventType]bool
	mutex         sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
}

// EnhancedWebSocketHandler handles WebSocket connections with pubsub integration
type EnhancedWebSocketHandler struct {
	api         utils.APIInterface
	upgrader    websocket.Upgrader
	hub         *pubsub.PubSub
	connections map[string]*WebSocketConnection
	mutex       sync.RWMutex
}

// NewEnhancedWebSocketHandler creates a new enhanced WebSocket handler
func NewEnhancedWebSocketHandler(api utils.APIInterface, hub *pubsub.PubSub) *EnhancedWebSocketHandler {
	return &EnhancedWebSocketHandler{
		api: api,
		hub: hub,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Allow all origins for now - should be configurable in production
				return true
			},
		},
		connections: make(map[string]*WebSocketConnection),
	}
}

// HandleUnifiedWebSocket handles unified WebSocket connections with subscription management
func (h *EnhancedWebSocketHandler) HandleUnifiedWebSocket(w http.ResponseWriter, r *http.Request) {
	logger.Blue("Enhanced WebSocket connection attempt")

	// Check if API adapter is available
	if h.api == nil {
		logger.Red("Enhanced WebSocket handler: API adapter is nil")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Red("WebSocket upgrade failed: %v", err)
		return
	}

	// Create connection context
	ctx, cancel := context.WithCancel(context.Background())

	// Create WebSocket connection object
	wsConn := &WebSocketConnection{
		conn:          conn,
		subscriptions: make(map[EventType]bool),
		ctx:           ctx,
		cancel:        cancel,
	}

	// Generate connection ID
	connID := fmt.Sprintf("ws_%d", time.Now().UnixNano())

	// Store connection
	h.mutex.Lock()
	h.connections[connID] = wsConn
	h.mutex.Unlock()

	logger.Green("Enhanced WebSocket connection established: %s", connID)

	// Handle connection cleanup
	defer func() {
		h.mutex.Lock()
		delete(h.connections, connID)
		h.mutex.Unlock()
		cancel()
		conn.Close()
		logger.Blue("Enhanced WebSocket connection closed: %s", connID)
	}()

	// Start message handler
	go h.handleMessages(wsConn)

	// Start event broadcaster
	h.startEventBroadcaster(wsConn)
}

// handleMessages handles incoming WebSocket messages
func (h *EnhancedWebSocketHandler) handleMessages(wsConn *WebSocketConnection) {
	for {
		select {
		case <-wsConn.ctx.Done():
			return
		default:
			messageType, data, err := wsConn.conn.ReadMessage()
			if err != nil {
				logger.Yellow("WebSocket read error: %v", err)
				return
			}

			if messageType == websocket.TextMessage {
				if err := h.processMessage(wsConn, data); err != nil {
					logger.Yellow("Error processing WebSocket message: %v", err)
				}
			}
		}
	}
}

// processMessage processes incoming WebSocket messages
func (h *EnhancedWebSocketHandler) processMessage(wsConn *WebSocketConnection, data []byte) error {
	var message map[string]interface{}
	if err := json.Unmarshal(data, &message); err != nil {
		return err
	}

	msgType, ok := message["type"].(string)
	if !ok {
		return nil // Ignore messages without type
	}

	switch msgType {
	case "ping":
		return h.handlePing(wsConn)
	case "subscribe":
		return h.handleSubscribe(wsConn, message)
	case "unsubscribe":
		return h.handleUnsubscribe(wsConn, message)
	default:
		logger.Yellow("Unknown WebSocket message type: %s", msgType)
	}

	return nil
}

// handlePing responds to ping messages
func (h *EnhancedWebSocketHandler) handlePing(wsConn *WebSocketConnection) error {
	response := map[string]interface{}{
		"type":      "pong",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	return wsConn.conn.WriteJSON(response)
}

// handleSubscribe handles subscription requests
func (h *EnhancedWebSocketHandler) handleSubscribe(wsConn *WebSocketConnection, message map[string]interface{}) error {
	channels, ok := message["channels"].([]interface{})
	if !ok {
		return fmt.Errorf("invalid channels format")
	}

	wsConn.mutex.Lock()
	defer wsConn.mutex.Unlock()

	subscribedChannels := make([]string, 0)
	for _, ch := range channels {
		if channelStr, ok := ch.(string); ok {
			eventType := EventType(channelStr)
			wsConn.subscriptions[eventType] = true
			subscribedChannels = append(subscribedChannels, channelStr)
		}
	}

	response := map[string]interface{}{
		"type":      "subscription_ack",
		"channels":  subscribedChannels,
		"status":    "subscribed",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	return wsConn.conn.WriteJSON(response)
}

// handleUnsubscribe handles unsubscription requests
func (h *EnhancedWebSocketHandler) handleUnsubscribe(wsConn *WebSocketConnection, message map[string]interface{}) error {
	channels, ok := message["channels"].([]interface{})
	if !ok {
		return fmt.Errorf("invalid channels format")
	}

	wsConn.mutex.Lock()
	defer wsConn.mutex.Unlock()

	unsubscribedChannels := make([]string, 0)
	for _, ch := range channels {
		if channelStr, ok := ch.(string); ok {
			eventType := EventType(channelStr)
			delete(wsConn.subscriptions, eventType)
			unsubscribedChannels = append(unsubscribedChannels, channelStr)
		}
	}

	response := map[string]interface{}{
		"type":      "unsubscription_ack",
		"channels":  unsubscribedChannels,
		"status":    "unsubscribed",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	return wsConn.conn.WriteJSON(response)
}

// startEventBroadcaster starts the event broadcasting for a connection
func (h *EnhancedWebSocketHandler) startEventBroadcaster(wsConn *WebSocketConnection) {
	// Subscribe to all event types from pubsub
	eventTypes := []string{
		// System Events
		string(EventSystemStats),
		string(EventSystemHealth),
		string(EventSystemLoad),
		string(EventCPUStats),
		string(EventMemoryStats),
		string(EventNetworkStats),

		// Storage Events
		string(EventStorageStatus),
		string(EventDiskStats),
		string(EventArrayStatus),
		string(EventParityStatus),
		string(EventDiskSMARTWarning),
		string(EventCacheStatus),

		// Container Events
		string(EventDockerEvents),
		string(EventContainerStats),
		string(EventContainerHealth),
		string(EventImageEvents),

		// VM Events
		string(EventVMEvents),
		string(EventVMStats),
		string(EventVMHealth),

		// Alert Events
		string(EventTemperatureAlert),
		string(EventResourceAlert),
		string(EventSecurityAlert),
		string(EventSystemAlert),

		// Infrastructure Events
		string(EventUPSStatus),
		string(EventFanStatus),
		string(EventPowerStatus),

		// Operational Events
		string(EventTaskProgress),
		string(EventBackupStatus),
		string(EventUpdateStatus),
	}

	eventChan := h.hub.Sub(eventTypes...)
	defer h.hub.Unsub(eventChan, eventTypes...)

	for {
		select {
		case <-wsConn.ctx.Done():
			return
		case event := <-eventChan:
			if err := h.broadcastEvent(wsConn, event); err != nil {
				logger.Yellow("Error broadcasting event: %v", err)
				return
			}
		}
	}
}

// broadcastEvent broadcasts an event to a WebSocket connection if subscribed
func (h *EnhancedWebSocketHandler) broadcastEvent(wsConn *WebSocketConnection, event interface{}) error {
	eventData, ok := event.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid event format")
	}

	eventType, ok := eventData["event_type"].(string)
	if !ok {
		return fmt.Errorf("missing event_type in event")
	}

	wsConn.mutex.RLock()
	subscribed := wsConn.subscriptions[EventType(eventType)]
	wsConn.mutex.RUnlock()

	if !subscribed {
		return nil // Not subscribed to this event type
	}

	return wsConn.conn.WriteJSON(eventData)
}

// BroadcastEvent broadcasts an event to all connected clients
func (h *EnhancedWebSocketHandler) BroadcastEvent(eventType EventType, data interface{}) {
	event := map[string]interface{}{
		"event_type": string(eventType),
		"data":       data,
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	}

	// Publish to pubsub system
	h.hub.Pub(event, string(eventType))
}
