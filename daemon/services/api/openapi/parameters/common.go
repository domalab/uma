package parameters

// GetCommonParameters returns all common reusable parameters
func GetCommonParameters() map[string]interface{} {
	return map[string]interface{}{
		"PageParameter":          GetPageParameter(),
		"LimitParameter":         GetLimitParameter(),
		"RequestIDParameter":     GetRequestIDParameter(),
		"ContainerIDParameter":   GetContainerIDParameter(),
		"VMIDParameter":          GetVMIDParameter(),
		"DiskIDParameter":        GetDiskIDParameter(),
		"AllContainersParameter": GetAllContainersParameter(),
		"ForceParameter":         GetForceParameter(),
		"TimeoutParameter":       GetTimeoutParameter(),
		"VerboseParameter":       GetVerboseParameter(),
	}
}

// Pagination parameters
func GetPaginationParameters() []interface{} {
	return []interface{}{
		GetPageParameter(),
		GetLimitParameter(),
	}
}

func GetPageParameter() map[string]interface{} {
	return map[string]interface{}{
		"name":        "page",
		"in":          "query",
		"description": "Page number for pagination (1-based)",
		"required":    false,
		"schema": map[string]interface{}{
			"type":    "integer",
			"minimum": 1,
			"default": 1,
			"example": 1,
		},
	}
}

func GetLimitParameter() map[string]interface{} {
	return map[string]interface{}{
		"name":        "limit",
		"in":          "query",
		"description": "Number of items per page",
		"required":    false,
		"schema": map[string]interface{}{
			"type":    "integer",
			"minimum": 1,
			"maximum": 1000,
			"default": 50,
			"example": 50,
		},
	}
}

// Request tracking parameters
func GetRequestIDParameter() map[string]interface{} {
	return map[string]interface{}{
		"name":        "X-Request-ID",
		"in":          "header",
		"description": "Optional request ID for tracing and debugging",
		"required":    false,
		"schema": map[string]interface{}{
			"type":    "string",
			"pattern": "^[a-zA-Z0-9-_]{1,64}$",
			"example": "req_1234567890_5678",
		},
	}
}

// Resource ID parameters
func GetContainerIDParameter() map[string]interface{} {
	return map[string]interface{}{
		"name":        "id",
		"in":          "path",
		"description": "Container ID or name",
		"required":    true,
		"schema": map[string]interface{}{
			"type":    "string",
			"pattern": "^[a-zA-Z0-9][a-zA-Z0-9_.-]+$",
			"example": "plex",
		},
	}
}

func GetVMIDParameter() map[string]interface{} {
	return map[string]interface{}{
		"name":        "id",
		"in":          "path",
		"description": "Virtual machine ID or name",
		"required":    true,
		"schema": map[string]interface{}{
			"type":    "string",
			"pattern": "^[a-zA-Z0-9][a-zA-Z0-9_.-]+$",
			"example": "Windows-10-Gaming",
		},
	}
}

func GetDiskIDParameter() map[string]interface{} {
	return map[string]interface{}{
		"name":        "id",
		"in":          "path",
		"description": "Disk identifier (e.g., disk1, parity, cache)",
		"required":    true,
		"schema": map[string]interface{}{
			"type":    "string",
			"pattern": "^(disk|parity|cache)\\d*$",
			"example": "disk1",
		},
	}
}

// Query parameters
func GetAllContainersParameter() map[string]interface{} {
	return map[string]interface{}{
		"name":        "all",
		"in":          "query",
		"description": "Include stopped containers in the response",
		"required":    false,
		"schema": map[string]interface{}{
			"type":    "boolean",
			"default": false,
			"example": false,
		},
	}
}

func GetForceParameter() map[string]interface{} {
	return map[string]interface{}{
		"name":        "force",
		"in":          "query",
		"description": "Force the operation (use with caution)",
		"required":    false,
		"schema": map[string]interface{}{
			"type":    "boolean",
			"default": false,
			"example": false,
		},
	}
}

func GetTimeoutParameter() map[string]interface{} {
	return map[string]interface{}{
		"name":        "timeout",
		"in":          "query",
		"description": "Operation timeout in seconds",
		"required":    false,
		"schema": map[string]interface{}{
			"type":    "integer",
			"minimum": 1,
			"maximum": 300,
			"default": 30,
			"example": 30,
		},
	}
}

func GetVerboseParameter() map[string]interface{} {
	return map[string]interface{}{
		"name":        "verbose",
		"in":          "query",
		"description": "Include detailed information in the response",
		"required":    false,
		"schema": map[string]interface{}{
			"type":    "boolean",
			"default": false,
			"example": false,
		},
	}
}

// Filter parameters
func GetStatusFilterParameter() map[string]interface{} {
	return map[string]interface{}{
		"name":        "status",
		"in":          "query",
		"description": "Filter by status",
		"required":    false,
		"schema": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "string",
			},
			"example": []string{"running", "stopped"},
		},
	}
}

func GetSinceParameter() map[string]interface{} {
	return map[string]interface{}{
		"name":        "since",
		"in":          "query",
		"description": "Show data since timestamp (ISO 8601 format)",
		"required":    false,
		"schema": map[string]interface{}{
			"type":    "string",
			"format":  "date-time",
			"example": "2025-06-16T14:30:00Z",
		},
	}
}

func GetUntilParameter() map[string]interface{} {
	return map[string]interface{}{
		"name":        "until",
		"in":          "query",
		"description": "Show data until timestamp (ISO 8601 format)",
		"required":    false,
		"schema": map[string]interface{}{
			"type":    "string",
			"format":  "date-time",
			"example": "2025-06-16T15:30:00Z",
		},
	}
}

// Content negotiation parameters
func GetAcceptParameter() map[string]interface{} {
	return map[string]interface{}{
		"name":        "Accept",
		"in":          "header",
		"description": "Preferred response content type",
		"required":    false,
		"schema": map[string]interface{}{
			"type":    "string",
			"enum":    []string{"application/json", "application/vnd.uma.v1+json"},
			"default": "application/json",
			"example": "application/vnd.uma.v1+json",
		},
	}
}

// System-specific parameters
func GetLogLevelParameter() map[string]interface{} {
	return map[string]interface{}{
		"name":        "level",
		"in":          "query",
		"description": "Filter logs by level",
		"required":    false,
		"schema": map[string]interface{}{
			"type": "array",
			"items": map[string]interface{}{
				"type": "string",
				"enum": []string{"debug", "info", "warn", "error", "fatal"},
			},
			"example": []string{"error", "fatal"},
		},
	}
}

func GetLogLinesParameter() map[string]interface{} {
	return map[string]interface{}{
		"name":        "lines",
		"in":          "query",
		"description": "Number of log lines to return",
		"required":    false,
		"schema": map[string]interface{}{
			"type":    "integer",
			"minimum": 1,
			"maximum": 10000,
			"default": 100,
			"example": 100,
		},
	}
}

// Storage-specific parameters
func GetSMARTParameter() map[string]interface{} {
	return map[string]interface{}{
		"name":        "smart",
		"in":          "query",
		"description": "Include SMART data in disk information",
		"required":    false,
		"schema": map[string]interface{}{
			"type":    "boolean",
			"default": false,
			"example": true,
		},
	}
}

func GetTemperatureParameter() map[string]interface{} {
	return map[string]interface{}{
		"name":        "temperature",
		"in":          "query",
		"description": "Include temperature data",
		"required":    false,
		"schema": map[string]interface{}{
			"type":    "boolean",
			"default": false,
			"example": true,
		},
	}
}
