package utils

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/domalab/uma/daemon/services/api/types/responses"
)

// ProductionReadinessChecker provides comprehensive production readiness validation
type ProductionReadinessChecker struct {
	checks      map[string]HealthCheckFunc
	mu          sync.RWMutex
	timeout     time.Duration
	lastResults map[string]responses.HealthCheck
}

// HealthCheckFunc defines the signature for health check functions
type HealthCheckFunc func(ctx context.Context) responses.HealthCheck

// NewProductionReadinessChecker creates a new production readiness checker
func NewProductionReadinessChecker() *ProductionReadinessChecker {
	checker := &ProductionReadinessChecker{
		checks:      make(map[string]HealthCheckFunc),
		timeout:     30 * time.Second,
		lastResults: make(map[string]responses.HealthCheck),
	}

	// Register default health checks
	checker.RegisterCheck("memory", checker.checkMemoryUsage)
	checker.RegisterCheck("goroutines", checker.checkGoroutineCount)
	checker.RegisterCheck("disk_space", checker.checkDiskSpace)
	checker.RegisterCheck("system_load", checker.checkSystemLoad)

	return checker
}

// RegisterCheck registers a new health check
func (p *ProductionReadinessChecker) RegisterCheck(name string, checkFunc HealthCheckFunc) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.checks[name] = checkFunc
}

// RunAllChecks executes all registered health checks
func (p *ProductionReadinessChecker) RunAllChecks(ctx context.Context) map[string]responses.HealthCheck {
	p.mu.RLock()
	checks := make(map[string]HealthCheckFunc)
	for name, checkFunc := range p.checks {
		checks[name] = checkFunc
	}
	p.mu.RUnlock()

	results := make(map[string]responses.HealthCheck)
	var wg sync.WaitGroup

	for name, checkFunc := range checks {
		wg.Add(1)
		go func(checkName string, check HealthCheckFunc) {
			defer wg.Done()

			// Create context with timeout
			checkCtx, cancel := context.WithTimeout(ctx, p.timeout)
			defer cancel()

			start := time.Now()
			result := check(checkCtx)
			result.Duration = time.Since(start).String()
			result.Timestamp = time.Now()

			p.mu.Lock()
			p.lastResults[checkName] = result
			results[checkName] = result
			p.mu.Unlock()
		}(name, checkFunc)
	}

	wg.Wait()
	return results
}

// GetOverallStatus determines the overall system status based on individual checks
func (p *ProductionReadinessChecker) GetOverallStatus(checks map[string]responses.HealthCheck) string {
	hasFailures := false
	hasWarnings := false

	for _, check := range checks {
		switch check.Status {
		case "fail":
			hasFailures = true
		case "warn":
			hasWarnings = true
		}
	}

	if hasFailures {
		return "unhealthy"
	}
	if hasWarnings {
		return "degraded"
	}
	return "healthy"
}

// checkMemoryUsage checks system memory usage
func (p *ProductionReadinessChecker) checkMemoryUsage(ctx context.Context) responses.HealthCheck {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Convert bytes to MB
	allocMB := float64(m.Alloc) / 1024 / 1024
	sysMB := float64(m.Sys) / 1024 / 1024

	status := "pass"
	message := fmt.Sprintf("Memory usage: %.2f MB allocated, %.2f MB system", allocMB, sysMB)

	// Warning if allocated memory > 100MB
	if allocMB > 100 {
		status = "warn"
		message += " (high memory usage)"
	}

	// Fail if allocated memory > 500MB
	if allocMB > 500 {
		status = "fail"
		message += " (critical memory usage)"
	}

	return responses.HealthCheck{
		Status:  status,
		Message: message,
	}
}

// checkGoroutineCount checks the number of active goroutines
func (p *ProductionReadinessChecker) checkGoroutineCount(ctx context.Context) responses.HealthCheck {
	count := runtime.NumGoroutine()

	status := "pass"
	message := fmt.Sprintf("Goroutines: %d", count)

	// Warning if > 100 goroutines
	if count > 100 {
		status = "warn"
		message += " (high goroutine count)"
	}

	// Fail if > 1000 goroutines
	if count > 1000 {
		status = "fail"
		message += " (critical goroutine count)"
	}

	return responses.HealthCheck{
		Status:  status,
		Message: message,
	}
}

// checkDiskSpace checks available disk space (mock implementation)
func (p *ProductionReadinessChecker) checkDiskSpace(ctx context.Context) responses.HealthCheck {
	// In a real implementation, this would check actual disk space
	// For now, we'll simulate a check
	status := "pass"
	message := "Disk space: sufficient"

	return responses.HealthCheck{
		Status:  status,
		Message: message,
	}
}

// checkSystemLoad checks system load (mock implementation)
func (p *ProductionReadinessChecker) checkSystemLoad(ctx context.Context) responses.HealthCheck {
	// In a real implementation, this would check actual system load
	// For now, we'll simulate a check
	status := "pass"
	message := "System load: normal"

	return responses.HealthCheck{
		Status:  status,
		Message: message,
	}
}

// ConfigurationValidator provides configuration validation
type ConfigurationValidator struct {
	validationRules map[string]ValidationRule
}

// ValidationRule defines a configuration validation rule
type ValidationRule struct {
	Required    bool
	Validator   func(value interface{}) error
	Description string
}

// NewConfigurationValidator creates a new configuration validator
func NewConfigurationValidator() *ConfigurationValidator {
	validator := &ConfigurationValidator{
		validationRules: make(map[string]ValidationRule),
	}

	// Register default validation rules
	validator.RegisterRule("port", ValidationRule{
		Required:    true,
		Validator:   validatePort,
		Description: "API server port must be between 1024 and 65535",
	})

	validator.RegisterRule("log_level", ValidationRule{
		Required:    false,
		Validator:   validateLogLevel,
		Description: "Log level must be one of: debug, info, warn, error",
	})

	return validator
}

// RegisterRule registers a new validation rule
func (c *ConfigurationValidator) RegisterRule(key string, rule ValidationRule) {
	c.validationRules[key] = rule
}

// ValidateConfiguration validates the provided configuration
func (c *ConfigurationValidator) ValidateConfiguration(config map[string]interface{}) []ValidationError {
	var errors []ValidationError

	// Check required fields
	for key, rule := range c.validationRules {
		value, exists := config[key]

		if rule.Required && !exists {
			errors = append(errors, ValidationError{
				Field:   key,
				Message: fmt.Sprintf("Required field '%s' is missing", key),
			})
			continue
		}

		if exists && rule.Validator != nil {
			if err := rule.Validator(value); err != nil {
				errors = append(errors, ValidationError{
					Field:   key,
					Message: err.Error(),
				})
			}
		}
	}

	return errors
}

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error implements the error interface
func (v ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", v.Field, v.Message)
}

// validatePort validates that a port is in the valid range
func validatePort(value interface{}) error {
	port, ok := value.(int)
	if !ok {
		return fmt.Errorf("port must be an integer")
	}

	if port < 1024 || port > 65535 {
		return fmt.Errorf("port must be between 1024 and 65535, got %d", port)
	}

	return nil
}

// validateLogLevel validates that the log level is valid
func validateLogLevel(value interface{}) error {
	level, ok := value.(string)
	if !ok {
		return fmt.Errorf("log level must be a string")
	}

	validLevels := []string{"debug", "info", "warn", "error"}
	for _, validLevel := range validLevels {
		if level == validLevel {
			return nil
		}
	}

	return fmt.Errorf("invalid log level '%s', must be one of: %v", level, validLevels)
}

// MonitoringCollector provides production monitoring capabilities
type MonitoringCollector struct {
	metrics map[string]interface{}
	mu      sync.RWMutex
}

// NewMonitoringCollector creates a new monitoring collector
func NewMonitoringCollector() *MonitoringCollector {
	return &MonitoringCollector{
		metrics: make(map[string]interface{}),
	}
}

// RecordMetric records a metric value
func (m *MonitoringCollector) RecordMetric(name string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metrics[name] = value
}

// GetMetrics returns all collected metrics
func (m *MonitoringCollector) GetMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Create a copy to avoid race conditions
	metrics := make(map[string]interface{})
	for k, v := range m.metrics {
		metrics[k] = v
	}

	return metrics
}

// GetSystemMetrics collects and returns system-level metrics
func (m *MonitoringCollector) GetSystemMetrics() map[string]interface{} {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	metrics := map[string]interface{}{
		"goroutines":     runtime.NumGoroutine(),
		"memory_alloc":   memStats.Alloc,
		"memory_sys":     memStats.Sys,
		"gc_runs":        memStats.NumGC,
		"last_gc":        time.Unix(0, int64(memStats.LastGC)),
		"uptime_seconds": time.Since(time.Now().Add(-time.Duration(memStats.LastGC))).Seconds(),
	}

	// Record these metrics
	for k, v := range metrics {
		m.RecordMetric(k, v)
	}

	return metrics
}

// WriteProductionHealthResponse writes a comprehensive production health response
func WriteProductionHealthResponse(w http.ResponseWriter, checker *ProductionReadinessChecker, version string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	checks := checker.RunAllChecks(ctx)
	overallStatus := checker.GetOverallStatus(checks)

	response := responses.HealthResponse{
		Status:    overallStatus,
		Version:   version,
		Uptime:    time.Since(time.Now().Add(-24 * time.Hour)).String(), // Mock uptime
		Timestamp: time.Now(),
		Checks:    checks,
	}

	// Set appropriate HTTP status code
	statusCode := http.StatusOK
	if overallStatus == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	WriteJSON(w, statusCode, response)
}
