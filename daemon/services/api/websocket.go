package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/domalab/uma/daemon/logger"
	"github.com/gorilla/websocket"
)

// WebSocketManager manages WebSocket connections and broadcasting
type WebSocketManager struct {
	upgrader    websocket.Upgrader
	connections map[string]map[*websocket.Conn]*WebSocketClient
	mutex       sync.RWMutex
	httpServer  *HTTPServer
}

// WebSocketClient represents a connected WebSocket client
type WebSocketClient struct {
	conn      *websocket.Conn
	endpoint  string
	requestID string
	version   string
	send      chan []byte
	done      chan struct{}
}

// WebSocketMessage represents a WebSocket message
type WebSocketMessage struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp string      `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
	Version   string      `json:"version,omitempty"`
}

// NewWebSocketManager creates a new WebSocket manager
func NewWebSocketManager(httpServer *HTTPServer) *WebSocketManager {
	return &WebSocketManager{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Allow all origins for now - in production, implement proper origin checking
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		connections: make(map[string]map[*websocket.Conn]*WebSocketClient),
		httpServer:  httpServer,
	}
}

// handleWebSocketUpgrade upgrades HTTP connection to WebSocket
func (wsm *WebSocketManager) handleWebSocketUpgrade(w http.ResponseWriter, r *http.Request, endpoint string) {
	// Extract request metadata
	requestID := wsm.httpServer.getRequestIDFromContext(r)
	if requestID == "" {
		requestID = wsm.httpServer.generateRequestID()
	}
	version := wsm.httpServer.getAPIVersionFromContext(r)

	// Upgrade connection
	conn, err := wsm.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Red("WebSocket upgrade failed for %s: %v", endpoint, err)
		return
	}

	// Create client
	client := &WebSocketClient{
		conn:      conn,
		endpoint:  endpoint,
		requestID: requestID,
		version:   version,
		send:      make(chan []byte, 256),
		done:      make(chan struct{}),
	}

	// Register client
	wsm.registerClient(endpoint, conn, client)

	// Record connection metrics
	connectionCount := len(wsm.connections[endpoint])
	RecordWebSocketConnection(endpoint, "connect", connectionCount)

	// Start client handlers
	go client.writePump()
	go client.readPump(wsm)

	logger.Green("WebSocket client connected to %s [%s]", endpoint, requestID)
}

// registerClient registers a new WebSocket client
func (wsm *WebSocketManager) registerClient(endpoint string, conn *websocket.Conn, client *WebSocketClient) {
	wsm.mutex.Lock()
	defer wsm.mutex.Unlock()

	if wsm.connections[endpoint] == nil {
		wsm.connections[endpoint] = make(map[*websocket.Conn]*WebSocketClient)
	}
	wsm.connections[endpoint][conn] = client
}

// unregisterClient removes a WebSocket client
func (wsm *WebSocketManager) unregisterClient(endpoint string, conn *websocket.Conn) {
	wsm.mutex.Lock()
	defer wsm.mutex.Unlock()

	if clients, exists := wsm.connections[endpoint]; exists {
		if client, exists := clients[conn]; exists {
			close(client.done)
			close(client.send)
			delete(clients, conn)
			conn.Close()

			// Record disconnection metrics
			connectionCount := len(clients) - 1 // -1 because we just removed it
			RecordWebSocketConnection(endpoint, "disconnect", connectionCount)
		}
	}
}

// broadcast sends a message to all clients connected to a specific endpoint
func (wsm *WebSocketManager) broadcast(endpoint string, messageType string, data interface{}) {
	wsm.mutex.RLock()
	clients := wsm.connections[endpoint]
	wsm.mutex.RUnlock()

	if len(clients) == 0 {
		return
	}

	message := WebSocketMessage{
		Type:      messageType,
		Data:      data,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		logger.Red("Failed to marshal WebSocket message: %v", err)
		return
	}

	messagesSent := 0
	for conn, client := range clients {
		select {
		case client.send <- messageBytes:
			messagesSent++
		default:
			// Client send buffer is full, disconnect
			wsm.unregisterClient(endpoint, conn)
		}
	}

	// Record message metrics
	if messagesSent > 0 {
		RecordWebSocketMessage(endpoint, messageType)
	}
}

// writePump handles writing messages to the WebSocket connection
func (client *WebSocketClient) writePump() {
	ticker := time.NewTicker(54 * time.Second) // Ping every 54 seconds
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
				logger.Red("WebSocket write error: %v", err)
				return
			}

		case <-ticker.C:
			client.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-client.done:
			return
		}
	}
}

// readPump handles reading messages from the WebSocket connection
func (client *WebSocketClient) readPump(wsm *WebSocketManager) {
	defer func() {
		wsm.unregisterClient(client.endpoint, client.conn)
	}()

	client.conn.SetReadLimit(512)
	client.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.conn.SetPongHandler(func(string) error {
		client.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Red("WebSocket read error: %v", err)
			}
			break
		}
	}
}

// WebSocket endpoint handlers

// handleSystemStatsWebSocket handles /api/v1/ws/system/stats
func (h *HTTPServer) handleSystemStatsWebSocket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	h.wsManager.handleWebSocketUpgrade(w, r, "system/stats")
}

// handleDockerEventsWebSocket handles /api/v1/ws/docker/events
func (h *HTTPServer) handleDockerEventsWebSocket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	h.wsManager.handleWebSocketUpgrade(w, r, "docker/events")
}

// handleStorageStatusWebSocket handles /api/v1/ws/storage/status
func (h *HTTPServer) handleStorageStatusWebSocket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	h.wsManager.handleWebSocketUpgrade(w, r, "storage/status")
}

// startWebSocketBroadcasters starts background goroutines for broadcasting real-time data
func (h *HTTPServer) startWebSocketBroadcasters() {
	// Start system stats broadcaster
	go h.systemStatsBroadcaster()

	// Start Docker events broadcaster
	go h.dockerEventsBroadcaster()

	// Start storage status broadcaster
	go h.storageStatusBroadcaster()
}

// systemStatsBroadcaster broadcasts system statistics every 5 seconds
func (h *HTTPServer) systemStatsBroadcaster() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		stats := h.collectSystemStats()
		h.wsManager.broadcast("system/stats", "system_stats", stats)
	}
}

// collectSystemStats collects current system statistics
func (h *HTTPServer) collectSystemStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// Get CPU info
	if cpuInfo, err := h.api.system.GetCPUInfo(); err == nil {
		stats["cpu"] = map[string]interface{}{
			"usage_percent": cpuInfo.Usage,
			"temperature":   cpuInfo.Temperature,
			"cores":         cpuInfo.Cores,
		}
	}

	// Get memory info
	if memInfo, err := h.api.system.GetMemoryInfo(); err == nil {
		stats["memory"] = map[string]interface{}{
			"total":           memInfo.Total,
			"used":            memInfo.Used,
			"free":            memInfo.Free,
			"used_percent":    memInfo.UsedPercent,
			"total_formatted": memInfo.TotalFormatted,
			"used_formatted":  memInfo.UsedFormatted,
		}
	}

	// Get uptime
	if uptimeInfo, err := h.api.system.GetUptimeInfo(); err == nil {
		stats["uptime"] = map[string]interface{}{
			"seconds":   uptimeInfo.Uptime,
			"formatted": fmt.Sprintf("%.0f seconds", uptimeInfo.Uptime),
		}
	}

	return stats
}

// dockerEventsBroadcaster broadcasts Docker events
func (h *HTTPServer) dockerEventsBroadcaster() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	var lastContainerStates map[string]string

	for range ticker.C {
		events := h.collectDockerEvents(&lastContainerStates)
		if len(events) > 0 {
			h.wsManager.broadcast("docker/events", "docker_events", events)
		}
	}
}

// collectDockerEvents collects Docker container state changes
func (h *HTTPServer) collectDockerEvents(lastStates *map[string]string) []map[string]interface{} {
	var events []map[string]interface{}

	if *lastStates == nil {
		*lastStates = make(map[string]string)
	}

	// Get current container states
	containers, err := h.api.docker.ListContainers(true) // Include all containers
	if err != nil {
		logger.Red("Failed to get Docker containers for events: %v", err)
		return events
	}

	currentStates := make(map[string]string)

	for _, container := range containers {
		containerID := container.ID
		currentState := container.State
		currentStates[containerID] = currentState

		// Check for state changes
		if lastState, exists := (*lastStates)[containerID]; exists {
			if lastState != currentState {
				// State changed
				events = append(events, map[string]interface{}{
					"container_id":   containerID,
					"container_name": container.Name,
					"previous_state": lastState,
					"current_state":  currentState,
					"event_type":     "state_change",
					"timestamp":      time.Now().UTC().Format(time.RFC3339),
				})
			}
		} else {
			// New container detected
			events = append(events, map[string]interface{}{
				"container_id":   containerID,
				"container_name": container.Name,
				"current_state":  currentState,
				"event_type":     "container_discovered",
				"timestamp":      time.Now().UTC().Format(time.RFC3339),
			})
		}
	}

	// Check for removed containers
	for containerID, lastState := range *lastStates {
		if _, exists := currentStates[containerID]; !exists {
			events = append(events, map[string]interface{}{
				"container_id": containerID,
				"last_state":   lastState,
				"event_type":   "container_removed",
				"timestamp":    time.Now().UTC().Format(time.RFC3339),
			})
		}
	}

	*lastStates = currentStates
	return events
}

// storageStatusBroadcaster broadcasts storage status updates
func (h *HTTPServer) storageStatusBroadcaster() {
	ticker := time.NewTicker(30 * time.Second) // Check storage every 30 seconds
	defer ticker.Stop()

	var lastDiskStates map[string]string

	for range ticker.C {
		updates := h.collectStorageUpdates(&lastDiskStates)
		if len(updates) > 0 {
			h.wsManager.broadcast("storage/status", "storage_updates", updates)
		}
	}
}

// collectStorageUpdates collects storage status changes
func (h *HTTPServer) collectStorageUpdates(lastStates *map[string]string) []map[string]interface{} {
	var updates []map[string]interface{}

	if *lastStates == nil {
		*lastStates = make(map[string]string)
	}

	// Get current disk information
	disksInfo, err := h.api.storage.GetConsolidatedDisksInfo()
	if err != nil {
		logger.Red("Failed to get disk info for storage updates: %v", err)
		return updates
	}

	currentStates := make(map[string]string)

	// Check all disk types
	allDisks := make([]interface{}, 0)
	for _, disk := range disksInfo.ArrayDisks {
		allDisks = append(allDisks, disk)
	}
	for _, disk := range disksInfo.ParityDisks {
		allDisks = append(allDisks, disk)
	}
	for _, disk := range disksInfo.CacheDisks {
		allDisks = append(allDisks, disk)
	}
	if disksInfo.BootDisk != nil {
		allDisks = append(allDisks, *disksInfo.BootDisk)
	}

	for _, diskInterface := range allDisks {
		// Type assertion to get disk information
		diskMap, ok := diskInterface.(map[string]interface{})
		if !ok {
			continue
		}

		device, _ := diskMap["device"].(string)
		health, _ := diskMap["health"].(string)
		status, _ := diskMap["status"].(string)

		if device == "" {
			continue
		}

		currentState := fmt.Sprintf("%s:%s", status, health)
		currentStates[device] = currentState

		// Check for state changes
		if lastState, exists := (*lastStates)[device]; exists {
			if lastState != currentState {
				updates = append(updates, map[string]interface{}{
					"device":         device,
					"previous_state": lastState,
					"current_state":  currentState,
					"health":         health,
					"status":         status,
					"event_type":     "disk_state_change",
					"timestamp":      time.Now().UTC().Format(time.RFC3339),
				})
			}
		} else {
			// New disk detected
			updates = append(updates, map[string]interface{}{
				"device":        device,
				"current_state": currentState,
				"health":        health,
				"status":        status,
				"event_type":    "disk_discovered",
				"timestamp":     time.Now().UTC().Format(time.RFC3339),
			})
		}
	}

	*lastStates = currentStates
	return updates
}
