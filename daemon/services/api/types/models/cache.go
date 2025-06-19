package models

import (
	"sync"
	"time"
)

// CacheEntry represents a cached data entry with expiration
type CacheEntry struct {
	Data      interface{}
	ExpiresAt time.Time
}

// GeneralFormatCache caches expensive operations for the general format endpoint
type GeneralFormatCache struct {
	mu                sync.RWMutex
	systemData        *CacheEntry
	dockerData        *CacheEntry
	vmData            *CacheEntry
	cacheDuration     time.Duration
	lastArrayInfoHash string
}
