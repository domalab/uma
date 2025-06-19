package models

import (
	"encoding/json"
	"testing"
	"time"
)

// TestNotificationModel tests the Notification model
func TestNotificationModel(t *testing.T) {
	t.Run("Create notification", func(t *testing.T) {
		notification := &Notification{
			ID:       "test-123",
			Type:     "info",
			Title:    "Test Notification",
			Message:  "This is a test notification",
			Created:  time.Now(),
			Updated:  time.Now(),
			Read:     false,
			Priority: 3,
		}

		if notification.ID != "test-123" {
			t.Errorf("Expected ID 'test-123', got '%s'", notification.ID)
		}
		if notification.Type != "info" {
			t.Errorf("Expected Type 'info', got '%s'", notification.Type)
		}
		if notification.Read {
			t.Error("Expected Read to be false")
		}
		if notification.Priority != 3 {
			t.Errorf("Expected Priority 3, got %d", notification.Priority)
		}
	})

	t.Run("JSON serialization", func(t *testing.T) {
		notification := &Notification{
			ID:       "json-test",
			Type:     "warning",
			Title:    "JSON Test",
			Message:  "Testing JSON serialization",
			Created:  time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			Updated:  time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			Read:     true,
			Priority: 2,
		}

		jsonData, err := json.Marshal(notification)
		if err != nil {
			t.Fatalf("Failed to marshal notification: %v", err)
		}

		var unmarshaled Notification
		if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
			t.Fatalf("Failed to unmarshal notification: %v", err)
		}

		if unmarshaled.ID != notification.ID {
			t.Errorf("ID mismatch after JSON round-trip: expected '%s', got '%s'", notification.ID, unmarshaled.ID)
		}
		if unmarshaled.Type != notification.Type {
			t.Errorf("Type mismatch after JSON round-trip: expected '%s', got '%s'", notification.Type, unmarshaled.Type)
		}
		if unmarshaled.Read != notification.Read {
			t.Errorf("Read mismatch after JSON round-trip: expected %v, got %v", notification.Read, unmarshaled.Read)
		}
	})

	t.Run("Notification fields", func(t *testing.T) {
		notification := &Notification{
			ID:       "field-test",
			Type:     "info",
			Title:    "Field Test",
			Message:  "Testing notification fields",
			Source:   "test-service",
			Category: "system",
			Priority: 3,
			Read:     false,
			Created:  time.Now(),
			Updated:  time.Now(),
		}

		// Test field access
		if notification.ID == "" {
			t.Error("ID should not be empty")
		}
		if notification.Type == "" {
			t.Error("Type should not be empty")
		}
		if notification.Priority < 1 || notification.Priority > 5 {
			t.Errorf("Priority should be between 1-5, got %d", notification.Priority)
		}
		if notification.Category == "" {
			t.Error("Category should not be empty")
		}
	})
}

// TestScriptModel tests the Script model
func TestScriptModel(t *testing.T) {
	t.Run("Create script", func(t *testing.T) {
		script := &Script{
			Name:        "Test Script",
			Description: "A test script",
			Path:        "/usr/local/emhttp/plugins/test/script.sh",
			Category:    "custom",
			Executable:  true,
			Size:        1024,
			Modified:    time.Now(),
			Permissions: "755",
			Owner:       "root",
			Group:       "root",
			Tags:        []string{"test", "automation"},
		}

		if script.Name != "Test Script" {
			t.Errorf("Expected Name 'Test Script', got '%s'", script.Name)
		}
		if !script.Executable {
			t.Error("Expected Executable to be true")
		}
		if len(script.Tags) != 2 {
			t.Errorf("Expected 2 tags, got %d", len(script.Tags))
		}
		if script.Permissions != "755" {
			t.Errorf("Expected Permissions '755', got '%s'", script.Permissions)
		}
	})

	t.Run("JSON serialization", func(t *testing.T) {
		script := &Script{
			Name:        "JSON Script",
			Description: "Testing JSON",
			Path:        "/test/script.sh",
			Category:    "system",
			Executable:  false,
			Size:        512,
			Modified:    time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			Permissions: "644",
			Owner:       "user",
			Group:       "users",
		}

		jsonData, err := json.Marshal(script)
		if err != nil {
			t.Fatalf("Failed to marshal script: %v", err)
		}

		var unmarshaled Script
		if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
			t.Fatalf("Failed to unmarshal script: %v", err)
		}

		if unmarshaled.Name != script.Name {
			t.Errorf("Name mismatch after JSON round-trip")
		}
		if unmarshaled.Size != script.Size {
			t.Errorf("Size mismatch after JSON round-trip")
		}
		if unmarshaled.Permissions != script.Permissions {
			t.Errorf("Permissions mismatch after JSON round-trip")
		}
	})

	t.Run("Script fields", func(t *testing.T) {
		script := &Script{
			Name:        "Field Test Script",
			Path:        "/test/field-test.sh",
			Category:    "maintenance",
			Executable:  true,
			Size:        2048,
			Permissions: "755",
			Owner:       "root",
			Group:       "root",
			Hash:        "sha256:abc123",
			Tags:        []string{"maintenance", "backup"},
			Metadata:    map[string]string{"version": "1.0", "author": "admin"},
		}

		// Test field access
		if script.Name == "" {
			t.Error("Name should not be empty")
		}
		if script.Path == "" {
			t.Error("Path should not be empty")
		}
		if script.Size <= 0 {
			t.Error("Size should be positive")
		}
		if len(script.Tags) != 2 {
			t.Errorf("Expected 2 tags, got %d", len(script.Tags))
		}
		if len(script.Metadata) != 2 {
			t.Errorf("Expected 2 metadata entries, got %d", len(script.Metadata))
		}
	})
}

// TestShareModel tests the Share model
func TestShareModel(t *testing.T) {
	t.Run("Create share", func(t *testing.T) {
		share := &Share{
			Name:             "TestShare",
			Comment:          "A test share",
			Path:             "/mnt/user/TestShare",
			AllocatorMethod:  "high-water",
			MinimumFreeSpace: "1GB",
			SplitLevel:       1,
			IncludedDisks:    []string{"disk1", "disk2"},
			ExcludedDisks:    []string{"disk3"},
			UseCache:         "yes",
			CachePool:        "cache",
			SMBEnabled:       true,
			SMBSecurity:      "secure",
			NFSEnabled:       false,
			AFPEnabled:       false,
			FTPEnabled:       false,
			Created:          time.Now(),
			Modified:         time.Now(),
		}

		if share.Name != "TestShare" {
			t.Errorf("Expected Name 'TestShare', got '%s'", share.Name)
		}
		if !share.SMBEnabled {
			t.Error("Expected SMBEnabled to be true")
		}
		if share.NFSEnabled {
			t.Error("Expected NFSEnabled to be false")
		}
		if len(share.IncludedDisks) != 2 {
			t.Errorf("Expected 2 included disks, got %d", len(share.IncludedDisks))
		}
	})

	t.Run("JSON serialization", func(t *testing.T) {
		share := &Share{
			Name:             "JSONShare",
			Comment:          "JSON test share",
			Path:             "/mnt/user/JSONShare",
			AllocatorMethod:  "most-free",
			MinimumFreeSpace: "500MB",
			SplitLevel:       2,
			UseCache:         "no",
			SMBEnabled:       false,
			SMBSecurity:      "public",
			Created:          time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			Modified:         time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		}

		jsonData, err := json.Marshal(share)
		if err != nil {
			t.Fatalf("Failed to marshal share: %v", err)
		}

		var unmarshaled Share
		if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
			t.Fatalf("Failed to unmarshal share: %v", err)
		}

		if unmarshaled.Name != share.Name {
			t.Errorf("Name mismatch after JSON round-trip")
		}
		if unmarshaled.SMBEnabled != share.SMBEnabled {
			t.Errorf("SMBEnabled mismatch after JSON round-trip")
		}
		if unmarshaled.AllocatorMethod != share.AllocatorMethod {
			t.Errorf("AllocatorMethod mismatch after JSON round-trip")
		}
	})

	t.Run("Share fields", func(t *testing.T) {
		share := &Share{
			Name:             "FieldTestShare",
			Comment:          "Testing share fields",
			Path:             "/mnt/user/FieldTestShare",
			AllocatorMethod:  "fill-up",
			MinimumFreeSpace: "2GB",
			SplitLevel:       3,
			IncludedDisks:    []string{"disk1", "disk2", "disk3"},
			ExcludedDisks:    []string{},
			UseCache:         "prefer",
			CachePool:        "cache2",
			SMBEnabled:       true,
			SMBSecurity:      "private",
			NFSEnabled:       true,
			AFPEnabled:       false,
			FTPEnabled:       true,
		}

		// Test field access
		if share.Name == "" {
			t.Error("Name should not be empty")
		}
		if share.Path == "" {
			t.Error("Path should not be empty")
		}
		if share.SplitLevel < 0 {
			t.Error("SplitLevel should not be negative")
		}
		if len(share.IncludedDisks) != 3 {
			t.Errorf("Expected 3 included disks, got %d", len(share.IncludedDisks))
		}

		// Test valid allocator methods
		validMethods := []string{"high-water", "most-free", "fill-up"}
		isValidMethod := false
		for _, method := range validMethods {
			if share.AllocatorMethod == method {
				isValidMethod = true
				break
			}
		}
		if !isValidMethod {
			t.Errorf("Invalid allocator method: %s", share.AllocatorMethod)
		}
	})
}
