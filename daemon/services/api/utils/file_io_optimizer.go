package utils

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/domalab/uma/daemon/logger"
)

// FileIOOptimizer provides intelligent batching for file I/O operations
type FileIOOptimizer struct {
	mutex sync.RWMutex

	// Batch configuration
	batchSize    int
	batchTimeout time.Duration

	// File content cache
	fileCache    map[string]*CachedFileContent
	cacheTimeout time.Duration

	// Batch processing
	pendingReads map[string][]chan FileReadResult
	batchTimer   *time.Timer

	// Statistics
	totalReads   int64
	cachedReads  int64
	batchedReads int64
}

// CachedFileContent represents cached file content with metadata
type CachedFileContent struct {
	Content     []string
	ModTime     time.Time
	Size        int64
	CachedAt    time.Time
	AccessCount int64
}

// FileReadResult represents the result of a file read operation
type FileReadResult struct {
	Content []string
	Error   error
	Cached  bool
}

// FileReadRequest represents a request to read a file
type FileReadRequest struct {
	Path     string
	Response chan FileReadResult
}

// NewFileIOOptimizer creates a new file I/O optimizer
func NewFileIOOptimizer() *FileIOOptimizer {
	return &FileIOOptimizer{
		batchSize:    10,                    // Batch up to 10 file reads
		batchTimeout: 50 * time.Millisecond, // Process batch after 50ms
		fileCache:    make(map[string]*CachedFileContent),
		cacheTimeout: 30 * time.Second, // Cache files for 30 seconds
		pendingReads: make(map[string][]chan FileReadResult),
	}
}

// ReadFile reads a file with intelligent caching and batching
func (fio *FileIOOptimizer) ReadFile(path string) ([]string, error) {
	// Check cache first
	if content, found := fio.getCachedContent(path); found {
		fio.mutex.Lock()
		fio.cachedReads++
		fio.mutex.Unlock()
		return content.Content, nil
	}

	// Create response channel
	responseChan := make(chan FileReadResult, 1)

	// Add to batch
	fio.addToBatch(path, responseChan)

	// Wait for result
	result := <-responseChan

	fio.mutex.Lock()
	fio.totalReads++
	if result.Error == nil && !result.Cached {
		fio.batchedReads++
	}
	fio.mutex.Unlock()

	return result.Content, result.Error
}

// ReadProcFile reads a file from /proc with optimized handling
func (fio *FileIOOptimizer) ReadProcFile(path string) ([]string, error) {
	// /proc files are typically small and change frequently, so we use shorter cache
	fullPath := filepath.Join("/proc", path)
	return fio.readFileWithCustomCache(fullPath, 5*time.Second)
}

// ReadUnraidConfigFile reads a file from /var/local/emhttp with optimized handling
func (fio *FileIOOptimizer) ReadUnraidConfigFile(path string) ([]string, error) {
	// Unraid config files change less frequently, so we can cache longer
	fullPath := filepath.Join("/var/local/emhttp", path)
	return fio.readFileWithCustomCache(fullPath, 60*time.Second)
}

// BatchReadFiles reads multiple files in a single batch operation
func (fio *FileIOOptimizer) BatchReadFiles(paths []string) (map[string][]string, error) {
	results := make(map[string][]string)
	errors := make([]string, 0)

	// Check cache for all files first
	uncachedPaths := make([]string, 0)
	for _, path := range paths {
		if content, found := fio.getCachedContent(path); found {
			results[path] = content.Content
			fio.mutex.Lock()
			fio.cachedReads++
			fio.mutex.Unlock()
		} else {
			uncachedPaths = append(uncachedPaths, path)
		}
	}

	// Read uncached files in batch
	if len(uncachedPaths) > 0 {
		batchResults := fio.performBatchRead(uncachedPaths)
		for path, result := range batchResults {
			if result.Error != nil {
				errors = append(errors, fmt.Sprintf("%s: %v", path, result.Error))
			} else {
				results[path] = result.Content
			}
		}
	}

	var err error
	if len(errors) > 0 {
		err = fmt.Errorf("batch read errors: %s", strings.Join(errors, "; "))
	}

	return results, err
}

// getCachedContent retrieves content from cache if valid
func (fio *FileIOOptimizer) getCachedContent(path string) (*CachedFileContent, bool) {
	fio.mutex.RLock()
	defer fio.mutex.RUnlock()

	cached, exists := fio.fileCache[path]
	if !exists {
		return nil, false
	}

	// Check if cache is still valid
	if time.Since(cached.CachedAt) > fio.cacheTimeout {
		return nil, false
	}

	// Check if file has been modified
	if stat, err := os.Stat(path); err == nil {
		if stat.ModTime().After(cached.ModTime) || stat.Size() != cached.Size {
			return nil, false
		}
	}

	// Update access count
	cached.AccessCount++

	return cached, true
}

// addToBatch adds a file read request to the current batch
func (fio *FileIOOptimizer) addToBatch(path string, responseChan chan FileReadResult) {
	fio.mutex.Lock()
	defer fio.mutex.Unlock()

	// Add to pending reads
	if _, exists := fio.pendingReads[path]; !exists {
		fio.pendingReads[path] = make([]chan FileReadResult, 0)
	}
	fio.pendingReads[path] = append(fio.pendingReads[path], responseChan)

	// Check if we should process the batch
	totalPending := 0
	for _, channels := range fio.pendingReads {
		totalPending += len(channels)
	}

	if totalPending >= fio.batchSize {
		// Process batch immediately
		go fio.processBatch()
	} else if fio.batchTimer == nil {
		// Start timer for batch processing
		fio.batchTimer = time.AfterFunc(fio.batchTimeout, fio.processBatch)
	}
}

// processBatch processes the current batch of file read requests
func (fio *FileIOOptimizer) processBatch() {
	fio.mutex.Lock()

	// Get current batch
	currentBatch := fio.pendingReads
	fio.pendingReads = make(map[string][]chan FileReadResult)

	// Reset timer
	if fio.batchTimer != nil {
		fio.batchTimer.Stop()
		fio.batchTimer = nil
	}

	fio.mutex.Unlock()

	// Process each file in the batch
	for path, channels := range currentBatch {
		result := fio.readSingleFile(path)

		// Send result to all waiting channels
		for _, ch := range channels {
			ch <- result
		}
	}
}

// performBatchRead performs a batch read operation for multiple files
func (fio *FileIOOptimizer) performBatchRead(paths []string) map[string]FileReadResult {
	results := make(map[string]FileReadResult)

	for _, path := range paths {
		results[path] = fio.readSingleFile(path)
	}

	return results
}

// readSingleFile reads a single file and caches the result
func (fio *FileIOOptimizer) readSingleFile(path string) FileReadResult {
	// Get file info
	stat, err := os.Stat(path)
	if err != nil {
		return FileReadResult{Error: err}
	}

	// Read file content
	file, err := os.Open(path)
	if err != nil {
		return FileReadResult{Error: err}
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return FileReadResult{Error: err}
	}

	// Cache the content
	fio.cacheContent(path, lines, stat.ModTime(), stat.Size())

	return FileReadResult{
		Content: lines,
		Error:   nil,
		Cached:  false,
	}
}

// readFileWithCustomCache reads a file with custom cache timeout
func (fio *FileIOOptimizer) readFileWithCustomCache(path string, cacheTimeout time.Duration) ([]string, error) {
	// Check cache with custom timeout
	fio.mutex.RLock()
	cached, exists := fio.fileCache[path]
	fio.mutex.RUnlock()

	if exists && time.Since(cached.CachedAt) <= cacheTimeout {
		// Check if file has been modified
		if stat, err := os.Stat(path); err == nil {
			if !stat.ModTime().After(cached.ModTime) && stat.Size() == cached.Size {
				fio.mutex.Lock()
				fio.cachedReads++
				cached.AccessCount++
				fio.mutex.Unlock()
				return cached.Content, nil
			}
		}
	}

	// Read file normally
	return fio.ReadFile(path)
}

// cacheContent stores file content in cache
func (fio *FileIOOptimizer) cacheContent(path string, content []string, modTime time.Time, size int64) {
	fio.mutex.Lock()
	defer fio.mutex.Unlock()

	fio.fileCache[path] = &CachedFileContent{
		Content:     content,
		ModTime:     modTime,
		Size:        size,
		CachedAt:    time.Now(),
		AccessCount: 1,
	}
}

// ClearCache clears the file content cache
func (fio *FileIOOptimizer) ClearCache() {
	fio.mutex.Lock()
	defer fio.mutex.Unlock()

	fio.fileCache = make(map[string]*CachedFileContent)
	logger.Blue("File I/O cache cleared")
}

// GetStats returns file I/O optimizer statistics
func (fio *FileIOOptimizer) GetStats() map[string]interface{} {
	fio.mutex.RLock()
	defer fio.mutex.RUnlock()

	cacheHitRate := float64(0)
	if fio.totalReads > 0 {
		cacheHitRate = float64(fio.cachedReads) / float64(fio.totalReads) * 100
	}

	return map[string]interface{}{
		"total_reads":    fio.totalReads,
		"cached_reads":   fio.cachedReads,
		"batched_reads":  fio.batchedReads,
		"cache_hit_rate": cacheHitRate,
		"cached_files":   len(fio.fileCache),
		"batch_size":     fio.batchSize,
		"cache_timeout":  fio.cacheTimeout.String(),
	}
}

// CleanupExpiredCache removes expired entries from cache
func (fio *FileIOOptimizer) CleanupExpiredCache() {
	fio.mutex.Lock()
	defer fio.mutex.Unlock()

	now := time.Now()
	expiredKeys := make([]string, 0)

	for path, cached := range fio.fileCache {
		if now.Sub(cached.CachedAt) > fio.cacheTimeout {
			expiredKeys = append(expiredKeys, path)
		}
	}

	for _, key := range expiredKeys {
		delete(fio.fileCache, key)
	}

	if len(expiredKeys) > 0 {
		logger.Blue("Cleaned up %d expired file cache entries", len(expiredKeys))
	}
}

// Global file I/O optimizer instance
var globalFileIOOptimizer = NewFileIOOptimizer()

// GetGlobalFileIOOptimizer returns the global file I/O optimizer
func GetGlobalFileIOOptimizer() *FileIOOptimizer {
	return globalFileIOOptimizer
}

// Helper functions for common file operations

// ReadProcFile reads a file from /proc using the global optimizer
func ReadProcFile(path string) ([]string, error) {
	return globalFileIOOptimizer.ReadProcFile(path)
}

// ReadUnraidConfigFile reads a file from /var/local/emhttp using the global optimizer
func ReadUnraidConfigFile(path string) ([]string, error) {
	return globalFileIOOptimizer.ReadUnraidConfigFile(path)
}

// BatchReadProcFiles reads multiple files from /proc in a batch
func BatchReadProcFiles(paths []string) (map[string][]string, error) {
	fullPaths := make([]string, len(paths))
	for i, path := range paths {
		fullPaths[i] = filepath.Join("/proc", path)
	}
	return globalFileIOOptimizer.BatchReadFiles(fullPaths)
}
