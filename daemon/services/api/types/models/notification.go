package models

import "time"

// Notification represents a system notification
type Notification struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"` // "info", "warning", "error", "success"
	Title      string                 `json:"title"`
	Message    string                 `json:"message"`
	Source     string                 `json:"source"`   // Source component/service
	Category   string                 `json:"category"` // "system", "storage", "docker", "vm", "security"
	Priority   int                    `json:"priority"` // 1-5 (1=lowest, 5=highest)
	Read       bool                   `json:"read"`
	Persistent bool                   `json:"persistent"` // Should persist across restarts
	Actions    []NotificationAction   `json:"actions,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Created    time.Time              `json:"created"`
	Updated    time.Time              `json:"updated"`
	ExpiresAt  *time.Time             `json:"expires_at,omitempty"`
}

// NotificationAction represents an action that can be taken on a notification
type NotificationAction struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Action      string `json:"action"` // "dismiss", "acknowledge", "resolve", "custom"
	URL         string `json:"url,omitempty"`
	Method      string `json:"method,omitempty"` // HTTP method for URL actions
	Destructive bool   `json:"destructive"`      // Indicates destructive action
}

// NotificationRule represents a notification rule configuration
type NotificationRule struct {
	ID          string                   `json:"id"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Enabled     bool                     `json:"enabled"`
	Conditions  []NotificationCondition  `json:"conditions"`
	Actions     []NotificationRuleAction `json:"actions"`
	Cooldown    int                      `json:"cooldown"` // Cooldown period in seconds
	Created     time.Time                `json:"created"`
	Updated     time.Time                `json:"updated"`
}

// NotificationCondition represents a condition for triggering notifications
type NotificationCondition struct {
	Field    string      `json:"field"`    // Field to check
	Operator string      `json:"operator"` // "eq", "ne", "gt", "lt", "contains", "regex"
	Value    interface{} `json:"value"`    // Value to compare against
}

// NotificationRuleAction represents an action to take when a rule is triggered
type NotificationRuleAction struct {
	Type       string                 `json:"type"`     // "email", "webhook", "sms", "push"
	Target     string                 `json:"target"`   // Email address, webhook URL, etc.
	Template   string                 `json:"template"` // Message template
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// NotificationChannel represents a notification delivery channel
type NotificationChannel struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"` // "email", "webhook", "slack", "discord", "telegram"
	Enabled     bool                   `json:"enabled"`
	Config      map[string]interface{} `json:"config"` // Channel-specific configuration
	TestMessage string                 `json:"test_message,omitempty"`
	LastTest    *time.Time             `json:"last_test,omitempty"`
	TestResult  string                 `json:"test_result,omitempty"` // "success", "failed"
	Created     time.Time              `json:"created"`
	Updated     time.Time              `json:"updated"`
}

// NotificationStats represents notification statistics
type NotificationStats struct {
	Total       int            `json:"total"`
	Unread      int            `json:"unread"`
	ByType      map[string]int `json:"by_type"`
	ByCategory  map[string]int `json:"by_category"`
	ByPriority  map[string]int `json:"by_priority"`
	LastUpdated time.Time      `json:"last_updated"`
}
