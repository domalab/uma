package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/domalab/uma/daemon/logger"
	"github.com/domalab/uma/daemon/services/api/utils"
	"github.com/gorilla/websocket"
)

// WebSocketHandler handles WebSocket-related HTTP requests
type WebSocketHandler struct {
	api      utils.APIInterface
	upgrader websocket.Upgrader
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(api utils.APIInterface) *WebSocketHandler {
	return &WebSocketHandler{
		api: api,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Allow all origins for now - should be configurable in production
				return true
			},
		},
	}
}

// HandleSystemStatsWebSocket handles WebSocket connections for system stats
func (h *WebSocketHandler) HandleSystemStatsWebSocket(w http.ResponseWriter, r *http.Request) {
	logger.Blue("WebSocket connection attempt for system stats")

	// Check if API adapter is available
	if h.api == nil {
		logger.Red("WebSocket handler: API adapter is nil")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Red("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	logger.Green("System stats WebSocket connection established")

	// Send initial stats immediately
	stats := h.getSystemStats()
	if err := conn.WriteJSON(stats); err != nil {
		logger.Yellow("Failed to send initial system stats: %v", err)
		return
	}

	// Send system stats every 5 seconds
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			stats := h.getSystemStats()
			if err := conn.WriteJSON(stats); err != nil {
				logger.Yellow("Failed to send system stats: %v", err)
				return
			}
		}
	}
}

// HandleDockerEventsWebSocket handles WebSocket connections for Docker events
func (h *WebSocketHandler) HandleDockerEventsWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Red("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	logger.Green("Docker events WebSocket connection established")

	// Send Docker events every 10 seconds
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			events := h.getDockerEvents()
			if err := conn.WriteJSON(events); err != nil {
				logger.Yellow("Failed to send Docker events: %v", err)
				return
			}
		}
	}
}

// HandleStorageStatusWebSocket handles WebSocket connections for storage status
func (h *WebSocketHandler) HandleStorageStatusWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Red("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	logger.Green("Storage status WebSocket connection established")

	// Send storage status every 15 seconds
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			status := h.getStorageStatus()
			if err := conn.WriteJSON(status); err != nil {
				logger.Yellow("Failed to send storage status: %v", err)
				return
			}
		}
	}
}

// Helper methods for WebSocket data

// getSystemStats returns current system statistics
func (h *WebSocketHandler) getSystemStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// Get CPU info
	if cpuInfo, err := h.api.GetSystem().GetCPUInfo(); err == nil {
		stats["cpu"] = cpuInfo
	} else {
		logger.Yellow("WebSocket: Failed to get CPU info: %v", err)
		stats["cpu"] = map[string]interface{}{"error": "failed to get CPU info"}
	}

	// Get memory info
	if memInfo, err := h.api.GetSystem().GetMemoryInfo(); err == nil {
		stats["memory"] = memInfo
	} else {
		logger.Yellow("WebSocket: Failed to get memory info: %v", err)
		stats["memory"] = map[string]interface{}{"error": "failed to get memory info"}
	}

	// Get load info
	if loadInfo, err := h.api.GetSystem().GetLoadInfo(); err == nil {
		stats["load"] = loadInfo
	} else {
		logger.Yellow("WebSocket: Failed to get load info: %v", err)
		stats["load"] = map[string]interface{}{"error": "failed to get load info"}
	}

	// Get network info
	if networkInfo, err := h.api.GetSystem().GetNetworkInfo(); err == nil {
		stats["network"] = networkInfo
	} else {
		logger.Yellow("WebSocket: Failed to get network info: %v", err)
		stats["network"] = map[string]interface{}{"error": "failed to get network info"}
	}

	stats["timestamp"] = time.Now().UTC().Format(time.RFC3339)
	stats["type"] = "system_stats"

	return stats
}

// getDockerEvents returns current Docker events/status
func (h *WebSocketHandler) getDockerEvents() map[string]interface{} {
	events := make(map[string]interface{})

	// Get container status
	if containers, err := h.api.GetDocker().GetContainers(); err == nil {
		events["containers"] = containers
	}

	// Get Docker system info
	if info, err := h.api.GetDocker().GetSystemInfo(); err == nil {
		events["system_info"] = info
	}

	events["timestamp"] = time.Now().UTC().Format(time.RFC3339)
	events["type"] = "docker_events"

	return events
}

// getStorageStatus returns current storage status
func (h *WebSocketHandler) getStorageStatus() map[string]interface{} {
	status := make(map[string]interface{})

	// Get array info
	if arrayInfo, err := h.api.GetStorage().GetArrayInfo(); err == nil {
		status["array"] = arrayInfo
	}

	// Get disk info
	if disks, err := h.api.GetStorage().GetDisks(); err == nil {
		status["disks"] = disks
	}

	// Get cache info
	if cacheInfo, err := h.api.GetStorage().GetCacheInfo(); err == nil {
		status["cache"] = cacheInfo
	}

	status["timestamp"] = time.Now().UTC().Format(time.RFC3339)
	status["type"] = "storage_status"

	return status
}

// HandleWebSocketMessage handles incoming WebSocket messages
func (h *WebSocketHandler) HandleWebSocketMessage(conn *websocket.Conn, messageType int, data []byte) error {
	var message map[string]interface{}
	if err := json.Unmarshal(data, &message); err != nil {
		return err
	}

	// Handle different message types
	msgType, ok := message["type"].(string)
	if !ok {
		return nil // Ignore messages without type
	}

	switch msgType {
	case "ping":
		// Respond to ping with pong
		response := map[string]interface{}{
			"type":      "pong",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}
		return conn.WriteJSON(response)

	case "subscribe":
		// Handle subscription requests
		return h.handleSubscription(conn, message)

	case "unsubscribe":
		// Handle unsubscription requests
		return h.handleUnsubscription(conn, message)

	default:
		logger.Yellow("Unknown WebSocket message type: %s", msgType)
	}

	return nil
}

// handleSubscription handles WebSocket subscription requests
func (h *WebSocketHandler) handleSubscription(conn *websocket.Conn, message map[string]interface{}) error {
	channel, ok := message["channel"].(string)
	if !ok {
		return nil
	}

	response := map[string]interface{}{
		"type":      "subscription_ack",
		"channel":   channel,
		"status":    "subscribed",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	return conn.WriteJSON(response)
}

// handleUnsubscription handles WebSocket unsubscription requests
func (h *WebSocketHandler) handleUnsubscription(conn *websocket.Conn, message map[string]interface{}) error {
	channel, ok := message["channel"].(string)
	if !ok {
		return nil
	}

	response := map[string]interface{}{
		"type":      "unsubscription_ack",
		"channel":   channel,
		"status":    "unsubscribed",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	return conn.WriteJSON(response)
}
