package utils

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/domalab/uma/daemon/services/api/types/responses"
)

// TestProductionReadinessChecker tests the production readiness checker
func TestProductionReadinessChecker(t *testing.T) {
	t.Run("NewProductionReadinessChecker", func(t *testing.T) {
		checker := NewProductionReadinessChecker()
		if checker == nil {
			t.Error("Expected non-nil ProductionReadinessChecker")
			return
		}

		if len(checker.checks) == 0 {
			t.Error("Expected default health checks to be registered")
		}

		// Check that default checks are registered
		expectedChecks := []string{"memory", "goroutines", "disk_space", "system_load"}
		for _, checkName := range expectedChecks {
			if _, exists := checker.checks[checkName]; !exists {
				t.Errorf("Expected default check '%s' to be registered", checkName)
			}
		}
	})

	t.Run("RegisterCheck", func(t *testing.T) {
		checker := NewProductionReadinessChecker()

		customCheck := func(ctx context.Context) responses.HealthCheck {
			return responses.HealthCheck{
				Status:  "pass",
				Message: "Custom check passed",
			}
		}

		checker.RegisterCheck("custom_check", customCheck)

		if _, exists := checker.checks["custom_check"]; !exists {
			t.Error("Expected custom check to be registered")
		}
	})

	t.Run("RunAllChecks", func(t *testing.T) {
		checker := NewProductionReadinessChecker()
		ctx := context.Background()

		results := checker.RunAllChecks(ctx)

		if len(results) == 0 {
			t.Error("Expected health check results")
		}

		// Verify all default checks ran
		expectedChecks := []string{"memory", "goroutines", "disk_space", "system_load"}
		for _, checkName := range expectedChecks {
			if result, exists := results[checkName]; !exists {
				t.Errorf("Expected result for check '%s'", checkName)
			} else {
				if result.Status == "" {
					t.Errorf("Expected status for check '%s'", checkName)
				}
				if result.Message == "" {
					t.Errorf("Expected message for check '%s'", checkName)
				}
				if result.Duration == "" {
					t.Errorf("Expected duration for check '%s'", checkName)
				}
			}
		}
	})

	t.Run("GetOverallStatus", func(t *testing.T) {
		checker := NewProductionReadinessChecker()

		// Test healthy status
		healthyChecks := map[string]responses.HealthCheck{
			"check1": {Status: "pass", Message: "OK"},
			"check2": {Status: "pass", Message: "OK"},
		}
		status := checker.GetOverallStatus(healthyChecks)
		if status != "healthy" {
			t.Errorf("Expected 'healthy' status, got '%s'", status)
		}

		// Test degraded status
		degradedChecks := map[string]responses.HealthCheck{
			"check1": {Status: "pass", Message: "OK"},
			"check2": {Status: "warn", Message: "Warning"},
		}
		status = checker.GetOverallStatus(degradedChecks)
		if status != "degraded" {
			t.Errorf("Expected 'degraded' status, got '%s'", status)
		}

		// Test unhealthy status
		unhealthyChecks := map[string]responses.HealthCheck{
			"check1": {Status: "pass", Message: "OK"},
			"check2": {Status: "fail", Message: "Failed"},
		}
		status = checker.GetOverallStatus(unhealthyChecks)
		if status != "unhealthy" {
			t.Errorf("Expected 'unhealthy' status, got '%s'", status)
		}
	})

	t.Run("MemoryCheck", func(t *testing.T) {
		checker := NewProductionReadinessChecker()
		ctx := context.Background()

		result := checker.checkMemoryUsage(ctx)

		if result.Status == "" {
			t.Error("Expected memory check to return a status")
		}
		if result.Message == "" {
			t.Error("Expected memory check to return a message")
		}

		// Status should be pass, warn, or fail
		validStatuses := []string{"pass", "warn", "fail"}
		found := false
		for _, validStatus := range validStatuses {
			if result.Status == validStatus {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected valid status, got '%s'", result.Status)
		}
	})

	t.Run("GoroutineCheck", func(t *testing.T) {
		checker := NewProductionReadinessChecker()
		ctx := context.Background()

		result := checker.checkGoroutineCount(ctx)

		if result.Status == "" {
			t.Error("Expected goroutine check to return a status")
		}
		if result.Message == "" {
			t.Error("Expected goroutine check to return a message")
		}

		// Should contain goroutine count in message
		if !strings.Contains(result.Message, "Goroutines:") {
			t.Error("Expected message to contain goroutine count")
		}
	})
}

// TestConfigurationValidator tests the configuration validator
func TestConfigurationValidator(t *testing.T) {
	t.Run("NewConfigurationValidator", func(t *testing.T) {
		validator := NewConfigurationValidator()
		if validator == nil {
			t.Error("Expected non-nil ConfigurationValidator")
			return
		}

		if len(validator.validationRules) == 0 {
			t.Error("Expected default validation rules to be registered")
		}
	})

	t.Run("ValidateConfiguration_Valid", func(t *testing.T) {
		validator := NewConfigurationValidator()

		config := map[string]interface{}{
			"port":      8080,
			"log_level": "info",
		}

		errors := validator.ValidateConfiguration(config)
		if len(errors) != 0 {
			t.Errorf("Expected no validation errors, got %d", len(errors))
		}
	})

	t.Run("ValidateConfiguration_MissingRequired", func(t *testing.T) {
		validator := NewConfigurationValidator()

		config := map[string]interface{}{
			"log_level": "info",
			// Missing required "port"
		}

		errors := validator.ValidateConfiguration(config)
		if len(errors) == 0 {
			t.Error("Expected validation errors for missing required field")
		}

		found := false
		for _, err := range errors {
			if err.Field == "port" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected validation error for missing 'port' field")
		}
	})

	t.Run("ValidateConfiguration_InvalidPort", func(t *testing.T) {
		validator := NewConfigurationValidator()

		config := map[string]interface{}{
			"port":      80, // Invalid port (< 1024)
			"log_level": "info",
		}

		errors := validator.ValidateConfiguration(config)
		if len(errors) == 0 {
			t.Error("Expected validation errors for invalid port")
		}

		found := false
		for _, err := range errors {
			if err.Field == "port" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected validation error for invalid port")
		}
	})

	t.Run("ValidateConfiguration_InvalidLogLevel", func(t *testing.T) {
		validator := NewConfigurationValidator()

		config := map[string]interface{}{
			"port":      8080,
			"log_level": "invalid", // Invalid log level
		}

		errors := validator.ValidateConfiguration(config)
		if len(errors) == 0 {
			t.Error("Expected validation errors for invalid log level")
		}

		found := false
		for _, err := range errors {
			if err.Field == "log_level" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected validation error for invalid log level")
		}
	})
}

// TestMonitoringCollector tests the monitoring collector
func TestMonitoringCollector(t *testing.T) {
	t.Run("NewMonitoringCollector", func(t *testing.T) {
		collector := NewMonitoringCollector()
		if collector == nil {
			t.Error("Expected non-nil MonitoringCollector")
			return
		}

		if collector.metrics == nil {
			t.Error("Expected non-nil metrics map")
		}
	})

	t.Run("RecordMetric", func(t *testing.T) {
		collector := NewMonitoringCollector()

		collector.RecordMetric("test_metric", 42)

		metrics := collector.GetMetrics()
		if value, exists := metrics["test_metric"]; !exists {
			t.Error("Expected recorded metric to exist")
		} else if value != 42 {
			t.Errorf("Expected metric value 42, got %v", value)
		}
	})

	t.Run("GetSystemMetrics", func(t *testing.T) {
		collector := NewMonitoringCollector()

		metrics := collector.GetSystemMetrics()

		expectedMetrics := []string{"goroutines", "memory_alloc", "memory_sys", "gc_runs", "last_gc", "uptime_seconds"}
		for _, metricName := range expectedMetrics {
			if _, exists := metrics[metricName]; !exists {
				t.Errorf("Expected system metric '%s' to exist", metricName)
			}
		}

		// Verify goroutines metric is reasonable
		if goroutines, ok := metrics["goroutines"].(int); ok {
			if goroutines <= 0 {
				t.Error("Expected positive goroutine count")
			}
		} else {
			t.Error("Expected goroutines metric to be an integer")
		}
	})

	t.Run("ConcurrentAccess", func(t *testing.T) {
		collector := NewMonitoringCollector()

		// Test concurrent access
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func(id int) {
				collector.RecordMetric("concurrent_test", id)
				_ = collector.GetMetrics()
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}

		// Should not panic or race
		metrics := collector.GetMetrics()
		if _, exists := metrics["concurrent_test"]; !exists {
			t.Error("Expected concurrent metric to exist")
		}
	})
}

// TestWriteProductionHealthResponse tests the production health response writer
func TestWriteProductionHealthResponse(t *testing.T) {
	t.Run("HealthyResponse", func(t *testing.T) {
		checker := NewProductionReadinessChecker()
		w := httptest.NewRecorder()

		WriteProductionHealthResponse(w, checker, "test-version")

		if w.Code != 200 {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
		}

		if w.Body.Len() == 0 {
			t.Error("Expected response body")
		}
	})

	t.Run("UnhealthyResponse", func(t *testing.T) {
		checker := NewProductionReadinessChecker()

		// Register a failing check
		checker.RegisterCheck("failing_check", func(ctx context.Context) responses.HealthCheck {
			return responses.HealthCheck{
				Status:  "fail",
				Message: "This check always fails",
			}
		})

		w := httptest.NewRecorder()
		WriteProductionHealthResponse(w, checker, "test-version")

		if w.Code != 503 {
			t.Errorf("Expected status 503 for unhealthy response, got %d", w.Code)
		}
	})
}
