package responses

import (
	"encoding/json"
	"testing"
	"time"
)

// TestCommonResponses tests common response types
func TestCommonResponses(t *testing.T) {
	t.Run("StandardResponse creation", func(t *testing.T) {
		response := &StandardResponse{
			Data:    map[string]interface{}{"result": "success"},
			Message: "Operation completed successfully",
			Meta: &ResponseMeta{
				RequestID: "req-123",
				Version:   "1.0.0",
				Timestamp: time.Now(),
			},
		}

		if response.Message == "" {
			t.Error("Message should not be empty")
		}
		if response.Data == nil {
			t.Error("Data should not be nil")
		}
		if response.Meta == nil {
			t.Error("Meta should not be nil")
		}
		if response.Meta.RequestID != "req-123" {
			t.Errorf("Expected RequestID 'req-123', got '%s'", response.Meta.RequestID)
		}
	})

	t.Run("OperationResponse creation", func(t *testing.T) {
		response := &OperationResponse{
			Success:     true,
			Message:     "Operation completed successfully",
			OperationID: "op-456",
		}

		if !response.Success {
			t.Error("Expected Success to be true")
		}
		if response.OperationID != "op-456" {
			t.Errorf("Expected OperationID 'op-456', got '%s'", response.OperationID)
		}
	})

	t.Run("BulkOperationResponse creation", func(t *testing.T) {
		results := []BulkOperationResult{
			{ID: "1", Success: true, Message: "Success"},
			{ID: "2", Success: false, Error: "Failed"},
		}

		response := &BulkOperationResponse{
			Success: true,
			Message: "Bulk operation completed",
			Results: results,
			Summary: BulkOperationSummary{
				Total:     2,
				Succeeded: 1,
				Failed:    1,
			},
		}

		if len(response.Results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(response.Results))
		}
		if response.Summary.Total != 2 {
			t.Errorf("Expected Total 2, got %d", response.Summary.Total)
		}
		if response.Summary.Succeeded != 1 {
			t.Errorf("Expected Succeeded 1, got %d", response.Summary.Succeeded)
		}
	})

	t.Run("JSON serialization", func(t *testing.T) {
		response := &StandardResponse{
			Message: "Test message",
			Data:    map[string]interface{}{"key": "value"},
			Meta: &ResponseMeta{
				RequestID: "test-123",
				Version:   "1.0.0",
				Timestamp: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			},
		}

		jsonData, err := json.Marshal(response)
		if err != nil {
			t.Fatalf("Failed to marshal response: %v", err)
		}

		var unmarshaled StandardResponse
		if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if unmarshaled.Message != response.Message {
			t.Error("Message mismatch after JSON round-trip")
		}
		if unmarshaled.Meta.RequestID != response.Meta.RequestID {
			t.Error("RequestID mismatch after JSON round-trip")
		}
	})
}

// TestAuthResponses tests authentication response types
func TestAuthResponses(t *testing.T) {
	t.Run("LoginResponse creation", func(t *testing.T) {
		now := time.Now()
		response := &LoginResponse{
			AccessToken:  "jwt-token-123",
			RefreshToken: "refresh-token-456",
			TokenType:    "Bearer",
			ExpiresIn:    3600,
			User: UserInfo{
				ID:       "user-123",
				Username: "testuser",
				Email:    "test@example.com",
				Roles:    []string{"admin", "user"},
				Enabled:  true,
				Created:  now,
				Updated:  now,
			},
			IssuedAt: now,
		}

		if response.AccessToken != "jwt-token-123" {
			t.Errorf("Expected AccessToken 'jwt-token-123', got '%s'", response.AccessToken)
		}
		if response.User.Username != "testuser" {
			t.Errorf("Expected Username 'testuser', got '%s'", response.User.Username)
		}
		if len(response.User.Roles) != 2 {
			t.Errorf("Expected 2 roles, got %d", len(response.User.Roles))
		}
	})

	t.Run("UserListResponse creation", func(t *testing.T) {
		now := time.Now()
		users := []UserInfo{
			{ID: "1", Username: "user1", Email: "user1@example.com", Enabled: true, Created: now, Updated: now},
			{ID: "2", Username: "user2", Email: "user2@example.com", Enabled: false, Created: now, Updated: now},
		}

		response := &UserListResponse{
			Users:       users,
			Total:       2,
			Active:      1,
			Inactive:    1,
			LastUpdated: now,
		}

		if len(response.Users) != 2 {
			t.Errorf("Expected 2 users, got %d", len(response.Users))
		}
		if response.Total != 2 {
			t.Errorf("Expected Total 2, got %d", response.Total)
		}
		if response.Active != 1 {
			t.Errorf("Expected Active 1, got %d", response.Active)
		}
	})

	t.Run("AuthStatsResponse creation", func(t *testing.T) {
		now := time.Now()
		response := &AuthStatsResponse{
			TotalUsers:          10,
			ActiveUsers:         8,
			InactiveUsers:       2,
			TotalSessions:       15,
			ActiveSessions:      5,
			FailedLogins24h:     2,
			SuccessfulLogins24h: 25,
			LastLogin:           &now,
			LastUpdated:         now,
		}

		if response.TotalUsers != 10 {
			t.Errorf("Expected TotalUsers 10, got %d", response.TotalUsers)
		}
		if response.ActiveSessions != 5 {
			t.Errorf("Expected ActiveSessions 5, got %d", response.ActiveSessions)
		}
		if response.FailedLogins24h != 2 {
			t.Errorf("Expected FailedLogins24h 2, got %d", response.FailedLogins24h)
		}
	})
}

// TestHealthResponse tests health response types
func TestHealthResponse(t *testing.T) {
	t.Run("HealthResponse creation", func(t *testing.T) {
		now := time.Now()
		checks := map[string]HealthCheck{
			"database": {
				Status:    "pass",
				Message:   "Database connection successful",
				Timestamp: now,
				Duration:  "5ms",
			},
			"redis": {
				Status:    "fail",
				Message:   "Redis connection failed",
				Timestamp: now,
				Duration:  "1000ms",
			},
		}

		response := &HealthResponse{
			Status:    "degraded",
			Version:   "1.0.0",
			Uptime:    88200, // 24h30m in seconds (24*3600 + 30*60)
			Timestamp: now,
			Checks:    checks,
		}

		if response.Status != "degraded" {
			t.Errorf("Expected Status 'degraded', got '%s'", response.Status)
		}
		if len(response.Checks) != 2 {
			t.Errorf("Expected 2 checks, got %d", len(response.Checks))
		}
		if response.Checks["database"].Status != "pass" {
			t.Error("Database check should pass")
		}
	})

	t.Run("HealthCheck creation", func(t *testing.T) {
		now := time.Now()
		check := &HealthCheck{
			Status:    "pass",
			Message:   "Service is healthy",
			Timestamp: now,
			Duration:  "10ms",
		}

		if check.Status != "pass" {
			t.Errorf("Expected Status 'pass', got '%s'", check.Status)
		}
		if check.Message == "" {
			t.Error("Message should not be empty")
		}
		if check.Duration != "10ms" {
			t.Errorf("Expected Duration '10ms', got '%s'", check.Duration)
		}
	})
}

// TestPaginationInfo tests pagination info types
func TestPaginationInfo(t *testing.T) {
	t.Run("PaginationInfo creation", func(t *testing.T) {
		pagination := &PaginationInfo{
			Page:       1,
			PageSize:   20,
			TotalPages: 5,
			TotalItems: 100,
			HasNext:    true,
			HasPrev:    false,
		}

		if pagination.Page != 1 {
			t.Errorf("Expected Page 1, got %d", pagination.Page)
		}
		if pagination.TotalItems != 100 {
			t.Errorf("Expected TotalItems 100, got %d", pagination.TotalItems)
		}
		if !pagination.HasNext {
			t.Error("Expected HasNext to be true")
		}
		if pagination.HasPrev {
			t.Error("Expected HasPrev to be false")
		}
	})

	t.Run("StandardResponse with pagination", func(t *testing.T) {
		now := time.Now()
		response := &StandardResponse{
			Data:    []interface{}{"item1", "item2", "item3"},
			Message: "Items retrieved successfully",
			Pagination: &PaginationInfo{
				Page:       2,
				PageSize:   10,
				TotalPages: 3,
				TotalItems: 25,
				HasNext:    true,
				HasPrev:    true,
			},
			Meta: &ResponseMeta{
				RequestID: "req-456",
				Version:   "1.0.0",
				Timestamp: now,
			},
		}

		if response.Pagination.Page != 2 {
			t.Errorf("Expected Page 2, got %d", response.Pagination.Page)
		}
		if !response.Pagination.HasNext {
			t.Error("Expected HasNext to be true")
		}
		if !response.Pagination.HasPrev {
			t.Error("Expected HasPrev to be true")
		}
	})
}

// TestResponseSerialization tests JSON serialization
func TestResponseSerialization(t *testing.T) {
	t.Run("StandardResponse JSON serialization", func(t *testing.T) {
		now := time.Now()
		response := &StandardResponse{
			Data:    map[string]interface{}{"result": "success", "count": 42},
			Message: "Operation completed",
			Meta: &ResponseMeta{
				RequestID: "test-789",
				Version:   "1.0.0",
				Timestamp: now,
			},
		}

		jsonData, err := json.Marshal(response)
		if err != nil {
			t.Fatalf("Failed to marshal StandardResponse: %v", err)
		}

		var unmarshaled StandardResponse
		if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
			t.Fatalf("Failed to unmarshal StandardResponse: %v", err)
		}

		if unmarshaled.Message != response.Message {
			t.Error("Message mismatch after JSON round-trip")
		}
		if unmarshaled.Meta.RequestID != response.Meta.RequestID {
			t.Error("RequestID mismatch after JSON round-trip")
		}
	})

	t.Run("OperationResponse JSON serialization", func(t *testing.T) {
		response := &OperationResponse{
			Success:     true,
			Message:     "Operation completed successfully",
			OperationID: "op-123",
		}

		jsonData, err := json.Marshal(response)
		if err != nil {
			t.Fatalf("Failed to marshal OperationResponse: %v", err)
		}

		var unmarshaled OperationResponse
		if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
			t.Fatalf("Failed to unmarshal OperationResponse: %v", err)
		}

		if unmarshaled.Success != response.Success {
			t.Error("Success mismatch after JSON round-trip")
		}
		if unmarshaled.OperationID != response.OperationID {
			t.Error("OperationID mismatch after JSON round-trip")
		}
	})
}
