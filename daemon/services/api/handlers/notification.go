package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/domalab/uma/daemon/logger"
	"github.com/domalab/uma/daemon/plugins/notifications"
	"github.com/domalab/uma/daemon/services/api/utils"
)

// NotificationHandler handles notification-related HTTP requests
type NotificationHandler struct {
	api utils.APIInterface
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(api utils.APIInterface) *NotificationHandler {
	return &NotificationHandler{
		api: api,
	}
}

// HandleNotifications handles GET /api/v1/notifications and POST /api/v1/notifications
func (h *NotificationHandler) HandleNotifications(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleGetNotifications(w, r)
	case http.MethodPost:
		h.handleCreateNotification(w, r)
	default:
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// HandleNotification handles GET/PUT/DELETE /api/v1/notifications/{id}
func (h *NotificationHandler) HandleNotification(w http.ResponseWriter, r *http.Request) {
	// Extract notification ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/notifications/")
	if path == "" {
		utils.WriteError(w, http.StatusBadRequest, "Notification ID required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.handleGetNotification(w, r, path)
	case http.MethodPut:
		h.handleUpdateNotification(w, r, path)
	case http.MethodDelete:
		h.handleDeleteNotification(w, r, path)
	default:
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// HandleNotificationsClear handles POST /api/v1/notifications/clear
func (h *NotificationHandler) HandleNotificationsClear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if err := h.api.GetNotifications().ClearAllNotifications(); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to clear notifications: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "All notifications cleared successfully",
	})
}

// HandleNotificationsStats handles GET /api/v1/notifications/stats
func (h *NotificationHandler) HandleNotificationsStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	stats, err := h.api.GetNotifications().GetNotificationStats()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get notification stats: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, stats)
}

// HandleNotificationsMarkAllRead handles POST /api/v1/notifications/mark-all-read
func (h *NotificationHandler) HandleNotificationsMarkAllRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if err := h.api.GetNotifications().MarkAllAsRead(); err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to mark all notifications as read: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "All notifications marked as read",
	})
}

// Helper methods

// handleGetNotifications handles GET /api/v1/notifications with filtering
func (h *NotificationHandler) handleGetNotifications(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	params := h.parsePaginationParams(r)
	usePagination := r.URL.Query().Get("page") != "" || r.URL.Query().Get("limit") != ""

	// Parse query parameters for filtering
	level := r.URL.Query().Get("level")
	category := r.URL.Query().Get("category")
	unreadOnly := r.URL.Query().Get("unread_only") == "true"

	var notificationList []*notifications.Notification
	var err error

	if usePagination {
		// Use paginated query
		result, err := h.api.GetNotifications().GetNotificationsPaginated(params.Page, params.Limit, level, category, unreadOnly)
		if err == nil {
			if notifList, ok := result.([]*notifications.Notification); ok {
				notificationList = notifList
			} else {
				// Handle interface{} slice conversion
				if resultSlice, ok := result.([]interface{}); ok {
					notificationList = make([]*notifications.Notification, 0, len(resultSlice))
					for _, item := range resultSlice {
						if notif, ok := item.(*notifications.Notification); ok {
							notificationList = append(notificationList, notif)
						}
					}
				}
			}
		}
	} else {
		// Use non-paginated query
		result, err := h.api.GetNotifications().GetNotifications(level, category, unreadOnly)
		if err == nil {
			if notifList, ok := result.([]*notifications.Notification); ok {
				notificationList = notifList
			} else {
				// Handle interface{} slice conversion
				if resultSlice, ok := result.([]interface{}); ok {
					notificationList = make([]*notifications.Notification, 0, len(resultSlice))
					for _, item := range resultSlice {
						if notif, ok := item.(*notifications.Notification); ok {
							notificationList = append(notificationList, notif)
						}
					}
				}
			}
		}
	}

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get notifications: %v", err))
		return
	}

	if usePagination {
		// Get total count for pagination
		totalCount, err := h.api.GetNotifications().GetNotificationCount(level, category, unreadOnly)
		if err != nil {
			logger.Red("Failed to get notification count: %v", err)
			totalCount = len(notificationList) // Fallback
		}

		response := map[string]interface{}{
			"notifications": notificationList,
			"pagination": map[string]interface{}{
				"page":        params.Page,
				"limit":       params.Limit,
				"total_count": totalCount,
				"total_pages": (totalCount + params.Limit - 1) / params.Limit,
			},
		}
		utils.WriteJSON(w, http.StatusOK, response)
	} else {
		utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"notifications": notificationList,
		})
	}
}

// handleCreateNotification handles POST /api/v1/notifications
func (h *NotificationHandler) handleCreateNotification(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title    string                             `json:"title"`
		Message  string                             `json:"message"`
		Level    notifications.NotificationLevel    `json:"level"`
		Category notifications.NotificationCategory `json:"category"`
		Metadata map[string]interface{}             `json:"metadata,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Validate required fields
	if req.Title == "" || req.Message == "" {
		utils.WriteError(w, http.StatusBadRequest, "Title and message are required")
		return
	}

	notification, err := h.api.GetNotifications().CreateNotification(req.Title, req.Message, req.Level, req.Category, req.Metadata)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create notification: %v", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, notification)
}

// handleGetNotification handles GET /api/v1/notifications/{id}
func (h *NotificationHandler) handleGetNotification(w http.ResponseWriter, r *http.Request, id string) {
	notification, err := h.api.GetNotifications().GetNotification(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.WriteError(w, http.StatusNotFound, "Notification not found")
		} else {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get notification: %v", err))
		}
		return
	}

	utils.WriteJSON(w, http.StatusOK, notification)
}

// handleUpdateNotification handles PUT /api/v1/notifications/{id}
func (h *NotificationHandler) handleUpdateNotification(w http.ResponseWriter, r *http.Request, id string) {
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	notification, err := h.api.GetNotifications().UpdateNotification(id, updates)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.WriteError(w, http.StatusNotFound, "Notification not found")
		} else {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to update notification: %v", err))
		}
		return
	}

	utils.WriteJSON(w, http.StatusOK, notification)
}

// handleDeleteNotification handles DELETE /api/v1/notifications/{id}
func (h *NotificationHandler) handleDeleteNotification(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.api.GetNotifications().DeleteNotification(id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.WriteError(w, http.StatusNotFound, "Notification not found")
		} else {
			utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to delete notification: %v", err))
		}
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Notification deleted successfully",
	})
}

// Helper methods for pagination

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Page  int
	Limit int
}

// parsePaginationParams parses pagination parameters from request
func (h *NotificationHandler) parsePaginationParams(r *http.Request) PaginationParams {
	params := PaginationParams{
		Page:  1,
		Limit: 50, // Default limit
	}

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			params.Page = page
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 1000 {
			params.Limit = limit
		}
	}

	return params
}
