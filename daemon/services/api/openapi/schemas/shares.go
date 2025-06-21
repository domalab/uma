package schemas

// GetSharesSchemas returns all shares-related schemas
func GetSharesSchemas() map[string]interface{} {
	return map[string]interface{}{
		"ShareInfo":   getShareInfoSchema(),
		"ShareCreate": getShareCreateSchema(),
	}
}

func getShareInfoSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Share name",
				"example":     "media",
			},
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Local path being shared",
				"example":     "/mnt/user/media",
			},
			"protocol": map[string]interface{}{
				"type":        "string",
				"description": "Sharing protocol",
				"enum":        []string{"smb", "nfs", "afp", "ftp"},
				"example":     "smb",
			},
			"enabled": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether share is enabled",
				"example":     true,
			},
			"public": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether share allows guest access",
				"example":     false,
			},
			"read_only": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether share is read-only",
				"example":     false,
			},
			"browseable": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether share is browseable",
				"example":     true,
			},
			"comment": map[string]interface{}{
				"type":        "string",
				"description": "Share description/comment",
				"example":     "Media files and entertainment content",
			},
			"users": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
				"description": "List of users with access",
				"example":     []string{"admin", "user1", "media"},
			},
			"groups": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
				"description": "List of groups with access",
				"example":     []string{"users", "media"},
			},
			"size": map[string]interface{}{
				"type":        "integer",
				"description": "Total size of shared directory in bytes",
				"example":     1073741824000,
				"minimum":     0,
			},
			"used": map[string]interface{}{
				"type":        "integer",
				"description": "Used space in bytes",
				"example":     536870912000,
				"minimum":     0,
			},
			"available": map[string]interface{}{
				"type":        "integer",
				"description": "Available space in bytes",
				"example":     536870912000,
				"minimum":     0,
			},
			"usage_percent": map[string]interface{}{
				"type":        "number",
				"format":      "float",
				"description": "Usage percentage",
				"example":     50.0,
				"minimum":     0,
				"maximum":     100,
			},
			"active_connections": map[string]interface{}{
				"type":        "integer",
				"description": "Number of active connections",
				"example":     3,
				"minimum":     0,
			},
			"created_at": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Share creation timestamp",
				"example":     "2025-06-16T14:30:00Z",
			},
			"modified_at": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Share last modification timestamp",
				"example":     "2025-06-20T10:15:00Z",
			},
		},
		"required": []string{"name", "path", "protocol", "enabled"},
	}
}

func getShareCreateSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Share name",
				"example":     "documents",
				"minLength":   1,
				"maxLength":   50,
				"pattern":     "^[a-zA-Z0-9_-]+$",
			},
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Local path to share",
				"example":     "/mnt/user/documents",
				"minLength":   1,
			},
			"protocol": map[string]interface{}{
				"type":        "string",
				"description": "Sharing protocol",
				"enum":        []string{"smb", "nfs", "afp", "ftp"},
				"example":     "smb",
			},
			"comment": map[string]interface{}{
				"type":        "string",
				"description": "Share description",
				"example":     "Document storage and collaboration",
				"maxLength":   200,
			},
			"public": map[string]interface{}{
				"type":        "boolean",
				"description": "Allow guest access",
				"example":     false,
				"default":     false,
			},
			"read_only": map[string]interface{}{
				"type":        "boolean",
				"description": "Make share read-only",
				"example":     false,
				"default":     false,
			},
			"browseable": map[string]interface{}{
				"type":        "boolean",
				"description": "Make share browseable",
				"example":     true,
				"default":     true,
			},
			"enabled": map[string]interface{}{
				"type":        "boolean",
				"description": "Enable share immediately",
				"example":     true,
				"default":     true,
			},
			"users": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
				"description": "Users with access to this share",
				"example":     []string{"admin", "user1"},
			},
			"groups": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
				"description": "Groups with access to this share",
				"example":     []string{"users"},
			},
		},
		"required": []string{"name", "path", "protocol"},
	}
}
