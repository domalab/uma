package notifications

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/domalab/uma/daemon/lib"
	"github.com/domalab/uma/daemon/logger"
)

// NotificationLevel represents the severity level of a notification
type NotificationLevel string

const (
	LevelInfo     NotificationLevel = "info"
	LevelWarning  NotificationLevel = "warning"
	LevelError    NotificationLevel = "error"
	LevelCritical NotificationLevel = "critical"
)

// NotificationCategory represents the category/source of a notification
type NotificationCategory string

const (
	CategorySystem   NotificationCategory = "system"
	CategoryArray    NotificationCategory = "array"
	CategoryDocker   NotificationCategory = "docker"
	CategoryVM       NotificationCategory = "vm"
	CategoryStorage  NotificationCategory = "storage"
	CategoryNetwork  NotificationCategory = "network"
	CategorySecurity NotificationCategory = "security"
	CategoryCustom   NotificationCategory = "custom"
)

// Notification represents a system notification
type Notification struct {
	ID         string               `json:"id"`
	Title      string               `json:"title"`
	Message    string               `json:"message"`
	Level      NotificationLevel    `json:"level"`
	Category   NotificationCategory `json:"category"`
	Timestamp  time.Time            `json:"timestamp"`
	Read       bool                 `json:"read"`
	Persistent bool                 `json:"persistent"`
	Source     string               `json:"source"`
	Actions    []NotificationAction `json:"actions,omitempty"`
	Metadata   map[string]string    `json:"metadata,omitempty"`
}

// NotificationAction represents an action that can be taken on a notification
type NotificationAction struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	URL         string `json:"url,omitempty"`
	Method      string `json:"method,omitempty"`
	Dangerous   bool   `json:"dangerous,omitempty"`
	Description string `json:"description,omitempty"`
}

// NotificationFilter represents filters for querying notifications
type NotificationFilter struct {
	Level      NotificationLevel    `json:"level,omitempty"`
	Category   NotificationCategory `json:"category,omitempty"`
	Read       *bool                `json:"read,omitempty"`
	Persistent *bool                `json:"persistent,omitempty"`
	Since      *time.Time           `json:"since,omitempty"`
	Until      *time.Time           `json:"until,omitempty"`
	Limit      int                  `json:"limit,omitempty"`
}

// NotificationManager manages system notifications
type NotificationManager struct {
	storageDir string
	nextID     int
}

// NewNotificationManager creates a new notification manager
func NewNotificationManager() *NotificationManager {
	storageDir := "/var/lib/uma/notifications"

	// Create storage directory if it doesn't exist
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		logger.Yellow("Failed to create notifications directory: %v", err)
		storageDir = "/tmp/uma-notifications"
		os.MkdirAll(storageDir, 0755)
	}

	nm := &NotificationManager{
		storageDir: storageDir,
		nextID:     1,
	}

	// Load existing notifications to determine next ID
	notifications, _ := nm.loadAllNotifications()
	for _, notification := range notifications {
		if id, err := strconv.Atoi(notification.ID); err == nil && id >= nm.nextID {
			nm.nextID = id + 1
		}
	}

	logger.Blue("Notification manager initialized with storage: %s", storageDir)
	return nm
}

// CreateNotification creates a new notification
func (nm *NotificationManager) CreateNotification(title, message string, level NotificationLevel, category NotificationCategory) (*Notification, error) {
	notification := &Notification{
		ID:         strconv.Itoa(nm.nextID),
		Title:      title,
		Message:    message,
		Level:      level,
		Category:   category,
		Timestamp:  time.Now(),
		Read:       false,
		Persistent: level == LevelError || level == LevelCritical,
		Source:     "uma",
		Metadata:   make(map[string]string),
	}

	nm.nextID++

	// Save notification to disk
	if err := nm.saveNotification(notification); err != nil {
		return nil, fmt.Errorf("failed to save notification: %v", err)
	}

	// Log to system log as well
	nm.logToSyslog(notification)

	logger.Blue("Created notification: %s [%s] %s", notification.Level, notification.Category, notification.Title)
	return notification, nil
}

// GetNotifications retrieves notifications with optional filtering
func (nm *NotificationManager) GetNotifications(filter *NotificationFilter) ([]*Notification, error) {
	notifications, err := nm.loadAllNotifications()
	if err != nil {
		return nil, err
	}

	// Apply filters
	filtered := make([]*Notification, 0)
	for _, notification := range notifications {
		if nm.matchesFilter(notification, filter) {
			filtered = append(filtered, notification)
		}
	}

	// Sort by timestamp (newest first)
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Timestamp.After(filtered[j].Timestamp)
	})

	// Apply limit
	if filter != nil && filter.Limit > 0 && len(filtered) > filter.Limit {
		filtered = filtered[:filter.Limit]
	}

	return filtered, nil
}

// GetNotification retrieves a specific notification by ID
func (nm *NotificationManager) GetNotification(id string) (*Notification, error) {
	filePath := filepath.Join(nm.storageDir, id+".json")

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("notification not found")
		}
		return nil, fmt.Errorf("failed to read notification: %v", err)
	}

	var notification Notification
	if err := json.Unmarshal(data, &notification); err != nil {
		return nil, fmt.Errorf("failed to parse notification: %v", err)
	}

	return &notification, nil
}

// UpdateNotification updates an existing notification
func (nm *NotificationManager) UpdateNotification(id string, updates map[string]interface{}) (*Notification, error) {
	notification, err := nm.GetNotification(id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if title, ok := updates["title"].(string); ok {
		notification.Title = title
	}
	if message, ok := updates["message"].(string); ok {
		notification.Message = message
	}
	if read, ok := updates["read"].(bool); ok {
		notification.Read = read
	}
	if persistent, ok := updates["persistent"].(bool); ok {
		notification.Persistent = persistent
	}

	// Save updated notification
	if err := nm.saveNotification(notification); err != nil {
		return nil, fmt.Errorf("failed to save updated notification: %v", err)
	}

	logger.Blue("Updated notification: %s", id)
	return notification, nil
}

// DeleteNotification deletes a notification
func (nm *NotificationManager) DeleteNotification(id string) error {
	filePath := filepath.Join(nm.storageDir, id+".json")

	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("notification not found")
		}
		return fmt.Errorf("failed to delete notification: %v", err)
	}

	logger.Blue("Deleted notification: %s", id)
	return nil
}

// ClearAllNotifications clears all notifications
func (nm *NotificationManager) ClearAllNotifications() error {
	files, err := ioutil.ReadDir(nm.storageDir)
	if err != nil {
		return fmt.Errorf("failed to read notifications directory: %v", err)
	}

	count := 0
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".json") {
			filePath := filepath.Join(nm.storageDir, file.Name())
			if err := os.Remove(filePath); err != nil {
				logger.Yellow("Failed to delete notification file %s: %v", file.Name(), err)
			} else {
				count++
			}
		}
	}

	logger.Blue("Cleared %d notifications", count)
	return nil
}

// MarkAllAsRead marks all notifications as read
func (nm *NotificationManager) MarkAllAsRead() error {
	notifications, err := nm.loadAllNotifications()
	if err != nil {
		return err
	}

	count := 0
	for _, notification := range notifications {
		if !notification.Read {
			notification.Read = true
			if err := nm.saveNotification(notification); err != nil {
				logger.Yellow("Failed to update notification %s: %v", notification.ID, err)
			} else {
				count++
			}
		}
	}

	logger.Blue("Marked %d notifications as read", count)
	return nil
}

// GetNotificationStats returns statistics about notifications
func (nm *NotificationManager) GetNotificationStats() (map[string]interface{}, error) {
	notifications, err := nm.loadAllNotifications()
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total":       len(notifications),
		"unread":      0,
		"persistent":  0,
		"by_level":    make(map[string]int),
		"by_category": make(map[string]int),
	}

	for _, notification := range notifications {
		if !notification.Read {
			stats["unread"] = stats["unread"].(int) + 1
		}
		if notification.Persistent {
			stats["persistent"] = stats["persistent"].(int) + 1
		}

		levelStats := stats["by_level"].(map[string]int)
		levelStats[string(notification.Level)]++

		categoryStats := stats["by_category"].(map[string]int)
		categoryStats[string(notification.Category)]++
	}

	return stats, nil
}

// Helper methods

// saveNotification saves a notification to disk
func (nm *NotificationManager) saveNotification(notification *Notification) error {
	filePath := filepath.Join(nm.storageDir, notification.ID+".json")

	data, err := json.MarshalIndent(notification, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %v", err)
	}

	if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write notification file: %v", err)
	}

	return nil
}

// loadAllNotifications loads all notifications from disk
func (nm *NotificationManager) loadAllNotifications() ([]*Notification, error) {
	files, err := ioutil.ReadDir(nm.storageDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read notifications directory: %v", err)
	}

	notifications := make([]*Notification, 0)
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".json") {
			filePath := filepath.Join(nm.storageDir, file.Name())

			data, err := ioutil.ReadFile(filePath)
			if err != nil {
				logger.Yellow("Failed to read notification file %s: %v", file.Name(), err)
				continue
			}

			var notification Notification
			if err := json.Unmarshal(data, &notification); err != nil {
				logger.Yellow("Failed to parse notification file %s: %v", file.Name(), err)
				continue
			}

			notifications = append(notifications, &notification)
		}
	}

	return notifications, nil
}

// matchesFilter checks if a notification matches the given filter
func (nm *NotificationManager) matchesFilter(notification *Notification, filter *NotificationFilter) bool {
	if filter == nil {
		return true
	}

	if filter.Level != "" && notification.Level != filter.Level {
		return false
	}

	if filter.Category != "" && notification.Category != filter.Category {
		return false
	}

	if filter.Read != nil && notification.Read != *filter.Read {
		return false
	}

	if filter.Persistent != nil && notification.Persistent != *filter.Persistent {
		return false
	}

	if filter.Since != nil && notification.Timestamp.Before(*filter.Since) {
		return false
	}

	if filter.Until != nil && notification.Timestamp.After(*filter.Until) {
		return false
	}

	return true
}

// logToSyslog logs the notification to system log
func (nm *NotificationManager) logToSyslog(notification *Notification) {
	priority := "info"
	switch notification.Level {
	case LevelWarning:
		priority = "warning"
	case LevelError:
		priority = "err"
	case LevelCritical:
		priority = "crit"
	}

	message := fmt.Sprintf("OmniRaid [%s/%s]: %s - %s",
		notification.Category, notification.Level, notification.Title, notification.Message)

	// Use logger command to write to syslog
	lib.GetCmdOutput("logger", "-p", "daemon."+priority, message)
}

// CreateSystemNotification creates a notification from system events
func (nm *NotificationManager) CreateSystemNotification(title, message string, level NotificationLevel) (*Notification, error) {
	return nm.CreateNotification(title, message, level, CategorySystem)
}

// CreateArrayNotification creates an array-related notification
func (nm *NotificationManager) CreateArrayNotification(title, message string, level NotificationLevel) (*Notification, error) {
	return nm.CreateNotification(title, message, level, CategoryArray)
}

// CreateDockerNotification creates a Docker-related notification
func (nm *NotificationManager) CreateDockerNotification(title, message string, level NotificationLevel) (*Notification, error) {
	return nm.CreateNotification(title, message, level, CategoryDocker)
}

// CreateVMNotification creates a VM-related notification
func (nm *NotificationManager) CreateVMNotification(title, message string, level NotificationLevel) (*Notification, error) {
	return nm.CreateNotification(title, message, level, CategoryVM)
}

// CreateStorageNotification creates a storage-related notification
func (nm *NotificationManager) CreateStorageNotification(title, message string, level NotificationLevel) (*Notification, error) {
	return nm.CreateNotification(title, message, level, CategoryStorage)
}

// CleanupOldNotifications removes old non-persistent notifications
func (nm *NotificationManager) CleanupOldNotifications(maxAge time.Duration) error {
	notifications, err := nm.loadAllNotifications()
	if err != nil {
		return err
	}

	cutoff := time.Now().Add(-maxAge)
	count := 0

	for _, notification := range notifications {
		if !notification.Persistent && notification.Timestamp.Before(cutoff) {
			if err := nm.DeleteNotification(notification.ID); err != nil {
				logger.Yellow("Failed to delete old notification %s: %v", notification.ID, err)
			} else {
				count++
			}
		}
	}

	if count > 0 {
		logger.Blue("Cleaned up %d old notifications", count)
	}

	return nil
}
