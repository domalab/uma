package schemas

// GetDiagnosticsSchemas returns diagnostic-related schemas
func GetDiagnosticsSchemas() map[string]interface{} {
	return map[string]interface{}{
		"DiagnosticsHealth": getDiagnosticsHealthSchema(),
		"DiagnosticsInfo":   getDiagnosticsInfoSchema(),
		"DiagnosticsRepair": getDiagnosticsRepairSchema(),
	}
}

func getDiagnosticsHealthSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"status": map[string]interface{}{
				"type":        "string",
				"description": "Overall health status",
				"enum":        []string{"healthy", "warning", "critical", "unknown"},
				"example":     "healthy",
			},
			"checks": map[string]interface{}{
				"type":        "array",
				"description": "List of individual health checks",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{
							"type":        "string",
							"description": "Name of the health check",
							"example":     "disk_health",
						},
						"status": map[string]interface{}{
							"type":        "string",
							"description": "Status of this check",
							"enum":        []string{"passed", "warning", "critical", "unknown"},
							"example":     "passed",
						},
						"message": map[string]interface{}{
							"type":        "string",
							"description": "Descriptive message about the check result",
							"example":     "All disks are healthy",
						},
						"critical": map[string]interface{}{
							"type":        "boolean",
							"description": "Whether this is a critical check",
							"example":     false,
						},
						"remediation": map[string]interface{}{
							"type":        "string",
							"description": "Suggested remediation steps",
							"example":     "No action required",
						},
						"last_updated": map[string]interface{}{
							"type":        "string",
							"format":      "date-time",
							"description": "When this check was last performed",
							"example":     "2024-01-01T12:00:00Z",
						},
					},
					"required": []string{"name", "status", "message"},
				},
			},
			"last_check": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "When the health check was last performed",
				"example":     "2024-01-01T12:00:00Z",
			},
			"message": map[string]interface{}{
				"type":        "string",
				"description": "Overall health message",
				"example":     "System is healthy",
			},
		},
		"required": []string{"status", "checks", "last_check"},
	}
}

func getDiagnosticsInfoSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"version": map[string]interface{}{
				"type":        "string",
				"description": "UMA version",
				"example":     "1.0.0",
			},
			"system": map[string]interface{}{
				"type":        "string",
				"description": "System identifier",
				"example":     "uma",
			},
			"diagnostics": map[string]interface{}{
				"type":        "string",
				"description": "Diagnostics system status",
				"example":     "enabled",
			},
			"capabilities": map[string]interface{}{
				"type":        "array",
				"description": "Available diagnostic capabilities",
				"items": map[string]interface{}{
					"type": "string",
				},
				"example": []string{"health_checks", "system_repair", "log_analysis"},
			},
			"last_run": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "When diagnostics were last run",
				"example":     "2024-01-01T12:00:00Z",
			},
			"message": map[string]interface{}{
				"type":        "string",
				"description": "Diagnostic system message",
				"example":     "Diagnostics system operational",
			},
		},
		"required": []string{"version", "system", "diagnostics"},
	}
}

func getDiagnosticsRepairSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"description": "Repair action that was performed",
				"example":     "fix_permissions",
			},
			"status": map[string]interface{}{
				"type":        "string",
				"description": "Status of the repair operation",
				"enum":        []string{"success", "failed", "partial", "in_progress"},
				"example":     "success",
			},
			"message": map[string]interface{}{
				"type":        "string",
				"description": "Result message from the repair operation",
				"example":     "Permissions fixed successfully",
			},
			"details": map[string]interface{}{
				"type":        "array",
				"description": "Detailed repair steps performed",
				"items": map[string]interface{}{
					"type": "string",
				},
				"example": []string{"Fixed /var/log permissions", "Corrected disk mount points"},
			},
			"timestamp": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "When the repair was performed",
				"example":     "2024-01-01T12:00:00Z",
			},
			"duration": map[string]interface{}{
				"type":        "number",
				"description": "Duration of repair operation in seconds",
				"example":     5.2,
			},
		},
		"required": []string{"action", "status", "message", "timestamp"},
	}
}
