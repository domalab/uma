package schemas

// GetScriptsSchemas returns all scripts-related schemas
func GetScriptsSchemas() map[string]interface{} {
	return map[string]interface{}{
		"ScriptInfo":   getScriptInfoSchema(),
		"ScriptCreate": getScriptCreateSchema(),
	}
}

func getScriptInfoSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "Unique script identifier",
				"example":     "backup-script-001",
			},
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Script name",
				"example":     "Daily Backup Script",
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "Script description",
				"example":     "Performs daily backup of user data",
			},
			"category": map[string]interface{}{
				"type":        "string",
				"description": "Script category",
				"enum":        []string{"user", "system", "maintenance", "backup", "monitoring"},
				"example":     "backup",
			},
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Full path to script file",
				"example":     "/boot/config/plugins/user.scripts/scripts/backup-script-001/script",
			},
			"enabled": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether script is enabled",
				"example":     true,
			},
			"executable": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether script file is executable",
				"example":     true,
			},
			"size": map[string]interface{}{
				"type":        "integer",
				"description": "Script file size in bytes",
				"example":     2048,
				"minimum":     0,
			},
			"created_at": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Script creation timestamp",
				"example":     "2025-06-16T14:30:00Z",
			},
			"modified_at": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Script last modification timestamp",
				"example":     "2025-06-20T10:15:00Z",
			},
			"last_run": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Last execution timestamp",
				"example":     "2025-06-20T02:00:00Z",
			},
			"run_count": map[string]interface{}{
				"type":        "integer",
				"description": "Number of times script has been executed",
				"example":     42,
				"minimum":     0,
			},
			"schedule": map[string]interface{}{
				"type":        "string",
				"description": "Cron schedule expression (if scheduled)",
				"example":     "0 2 * * *",
			},
			"timeout": map[string]interface{}{
				"type":        "integer",
				"description": "Script timeout in seconds",
				"example":     3600,
				"minimum":     1,
			},
			"permissions": map[string]interface{}{
				"type":        "string",
				"description": "File permissions in octal format",
				"example":     "755",
			},
		},
		"required": []string{"id", "name", "path", "enabled", "executable"},
	}
}

func getScriptCreateSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Script name",
				"example":     "My Custom Script",
				"minLength":   1,
				"maxLength":   100,
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "Script description",
				"example":     "Custom script for system maintenance",
				"maxLength":   500,
			},
			"category": map[string]interface{}{
				"type":        "string",
				"description": "Script category",
				"enum":        []string{"user", "system", "maintenance", "backup", "monitoring"},
				"example":     "maintenance",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "Script content (for text-based creation)",
				"example":     "#!/bin/bash\necho \"Hello World\"",
			},
			"schedule": map[string]interface{}{
				"type":        "string",
				"description": "Cron schedule expression (optional)",
				"example":     "0 2 * * *",
			},
			"timeout": map[string]interface{}{
				"type":        "integer",
				"description": "Script timeout in seconds",
				"example":     3600,
				"minimum":     1,
				"maximum":     86400,
				"default":     3600,
			},
			"enabled": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether script should be enabled",
				"example":     true,
				"default":     true,
			},
		},
		"required": []string{"name"},
	}
}
