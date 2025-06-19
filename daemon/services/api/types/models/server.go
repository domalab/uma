package models

// Context key types for type safety
type ContextKey string

const (
	RequestIDKey  ContextKey = "request_id"
	APIVersionKey ContextKey = "api_version"
)

// Legacy context key type for backward compatibility
type LegacyContextKey string

const (
	LegacyRequestIDKey  LegacyContextKey = "request_id"
	LegacyAPIVersionKey LegacyContextKey = "api_version"
)
