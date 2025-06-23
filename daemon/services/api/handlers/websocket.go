package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
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
	clientIP      string
	messageCount  int
	lastMessage   time.Time
}

// WebSocketHandler handles WebSocket connections with pubsub integration
type WebSocketHandler struct {
	api         utils.APIInterface
	upgrader    websocket.Upgrader
	hub         *pubsub.PubSub
	connections map[string]*WebSocketConnection
	mutex       sync.RWMutex

	// Security configuration
	maxConnections    int
	maxMessageSize    int64
	messagesPerMinute int
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(api utils.APIInterface, hub *pubsub.PubSub) *WebSocketHandler {
	return &WebSocketHandler{
		api: api,
		hub: hub,
		upgrader: websocket.Upgrader{
			CheckOrigin:     isValidOrigin,
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		connections:       make(map[string]*WebSocketConnection),
		maxConnections:    50,          // Reasonable limit for internal network usage
		maxMessageSize:    1024 * 1024, // 1MB message size limit
		messagesPerMinute: 100,         // 100 messages per minute per connection
	}
}

// isValidOrigin validates WebSocket connection origins for local network usage
func isValidOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		// Allow connections without origin header (e.g., direct WebSocket clients)
		return true
	}

	// Parse the origin URL
	if !strings.HasPrefix(origin, "http://") && !strings.HasPrefix(origin, "https://") {
		return false
	}

	// Extract host from origin
	var host string
	if strings.HasPrefix(origin, "http://") {
		host = strings.TrimPrefix(origin, "http://")
	} else {
		host = strings.TrimPrefix(origin, "https://")
	}

	// Remove port if present
	if colonIndex := strings.Index(host, ":"); colonIndex != -1 {
		host = host[:colonIndex]
	}

	// Allow localhost and local network ranges
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return true
	}

	// Parse IP address
	ip := net.ParseIP(host)
	if ip != nil {
		// Allow private IP ranges (RFC 1918)
		return isPrivateIP(ip)
	}

	// Allow local domain names (no dots or ending with .local)
	if !strings.Contains(host, ".") || strings.HasSuffix(host, ".local") {
		return true
	}

	// Reject all other origins
	logger.Yellow("WebSocket connection rejected from origin: %s", origin)
	return false
}

// isPrivateIP checks if an IP address is in a private range
func isPrivateIP(ip net.IP) bool {
	// Private IPv4 ranges
	private4 := []string{
		"10.0.0.0/8",     // Class A private
		"172.16.0.0/12",  // Class B private
		"192.168.0.0/16", // Class C private
	}

	for _, cidr := range private4 {
		_, network, _ := net.ParseCIDR(cidr)
		if network.Contains(ip) {
			return true
		}
	}

	// Private IPv6 ranges
	if ip.To4() == nil { // IPv6
		// Link-local addresses (fe80::/10)
		if ip[0] == 0xfe && (ip[1]&0xc0) == 0x80 {
			return true
		}
		// Unique local addresses (fc00::/7)
		if (ip[0] & 0xfe) == 0xfc {
			return true
		}
	}

	return false
}

// HandleWebSocket handles WebSocket connections with subscription management
func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	logger.Blue("WebSocket connection attempt from %s", r.RemoteAddr)

	// Check if API adapter is available
	if h.api == nil {
		logger.Red("WebSocket handler: API adapter is nil")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Check connection limit
	h.mutex.RLock()
	connectionCount := len(h.connections)
	h.mutex.RUnlock()

	if connectionCount >= h.maxConnections {
		logger.Yellow("WebSocket connection rejected: maximum connections (%d) reached", h.maxConnections)
		http.Error(w, "Maximum connections reached", http.StatusServiceUnavailable)
		return
	}

	// Set message size limit
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Red("WebSocket upgrade failed: %v", err)
		return
	}

	// Set read limit for message size
	conn.SetReadLimit(h.maxMessageSize)

	// Get client IP
	clientIP := getClientIP(r)

	// Create connection context
	ctx, cancel := context.WithCancel(context.Background())

	// Create WebSocket connection object
	wsConn := &WebSocketConnection{
		conn:          conn,
		subscriptions: make(map[EventType]bool),
		ctx:           ctx,
		cancel:        cancel,
		clientIP:      clientIP,
		messageCount:  0,
		lastMessage:   time.Now(),
	}

	// Generate connection ID
	connID := fmt.Sprintf("ws_%d", time.Now().UnixNano())

	// Store connection
	h.mutex.Lock()
	h.connections[connID] = wsConn
	h.mutex.Unlock()

	logger.Green("WebSocket connection established: %s from %s", connID, clientIP)

	// Handle connection cleanup
	defer func() {
		h.mutex.Lock()
		delete(h.connections, connID)
		h.mutex.Unlock()
		cancel()
		conn.Close()
		logger.Blue("WebSocket connection closed: %s", connID)
	}()

	// Start message handler
	go h.handleMessages(wsConn)

	// Start event broadcaster
	h.startEventBroadcaster(wsConn)
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the list
		if commaIndex := strings.Index(xff, ","); commaIndex != -1 {
			return strings.TrimSpace(xff[:commaIndex])
		}
		return strings.TrimSpace(xff)
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// handleMessages handles incoming WebSocket messages
func (h *WebSocketHandler) handleMessages(wsConn *WebSocketConnection) {
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
func (h *WebSocketHandler) processMessage(wsConn *WebSocketConnection, data []byte) error {
	// Check rate limiting
	if !h.checkRateLimit(wsConn) {
		return fmt.Errorf("rate limit exceeded")
	}

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

// checkRateLimit checks if the connection is within rate limits
func (h *WebSocketHandler) checkRateLimit(wsConn *WebSocketConnection) bool {
	now := time.Now()

	// Reset counter if more than a minute has passed
	if now.Sub(wsConn.lastMessage) > time.Minute {
		wsConn.messageCount = 0
		wsConn.lastMessage = now
	}

	// Increment message count
	wsConn.messageCount++

	// Check if rate limit is exceeded
	if wsConn.messageCount > h.messagesPerMinute {
		logger.Yellow("Rate limit exceeded for WebSocket connection from %s: %d messages in last minute",
			wsConn.clientIP, wsConn.messageCount)
		return false
	}

	return true
}

// handlePing responds to ping messages
func (h *WebSocketHandler) handlePing(wsConn *WebSocketConnection) error {
	response := map[string]interface{}{
		"type":      "pong",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	return wsConn.conn.WriteJSON(response)
}

// handleSubscribe handles subscription requests
func (h *WebSocketHandler) handleSubscribe(wsConn *WebSocketConnection, message map[string]interface{}) error {
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
func (h *WebSocketHandler) handleUnsubscribe(wsConn *WebSocketConnection, message map[string]interface{}) error {
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
func (h *WebSocketHandler) startEventBroadcaster(wsConn *WebSocketConnection) {
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
func (h *WebSocketHandler) broadcastEvent(wsConn *WebSocketConnection, event interface{}) error {
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
func (h *WebSocketHandler) BroadcastEvent(eventType EventType, data interface{}) {
	event := map[string]interface{}{
		"event_type": string(eventType),
		"data":       data,
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	}

	// Publish to pubsub system
	h.hub.Pub(event, string(eventType))
}
