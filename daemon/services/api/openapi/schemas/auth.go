package schemas

// GetAuthSchemas returns authentication-related schemas
func GetAuthSchemas() map[string]interface{} {
	return map[string]interface{}{
		"LoginRequest":   getLoginRequestSchema(),
		"LoginResponse":  getLoginResponseSchema(),
		"TokenResponse":  getTokenResponseSchema(),
		"RefreshRequest": getRefreshRequestSchema(),
		"UserInfo":       getUserInfoSchema(),
		"APIKeyInfo":     getAPIKeyInfoSchema(),
		"AuthError":      getAuthErrorSchema(),
		"AuthStats":      getAuthStatsSchema(),
		"AuthUser":       getAuthUserSchema(),
		"SessionInfo":    getSessionInfoSchema(),
		"PermissionInfo": getPermissionInfoSchema(),
	}
}

func getLoginRequestSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"username": map[string]interface{}{
				"type":        "string",
				"description": "Username for authentication",
				"example":     "admin",
				"minLength":   1,
				"maxLength":   50,
				"pattern":     "^[a-zA-Z0-9_.-]+$",
			},
			"password": map[string]interface{}{
				"type":        "string",
				"description": "Password for authentication",
				"example":     "secure_password",
				"minLength":   1,
				"maxLength":   100,
				"format":      "password",
			},
			"remember_me": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether to create a long-lived session",
				"example":     false,
				"default":     false,
			},
			"client_info": map[string]interface{}{
				"type":        "object",
				"description": "Optional client information",
				"properties": map[string]interface{}{
					"user_agent": map[string]interface{}{
						"type":        "string",
						"description": "Client user agent",
						"example":     "UMA-Client/1.0",
					},
					"ip_address": map[string]interface{}{
						"type":        "string",
						"description": "Client IP address",
						"example":     "192.168.1.100",
					},
				},
			},
		},
		"required": []string{"username", "password"},
	}
}

func getLoginResponseSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"success": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether login was successful",
				"example":     true,
			},
			"token": map[string]interface{}{
				"type":        "string",
				"description": "JWT access token",
				"example":     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
			},
			"refresh_token": map[string]interface{}{
				"type":        "string",
				"description": "JWT refresh token",
				"example":     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
			},
			"expires_in": map[string]interface{}{
				"type":        "integer",
				"description": "Token expiration time in seconds",
				"example":     3600,
				"minimum":     1,
			},
			"token_type": map[string]interface{}{
				"type":        "string",
				"description": "Token type",
				"example":     "Bearer",
				"default":     "Bearer",
			},
			"user": map[string]interface{}{
				"$ref": "#/components/schemas/UserInfo",
			},
			"permissions": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
				"description": "User permissions",
				"example":     []string{"read", "write", "admin"},
			},
			"session_id": map[string]interface{}{
				"type":        "string",
				"description": "Session identifier",
				"example":     "sess_1234567890",
			},
		},
		"required": []string{"success", "token", "expires_in", "token_type", "user"},
	}
}

func getTokenResponseSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"access_token": map[string]interface{}{
				"type":        "string",
				"description": "JWT access token",
				"example":     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
			},
			"refresh_token": map[string]interface{}{
				"type":        "string",
				"description": "JWT refresh token",
				"example":     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
			},
			"token_type": map[string]interface{}{
				"type":        "string",
				"description": "Token type",
				"example":     "Bearer",
				"default":     "Bearer",
			},
			"expires_in": map[string]interface{}{
				"type":        "integer",
				"description": "Token expiration time in seconds",
				"example":     3600,
				"minimum":     1,
			},
			"scope": map[string]interface{}{
				"type":        "string",
				"description": "Token scope",
				"example":     "read write admin",
			},
			"issued_at": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Token issuance timestamp",
				"example":     "2025-06-16T14:30:00Z",
			},
		},
		"required": []string{"access_token", "token_type", "expires_in"},
	}
}

func getRefreshRequestSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"refresh_token": map[string]interface{}{
				"type":        "string",
				"description": "JWT refresh token",
				"example":     "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
			},
		},
		"required": []string{"refresh_token"},
	}
}

func getUserInfoSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "User ID",
				"example":     "user_1234567890",
			},
			"username": map[string]interface{}{
				"type":        "string",
				"description": "Username",
				"example":     "admin",
			},
			"email": map[string]interface{}{
				"type":        "string",
				"format":      "email",
				"description": "User email address",
				"example":     "admin@example.com",
			},
			"full_name": map[string]interface{}{
				"type":        "string",
				"description": "User full name",
				"example":     "System Administrator",
			},
			"role": map[string]interface{}{
				"type":        "string",
				"description": "User role",
				"enum":        []string{"admin", "user", "readonly"},
				"example":     "admin",
			},
			"permissions": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
				"description": "User permissions",
				"example":     []string{"read", "write", "admin"},
			},
			"created_at": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "User creation timestamp",
				"example":     "2025-06-16T14:30:00Z",
			},
			"last_login": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Last login timestamp",
				"example":     "2025-06-16T14:30:00Z",
			},
			"active": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether user account is active",
				"example":     true,
			},
		},
		"required": []string{"id", "username", "role", "permissions", "active"},
	}
}

func getAPIKeyInfoSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "API key ID",
				"example":     "key_1234567890",
			},
			"name": map[string]interface{}{
				"type":        "string",
				"description": "API key name",
				"example":     "Home Assistant Integration",
			},
			"key": map[string]interface{}{
				"type":        "string",
				"description": "API key value (only shown on creation)",
				"example":     "uma_1234567890abcdef",
			},
			"permissions": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
				"description": "API key permissions",
				"example":     []string{"read", "write"},
			},
			"created_at": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "API key creation timestamp",
				"example":     "2025-06-16T14:30:00Z",
			},
			"last_used": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Last usage timestamp",
				"example":     "2025-06-16T14:30:00Z",
			},
			"expires_at": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "API key expiration timestamp",
				"example":     "2026-06-16T14:30:00Z",
			},
			"active": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether API key is active",
				"example":     true,
			},
		},
		"required": []string{"id", "name", "permissions", "created_at", "active"},
	}
}

func getAuthErrorSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"error": map[string]interface{}{
				"type":        "string",
				"description": "Authentication error message",
				"example":     "Invalid credentials",
			},
			"error_code": map[string]interface{}{
				"type":        "string",
				"description": "Authentication error code",
				"enum": []string{
					"INVALID_CREDENTIALS", "TOKEN_EXPIRED", "TOKEN_INVALID",
					"INSUFFICIENT_PERMISSIONS", "ACCOUNT_DISABLED", "RATE_LIMITED",
				},
				"example": "INVALID_CREDENTIALS",
			},
			"details": map[string]interface{}{
				"type":                 "object",
				"description":          "Additional error details",
				"additionalProperties": true,
				"example": map[string]interface{}{
					"attempts_remaining": 2,
					"lockout_duration":   300,
				},
			},
			"timestamp": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Error timestamp",
				"example":     "2025-06-16T14:30:00Z",
			},
		},
		"required": []string{"error", "error_code", "timestamp"},
	}
}

func getAuthStatsSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"total_users": map[string]interface{}{
				"type":        "integer",
				"description": "Total number of users",
				"example":     5,
				"minimum":     0,
			},
			"active_users": map[string]interface{}{
				"type":        "integer",
				"description": "Number of active users",
				"example":     3,
				"minimum":     0,
			},
			"active_sessions": map[string]interface{}{
				"type":        "integer",
				"description": "Number of active sessions",
				"example":     2,
				"minimum":     0,
			},
			"api_keys": map[string]interface{}{
				"type":        "integer",
				"description": "Number of active API keys",
				"example":     4,
				"minimum":     0,
			},
			"failed_logins_24h": map[string]interface{}{
				"type":        "integer",
				"description": "Failed login attempts in last 24 hours",
				"example":     1,
				"minimum":     0,
			},
			"successful_logins_24h": map[string]interface{}{
				"type":        "integer",
				"description": "Successful logins in last 24 hours",
				"example":     15,
				"minimum":     0,
			},
			"last_updated": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Last update timestamp",
				"example":     "2025-06-16T14:30:00Z",
			},
		},
		"required": []string{"total_users", "active_users", "active_sessions", "api_keys", "last_updated"},
	}
}

func getAuthUserSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "User ID",
				"example":     "user_1234567890",
			},
			"username": map[string]interface{}{
				"type":        "string",
				"description": "Username",
				"example":     "admin",
			},
			"email": map[string]interface{}{
				"type":        "string",
				"format":      "email",
				"description": "User email address",
				"example":     "admin@example.com",
			},
			"full_name": map[string]interface{}{
				"type":        "string",
				"description": "User full name",
				"example":     "System Administrator",
			},
			"role": map[string]interface{}{
				"type":        "string",
				"description": "User role",
				"enum":        []string{"admin", "user", "readonly"},
				"example":     "admin",
			},
			"permissions": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
				"description": "User permissions",
				"example":     []string{"read", "write", "admin"},
			},
			"created_at": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "User creation timestamp",
				"example":     "2025-06-16T14:30:00Z",
			},
			"last_login": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Last login timestamp",
				"example":     "2025-06-16T14:30:00Z",
			},
			"active": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether user account is active",
				"example":     true,
			},
			"login_count": map[string]interface{}{
				"type":        "integer",
				"description": "Total number of logins",
				"example":     42,
				"minimum":     0,
			},
			"last_ip": map[string]interface{}{
				"type":        "string",
				"description": "Last login IP address",
				"example":     "192.168.1.100",
			},
		},
		"required": []string{"id", "username", "role", "permissions", "active"},
	}
}

func getSessionInfoSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "Session ID",
				"example":     "sess_1234567890",
			},
			"user_id": map[string]interface{}{
				"type":        "string",
				"description": "User ID",
				"example":     "user_1234567890",
			},
			"username": map[string]interface{}{
				"type":        "string",
				"description": "Username",
				"example":     "admin",
			},
			"ip_address": map[string]interface{}{
				"type":        "string",
				"description": "Client IP address",
				"example":     "192.168.1.100",
			},
			"user_agent": map[string]interface{}{
				"type":        "string",
				"description": "Client user agent",
				"example":     "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
			},
			"created_at": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Session creation timestamp",
				"example":     "2025-06-16T14:30:00Z",
			},
			"expires_at": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Session expiration timestamp",
				"example":     "2025-06-16T18:30:00Z",
			},
			"last_activity": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Last activity timestamp",
				"example":     "2025-06-16T14:30:00Z",
			},
			"active": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether session is active",
				"example":     true,
			},
		},
		"required": []string{"id", "user_id", "username", "created_at", "expires_at", "active"},
	}
}

func getPermissionInfoSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Permission name",
				"example":     "docker.containers.manage",
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "Permission description",
				"example":     "Manage Docker containers (start, stop, restart)",
			},
			"category": map[string]interface{}{
				"type":        "string",
				"description": "Permission category",
				"enum":        []string{"system", "docker", "storage", "vm", "auth", "monitoring"},
				"example":     "docker",
			},
			"level": map[string]interface{}{
				"type":        "string",
				"description": "Permission level",
				"enum":        []string{"read", "write", "admin"},
				"example":     "write",
			},
			"resource": map[string]interface{}{
				"type":        "string",
				"description": "Resource this permission applies to",
				"example":     "containers",
			},
		},
		"required": []string{"name", "description", "category", "level"},
	}
}
