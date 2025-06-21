package models

import (
	"time"
)

// CacheEntry represents a cached data entry with expiration
type CacheEntry struct {
	Data      interface{}
	ExpiresAt time.Time
}

// GeneralFormatCache caches expensive operations for the general format endpoint
type GeneralFormatCache struct {
	// Removed unused fields: mu, systemData, dockerData, vmData, cacheDuration, lastArrayInfoHash
}
