package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/domalab/uma/daemon/domain"
)

// TestNewAuthService tests the creation of a new auth service
func TestNewAuthService(t *testing.T) {
	config := domain.AuthConfig{
		Enabled:   true,
		JWTSecret: "test-secret",
	}

	service := NewAuthService(config)

	if service == nil {
		t.Fatal("Expected non-nil auth service")
	}

	if !service.IsEnabled() {
		t.Error("Expected auth service to be enabled")
	}

	if string(service.jwtSecret) != "test-secret" {
		t.Error("Expected JWT secret to be set correctly")
	}

	// Should have created default admin user
	if len(service.users) != 1 {
		t.Errorf("Expected 1 default user, got %d", len(service.users))
	}

	if len(service.apiKeys) != 1 {
		t.Errorf("Expected 1 API key, got %d", len(service.apiKeys))
	}
}

// TestNewAuthServiceDisabled tests auth service when disabled
func TestNewAuthServiceDisabled(t *testing.T) {
	config := domain.AuthConfig{
		Enabled: false,
	}

	service := NewAuthService(config)

	if service.IsEnabled() {
		t.Error("Expected auth service to be disabled")
	}

	// Should not create default admin when disabled
	if len(service.users) != 0 {
		t.Errorf("Expected 0 users when disabled, got %d", len(service.users))
	}
}

// TestNewAuthServiceWithLegacyAPIKey tests auth service with legacy API key
func TestNewAuthServiceWithLegacyAPIKey(t *testing.T) {
	config := domain.AuthConfig{
		Enabled: true,
		APIKey:  "legacy-key-123",
	}

	service := NewAuthService(config)

	// Should not create default admin when legacy API key exists
	if len(service.users) != 0 {
		t.Errorf("Expected 0 users with legacy API key, got %d", len(service.users))
	}
}

// TestGenerateAPIKey tests API key generation
func TestGenerateAPIKey(t *testing.T) {
	service := NewAuthService(domain.AuthConfig{Enabled: true})

	apiKey1, err := service.GenerateAPIKey()
	if err != nil {
		t.Fatalf("Failed to generate API key: %v", err)
	}

	if apiKey1 == "" {
		t.Error("Expected non-empty API key")
	}

	if len(apiKey1) != 64 { // 32 bytes hex encoded
		t.Errorf("Expected API key length 64, got %d", len(apiKey1))
	}

	// Generate another key to ensure uniqueness
	apiKey2, err := service.GenerateAPIKey()
	if err != nil {
		t.Fatalf("Failed to generate second API key: %v", err)
	}

	if apiKey1 == apiKey2 {
		t.Error("Expected unique API keys")
	}
}

// TestValidateAPIKey tests API key validation
func TestValidateAPIKey(t *testing.T) {
	config := domain.AuthConfig{
		Enabled: true,
		APIKey:  "legacy-key-123",
	}
	service := NewAuthService(config)

	// Test legacy API key
	user, err := service.ValidateAPIKey("legacy-key-123")
	if err != nil {
		t.Fatalf("Failed to validate legacy API key: %v", err)
	}

	if user == nil {
		t.Fatal("Expected user for valid legacy API key")
	}

	if user.Username != "legacy-admin" {
		t.Errorf("Expected legacy-admin username, got %s", user.Username)
	}

	if user.Role != RoleAdmin {
		t.Errorf("Expected admin role, got %s", user.Role)
	}

	// Test invalid API key
	_, err = service.ValidateAPIKey("invalid-key")
	if err == nil {
		t.Error("Expected error for invalid API key")
	}

	// Test empty API key
	_, err = service.ValidateAPIKey("")
	if err == nil {
		t.Error("Expected error for empty API key")
	}
}

// TestValidateAPIKeyWithUserKeys tests API key validation with user-based keys
func TestValidateAPIKeyWithUserKeys(t *testing.T) {
	service := NewAuthService(domain.AuthConfig{Enabled: true})

	// Create a test user
	user, err := service.CreateUser("testuser", RoleOperator)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Test valid user API key
	validatedUser, err := service.ValidateAPIKey(user.APIKey)
	if err != nil {
		t.Fatalf("Failed to validate user API key: %v", err)
	}

	if validatedUser.ID != user.ID {
		t.Errorf("Expected user ID %s, got %s", user.ID, validatedUser.ID)
	}

	// Test that last used timestamp is updated
	originalLastUsed := validatedUser.LastUsed
	time.Sleep(10 * time.Millisecond)

	validatedUser2, err := service.ValidateAPIKey(user.APIKey)
	if err != nil {
		t.Fatalf("Failed to validate user API key second time: %v", err)
	}

	if !validatedUser2.LastUsed.After(originalLastUsed) {
		t.Error("Expected last used timestamp to be updated")
	}

	// Test disabled user
	service.UpdateUser(user.ID, map[string]interface{}{"active": false})
	_, err = service.ValidateAPIKey(user.APIKey)
	if err == nil {
		t.Error("Expected error for disabled user API key")
	}
}

// TestValidateAPIKeyDisabled tests API key validation when auth is disabled
func TestValidateAPIKeyDisabled(t *testing.T) {
	service := NewAuthService(domain.AuthConfig{Enabled: false})

	user, err := service.ValidateAPIKey("any-key")
	if err != nil {
		t.Errorf("Expected no error when auth disabled, got: %v", err)
	}

	if user != nil {
		t.Error("Expected nil user when auth disabled")
	}
}

// TestGenerateJWT tests JWT token generation
func TestGenerateJWT(t *testing.T) {
	service := NewAuthService(domain.AuthConfig{
		Enabled:   true,
		JWTSecret: "test-secret-key",
	})

	user := &User{
		ID:       "test-user-1",
		Username: "testuser",
		Role:     RoleOperator,
		Active:   true,
	}

	token, err := service.GenerateJWT(user)
	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}

	if token == "" {
		t.Error("Expected non-empty JWT token")
	}

	// Verify token structure
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Errorf("Expected JWT with 3 parts, got %d", len(parts))
	}
}

// TestGenerateJWTDisabled tests JWT generation when auth is disabled
func TestGenerateJWTDisabled(t *testing.T) {
	service := NewAuthService(domain.AuthConfig{Enabled: false})

	user := &User{
		ID:       "test-user-1",
		Username: "testuser",
		Role:     RoleOperator,
	}

	_, err := service.GenerateJWT(user)
	if err == nil {
		t.Error("Expected error when generating JWT with auth disabled")
	}
}

// TestValidateJWT tests JWT token validation
func TestValidateJWT(t *testing.T) {
	service := NewAuthService(domain.AuthConfig{
		Enabled:   true,
		JWTSecret: "test-secret-key",
	})

	user := &User{
		ID:       "test-user-1",
		Username: "testuser",
		Role:     RoleOperator,
		Active:   true,
	}

	// Generate a valid token
	token, err := service.GenerateJWT(user)
	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}

	// Validate the token
	claims, err := service.ValidateJWT(token)
	if err != nil {
		t.Fatalf("Failed to validate JWT: %v", err)
	}

	if claims == nil {
		t.Fatal("Expected non-nil claims")
	}

	if claims.UserID != user.ID {
		t.Errorf("Expected user ID %s, got %s", user.ID, claims.UserID)
	}

	if claims.Username != user.Username {
		t.Errorf("Expected username %s, got %s", user.Username, claims.Username)
	}

	if claims.Role != user.Role {
		t.Errorf("Expected role %s, got %s", user.Role, claims.Role)
	}

	// Test invalid token
	_, err = service.ValidateJWT("invalid.token.here")
	if err == nil {
		t.Error("Expected error for invalid JWT")
	}

	// Test empty token
	_, err = service.ValidateJWT("")
	if err == nil {
		t.Error("Expected error for empty JWT")
	}
}

// TestValidateJWTDisabled tests JWT validation when auth is disabled
func TestValidateJWTDisabled(t *testing.T) {
	service := NewAuthService(domain.AuthConfig{Enabled: false})

	claims, err := service.ValidateJWT("any.token.here")
	if err != nil {
		t.Errorf("Expected no error when auth disabled, got: %v", err)
	}

	if claims != nil {
		t.Error("Expected nil claims when auth disabled")
	}
}

// TestValidateJWTWithWrongSecret tests JWT validation with wrong secret
func TestValidateJWTWithWrongSecret(t *testing.T) {
	// Create service with one secret
	service1 := NewAuthService(domain.AuthConfig{
		Enabled:   true,
		JWTSecret: "secret-1",
	})

	user := &User{
		ID:       "test-user-1",
		Username: "testuser",
		Role:     RoleOperator,
	}

	// Generate token with first service
	token, err := service1.GenerateJWT(user)
	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}

	// Create service with different secret
	service2 := NewAuthService(domain.AuthConfig{
		Enabled:   true,
		JWTSecret: "secret-2",
	})

	// Try to validate with second service (wrong secret)
	_, err = service2.ValidateJWT(token)
	if err == nil {
		t.Error("Expected error when validating JWT with wrong secret")
	}
}

// TestHasPermission tests permission checking
func TestHasPermission(t *testing.T) {
	service := NewAuthService(domain.AuthConfig{Enabled: true})

	// Test admin permissions
	adminUser := &User{Role: RoleAdmin}
	if !service.HasPermission(adminUser, "any.action") {
		t.Error("Admin should have all permissions")
	}

	// Test operator permissions
	operatorUser := &User{Role: RoleOperator}
	if !service.HasPermission(operatorUser, "read.data") {
		t.Error("Operator should have read permissions")
	}

	if !service.HasPermission(operatorUser, "write.data") {
		t.Error("Operator should have write permissions")
	}

	if service.HasPermission(operatorUser, "user.create") {
		t.Error("Operator should not have user management permissions")
	}

	// Test viewer permissions
	viewerUser := &User{Role: RoleViewer}
	if !service.HasPermission(viewerUser, "read.data") {
		t.Error("Viewer should have read permissions")
	}

	if service.HasPermission(viewerUser, "write.data") {
		t.Error("Viewer should not have write permissions")
	}

	if service.HasPermission(viewerUser, "user.create") {
		t.Error("Viewer should not have user management permissions")
	}

	// Test unknown role
	unknownUser := &User{Role: Role("unknown")}
	if service.HasPermission(unknownUser, "read.data") {
		t.Error("Unknown role should have no permissions")
	}
}

// TestHasPermissionDisabled tests permission checking when auth is disabled
func TestHasPermissionDisabled(t *testing.T) {
	service := NewAuthService(domain.AuthConfig{Enabled: false})

	// When auth is disabled, all permissions should be granted
	if !service.HasPermission(nil, "any.action") {
		t.Error("Should allow all permissions when auth disabled")
	}

	user := &User{Role: RoleViewer}
	if !service.HasPermission(user, "admin.action") {
		t.Error("Should allow all permissions when auth disabled")
	}
}

// TestCreateUser tests user creation
func TestCreateUser(t *testing.T) {
	service := NewAuthService(domain.AuthConfig{Enabled: true})

	user, err := service.CreateUser("testuser", RoleOperator)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	if user == nil {
		t.Fatal("Expected non-nil user")
	}

	if user.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", user.Username)
	}

	if user.Role != RoleOperator {
		t.Errorf("Expected role 'operator', got '%s'", user.Role)
	}

	if !user.Active {
		t.Error("Expected user to be active")
	}

	if user.APIKey == "" {
		t.Error("Expected non-empty API key")
	}

	if !strings.HasPrefix(user.APIKey, "uma_") {
		t.Error("Expected API key to have 'uma_' prefix")
	}

	// Test duplicate username
	_, err = service.CreateUser("testuser", RoleViewer)
	if err == nil {
		t.Error("Expected error for duplicate username")
	}
}

// TestCreateUserDisabled tests user creation when auth is disabled
func TestCreateUserDisabled(t *testing.T) {
	service := NewAuthService(domain.AuthConfig{Enabled: false})

	_, err := service.CreateUser("testuser", RoleOperator)
	if err == nil {
		t.Error("Expected error when creating user with auth disabled")
	}
}

// TestGetUsers tests user listing
func TestGetUsers(t *testing.T) {
	service := NewAuthService(domain.AuthConfig{Enabled: true})

	// Should have default admin user
	users, err := service.GetUsers()
	if err != nil {
		t.Fatalf("Failed to get users: %v", err)
	}

	if len(users) != 1 {
		t.Logf("Users found: %d", len(users))
		for i, user := range users {
			t.Logf("User %d: %s (%s)", i, user.Username, user.Role)
		}
		t.Errorf("Expected 1 user initially, got %d", len(users))
	}

	// Create additional users
	user1, err := service.CreateUser("user1", RoleOperator)
	if err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}
	t.Logf("Created user1: %s", user1.Username)

	user2, err := service.CreateUser("user2", RoleViewer)
	if err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}
	t.Logf("Created user2: %s", user2.Username)

	users, err = service.GetUsers()
	if err != nil {
		t.Fatalf("Failed to get users: %v", err)
	}

	if len(users) != 3 {
		t.Logf("Users found after creation: %d", len(users))
		for i, user := range users {
			t.Logf("User %d: %s (%s)", i, user.Username, user.Role)
		}
		t.Errorf("Expected 3 users, got %d", len(users))
	}

	// Verify API keys are not included in response
	for _, user := range users {
		if user.APIKey != "" {
			t.Error("API key should not be included in user list response")
		}
	}
}

// TestGetUser tests individual user retrieval
func TestGetUser(t *testing.T) {
	service := NewAuthService(domain.AuthConfig{Enabled: true})

	// Create a test user
	createdUser, err := service.CreateUser("testuser", RoleOperator)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Get the user
	retrievedUser, err := service.GetUser(createdUser.ID)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}

	if retrievedUser.ID != createdUser.ID {
		t.Errorf("Expected user ID %s, got %s", createdUser.ID, retrievedUser.ID)
	}

	if retrievedUser.Username != createdUser.Username {
		t.Errorf("Expected username %s, got %s", createdUser.Username, retrievedUser.Username)
	}

	// Verify API key is not included
	if retrievedUser.APIKey != "" {
		t.Error("API key should not be included in user response")
	}

	// Test non-existent user
	_, err = service.GetUser("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent user")
	}
}

// TestUpdateUser tests user updates
func TestUpdateUser(t *testing.T) {
	service := NewAuthService(domain.AuthConfig{Enabled: true})

	// Create a test user
	user, err := service.CreateUser("testuser", RoleOperator)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Update user role
	err = service.UpdateUser(user.ID, map[string]interface{}{
		"role": string(RoleViewer),
	})
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	// Verify update
	updatedUser, err := service.GetUser(user.ID)
	if err != nil {
		t.Fatalf("Failed to get updated user: %v", err)
	}

	if updatedUser.Role != RoleViewer {
		t.Errorf("Expected role %s, got %s", RoleViewer, updatedUser.Role)
	}

	// Update user active status
	err = service.UpdateUser(user.ID, map[string]interface{}{
		"active": false,
	})
	if err != nil {
		t.Fatalf("Failed to update user active status: %v", err)
	}

	// Test non-existent user
	err = service.UpdateUser("non-existent", map[string]interface{}{
		"role": string(RoleAdmin),
	})
	if err == nil {
		t.Error("Expected error for non-existent user")
	}
}

// TestDeleteUser tests user deletion
func TestDeleteUser(t *testing.T) {
	service := NewAuthService(domain.AuthConfig{Enabled: true})

	// Create a test user
	user, err := service.CreateUser("testuser", RoleOperator)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Delete the user
	err = service.DeleteUser(user.ID)
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	// Verify user is deleted
	_, err = service.GetUser(user.ID)
	if err == nil {
		t.Error("Expected error when getting deleted user")
	}

	// Verify API key is also removed
	_, err = service.ValidateAPIKey(user.APIKey)
	if err == nil {
		t.Error("Expected error when validating deleted user's API key")
	}

	// Test deleting non-existent user
	err = service.DeleteUser("non-existent")
	if err == nil {
		t.Error("Expected error when deleting non-existent user")
	}
}

// TestRegenerateAPIKey tests API key regeneration
func TestRegenerateAPIKey(t *testing.T) {
	service := NewAuthService(domain.AuthConfig{Enabled: true})

	// Create a test user
	user, err := service.CreateUser("testuser", RoleOperator)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	originalAPIKey := user.APIKey

	// Regenerate API key
	newAPIKey, err := service.RegenerateAPIKey(user.ID)
	if err != nil {
		t.Fatalf("Failed to regenerate API key: %v", err)
	}

	if newAPIKey == originalAPIKey {
		t.Error("Expected new API key to be different from original")
	}

	if !strings.HasPrefix(newAPIKey, "uma_") {
		t.Error("Expected new API key to have 'uma_' prefix")
	}

	// Verify old API key no longer works
	_, err = service.ValidateAPIKey(originalAPIKey)
	if err == nil {
		t.Error("Expected error when validating old API key")
	}

	// Verify new API key works
	validatedUser, err := service.ValidateAPIKey(newAPIKey)
	if err != nil {
		t.Fatalf("Failed to validate new API key: %v", err)
	}

	if validatedUser.ID != user.ID {
		t.Errorf("Expected user ID %s, got %s", user.ID, validatedUser.ID)
	}

	// Test regenerating for non-existent user
	_, err = service.RegenerateAPIKey("non-existent")
	if err == nil {
		t.Error("Expected error when regenerating API key for non-existent user")
	}
}

// TestAuthMiddleware tests the authentication middleware
func TestAuthMiddleware(t *testing.T) {
	service := NewAuthService(domain.AuthConfig{
		Enabled:   true,
		JWTSecret: "test-secret",
	})

	// Create a test user
	user, err := service.CreateUser("testuser", RoleOperator)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if user is in context
		contextUser := r.Context().Value(UserContextKey)
		if contextUser == nil {
			t.Error("Expected user in request context")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if authUser, ok := contextUser.(*User); ok {
			if authUser.Username != "testuser" {
				t.Errorf("Expected username 'testuser', got '%s'", authUser.Username)
			}
		} else {
			t.Error("Expected User type in context")
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	middleware := service.AuthMiddleware(testHandler)

	// Test with valid API key
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", user.APIKey)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Test with valid JWT token
	token, err := service.GenerateJWT(user)
	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}

	req = httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 with JWT, got %d", w.Code)
	}

	// Test with invalid API key
	req = httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-API-Key", "invalid-key")
	w = httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 with invalid key, got %d", w.Code)
	}

	// Test with no authentication
	req = httptest.NewRequest("GET", "/test", nil)
	w = httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 with no auth, got %d", w.Code)
	}
}

// TestAuthMiddlewareDisabled tests middleware when auth is disabled
func TestAuthMiddlewareDisabled(t *testing.T) {
	service := NewAuthService(domain.AuthConfig{Enabled: false})

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	middleware := service.AuthMiddleware(testHandler)

	// Should allow access without authentication when disabled
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 when auth disabled, got %d", w.Code)
	}
}
