package requests

import (
	"encoding/json"
	"testing"
)

// TestAuthRequests tests authentication request types
func TestAuthRequests(t *testing.T) {
	t.Run("LoginRequest creation", func(t *testing.T) {
		request := LoginRequest{
			Username: "testuser",
			Password: "testpass123",
		}

		if request.Username != "testuser" {
			t.Errorf("Expected Username 'testuser', got '%s'", request.Username)
		}
		if request.Password != "testpass123" {
			t.Errorf("Expected Password 'testpass123', got '%s'", request.Password)
		}
	})

	t.Run("LoginRequest fields", func(t *testing.T) {
		tests := []struct {
			name    string
			request LoginRequest
			hasUser bool
			hasPass bool
		}{
			{
				name: "Complete request",
				request: LoginRequest{
					Username: "testuser",
					Password: "testpass123",
				},
				hasUser: true,
				hasPass: true,
			},
			{
				name: "Empty username",
				request: LoginRequest{
					Password: "testpass123",
				},
				hasUser: false,
				hasPass: true,
			},
			{
				name: "Empty password",
				request: LoginRequest{
					Username: "testuser",
				},
				hasUser: true,
				hasPass: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				hasUser := tt.request.Username != ""
				hasPass := tt.request.Password != ""

				if hasUser != tt.hasUser {
					t.Errorf("Expected hasUser %v, got %v", tt.hasUser, hasUser)
				}
				if hasPass != tt.hasPass {
					t.Errorf("Expected hasPass %v, got %v", tt.hasPass, hasPass)
				}
			})
		}
	})

	t.Run("LoginRequest JSON serialization", func(t *testing.T) {
		request := LoginRequest{
			Username: "jsonuser",
			Password: "jsonpass",
		}

		jsonData, err := json.Marshal(request)
		if err != nil {
			t.Fatalf("Failed to marshal login request: %v", err)
		}

		var unmarshaled LoginRequest
		if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
			t.Fatalf("Failed to unmarshal login request: %v", err)
		}

		if unmarshaled.Username != request.Username {
			t.Errorf("Username mismatch after JSON round-trip")
		}
		if unmarshaled.Password != request.Password {
			t.Errorf("Password mismatch after JSON round-trip")
		}
	})

	t.Run("PasswordChangeRequest creation", func(t *testing.T) {
		request := PasswordChangeRequest{
			CurrentPassword: "oldpass",
			NewPassword:     "newpass123",
		}

		if request.CurrentPassword != "oldpass" {
			t.Errorf("Expected CurrentPassword 'oldpass', got '%s'", request.CurrentPassword)
		}
		if request.NewPassword != "newpass123" {
			t.Errorf("Expected NewPassword 'newpass123', got '%s'", request.NewPassword)
		}
	})

	t.Run("UserCreateRequest creation", func(t *testing.T) {
		request := UserCreateRequest{
			Username:    "newuser",
			Password:    "password123",
			Email:       "newuser@example.com",
			FullName:    "New User",
			Roles:       []string{"user"},
			Permissions: []string{"read"},
			Enabled:     true,
		}

		if request.Username != "newuser" {
			t.Errorf("Expected Username 'newuser', got '%s'", request.Username)
		}
		if len(request.Roles) != 1 {
			t.Errorf("Expected 1 role, got %d", len(request.Roles))
		}
		if !request.Enabled {
			t.Error("Expected Enabled to be true")
		}
	})
}

// TestDockerRequests tests Docker request types
func TestDockerRequests(t *testing.T) {
	t.Run("DockerContainerCreateRequest creation", func(t *testing.T) {
		request := DockerContainerCreateRequest{
			Name:          "test-container",
			Image:         "nginx:latest",
			RestartPolicy: "always",
			Ports:         map[string]string{"80": "80", "443": "443"},
			Environment:   map[string]string{"ENV": "production"},
			Command:       []string{"nginx", "-g", "daemon off;"},
			Networks:      []string{"bridge"},
			Privileged:    false,
			AutoRemove:    false,
		}

		if request.Name != "test-container" {
			t.Errorf("Expected Name 'test-container', got '%s'", request.Name)
		}
		if request.Image != "nginx:latest" {
			t.Errorf("Expected Image 'nginx:latest', got '%s'", request.Image)
		}
		if len(request.Ports) != 2 {
			t.Errorf("Expected 2 ports, got %d", len(request.Ports))
		}
		if len(request.Environment) != 1 {
			t.Errorf("Expected 1 environment variable, got %d", len(request.Environment))
		}
	})

	t.Run("DockerContainerCreateRequest fields", func(t *testing.T) {
		tests := []struct {
			name     string
			request  DockerContainerCreateRequest
			hasName  bool
			hasImage bool
		}{
			{
				name: "Complete request",
				request: DockerContainerCreateRequest{
					Name:  "test-container",
					Image: "nginx:latest",
				},
				hasName:  true,
				hasImage: true,
			},
			{
				name: "Empty name",
				request: DockerContainerCreateRequest{
					Image: "nginx:latest",
				},
				hasName:  false,
				hasImage: true,
			},
			{
				name: "Empty image",
				request: DockerContainerCreateRequest{
					Name: "test-container",
				},
				hasName:  true,
				hasImage: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				hasName := tt.request.Name != ""
				hasImage := tt.request.Image != ""

				if hasName != tt.hasName {
					t.Errorf("Expected hasName %v, got %v", tt.hasName, hasName)
				}
				if hasImage != tt.hasImage {
					t.Errorf("Expected hasImage %v, got %v", tt.hasImage, hasImage)
				}
			})
		}
	})

	t.Run("DockerContainerActionRequest creation", func(t *testing.T) {
		request := DockerContainerActionRequest{
			Action: "start",
			Force:  false,
		}

		if request.Action != "start" {
			t.Errorf("Expected Action 'start', got '%s'", request.Action)
		}
		if request.Force {
			t.Error("Expected Force to be false")
		}
	})

	t.Run("DockerContainerActionRequest actions", func(t *testing.T) {
		validActions := []string{"start", "stop", "restart", "pause", "unpause", "remove"}

		for _, action := range validActions {
			request := DockerContainerActionRequest{
				Action: action,
				Force:  true,
			}

			if request.Action != action {
				t.Errorf("Expected Action '%s', got '%s'", action, request.Action)
			}
			if !request.Force {
				t.Error("Expected Force to be true")
			}
		}
	})

	t.Run("DockerBulkActionRequest creation", func(t *testing.T) {
		request := DockerBulkActionRequest{
			ContainerIDs: []string{"container1", "container2"},
			Action:       "start",
			Force:        false,
		}

		if len(request.ContainerIDs) != 2 {
			t.Errorf("Expected 2 container IDs, got %d", len(request.ContainerIDs))
		}
		if request.Action != "start" {
			t.Errorf("Expected Action 'start', got '%s'", request.Action)
		}
		if request.Force {
			t.Error("Expected Force to be false")
		}
	})
}

// TestSystemRequests tests system request types
func TestSystemRequests(t *testing.T) {
	t.Run("CommandExecuteRequest creation", func(t *testing.T) {
		request := CommandExecuteRequest{
			Command:          "ls -la",
			Timeout:          30,
			WorkingDirectory: "/tmp",
		}

		if request.Command != "ls -la" {
			t.Errorf("Expected Command 'ls -la', got '%s'", request.Command)
		}
		if request.Timeout != 30 {
			t.Errorf("Expected Timeout 30, got %d", request.Timeout)
		}
		if request.WorkingDirectory != "/tmp" {
			t.Errorf("Expected WorkingDirectory '/tmp', got '%s'", request.WorkingDirectory)
		}
	})

	t.Run("SystemShutdownRequest creation", func(t *testing.T) {
		request := SystemShutdownRequest{
			DelaySeconds: 60,
			Message:      "System will shutdown in 1 minute",
			Force:        false,
		}

		if request.DelaySeconds != 60 {
			t.Errorf("Expected DelaySeconds 60, got %d", request.DelaySeconds)
		}
		if request.Message == "" {
			t.Error("Message should not be empty")
		}
		if request.Force {
			t.Error("Expected Force to be false")
		}
	})

	t.Run("SystemRebootRequest creation", func(t *testing.T) {
		request := SystemRebootRequest{
			DelaySeconds: 120,
			Message:      "System will reboot in 2 minutes",
			Force:        true,
		}

		if request.DelaySeconds != 120 {
			t.Errorf("Expected DelaySeconds 120, got %d", request.DelaySeconds)
		}
		if !request.Force {
			t.Error("Expected Force to be true")
		}
	})
}

// TestCommonRequests tests common request types
func TestCommonRequests(t *testing.T) {
	t.Run("PaginationRequest creation", func(t *testing.T) {
		request := PaginationRequest{
			Page:     1,
			PageSize: 20,
			Offset:   0,
			Limit:    100,
		}

		if request.Page != 1 {
			t.Errorf("Expected Page 1, got %d", request.Page)
		}
		if request.PageSize != 20 {
			t.Errorf("Expected PageSize 20, got %d", request.PageSize)
		}
		if request.Limit != 100 {
			t.Errorf("Expected Limit 100, got %d", request.Limit)
		}
	})

	t.Run("FilterRequest creation", func(t *testing.T) {
		request := FilterRequest{
			Search: "test",
			Filters: map[string]string{
				"status": "running",
				"name":   "test",
			},
			SortBy:   "created",
			SortDesc: true,
		}

		if request.Search != "test" {
			t.Errorf("Expected Search 'test', got '%s'", request.Search)
		}
		if len(request.Filters) != 2 {
			t.Errorf("Expected 2 filters, got %d", len(request.Filters))
		}
		if !request.SortDesc {
			t.Error("Expected SortDesc to be true")
		}
	})

	t.Run("BulkOperationRequest creation", func(t *testing.T) {
		request := BulkOperationRequest{
			IDs:       []string{"id1", "id2", "id3"},
			Operation: "delete",
			Force:     true,
		}

		if len(request.IDs) != 3 {
			t.Errorf("Expected 3 IDs, got %d", len(request.IDs))
		}
		if request.Operation != "delete" {
			t.Errorf("Expected Operation 'delete', got '%s'", request.Operation)
		}
		if !request.Force {
			t.Error("Expected Force to be true")
		}
	})
}
