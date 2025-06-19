package schemas

// GetSystemSchemas returns system monitoring and control schemas
func GetSystemSchemas() map[string]interface{} {
	return map[string]interface{}{
		"SystemInfo":        getSystemInfoSchema(),
		"CPUInfo":           getCPUInfoSchema(),
		"MemoryInfo":        getMemoryInfoSchema(),
		"TemperatureData":   getTemperatureDataSchema(),
		"FanData":           getFanDataSchema(),
		"GPUInfo":           getGPUInfoSchema(),
		"UPSInfo":           getUPSInfoSchema(),
		"NetworkInfo":       getNetworkInfoSchema(),
		"SystemResources":   getSystemResourcesSchema(),
		"FilesystemInfo":    getFilesystemInfoSchema(),
		"SystemScript":      getSystemScriptSchema(),
		"ExecuteRequest":    getExecuteRequestSchema(),
		"ExecuteResponse":   getExecuteResponseSchema(),
		"LogEntry":          getLogEntrySchema(),
		"SensorChip":        getSensorChipSchema(),
		"FanInput":          getFanInputSchema(),
		"TemperatureInput":  getTemperatureInputSchema(),
		"FanInfo":           getFanInfoSchema(),
		"SystemLogs":        getSystemLogsSchema(),
		"ParityCheckStatus": getParityCheckStatusSchema(),
		"ParityDiskInfo":    getParityDiskInfoSchema(),
		"TemperatureInfo":   getTemperatureInfoSchema(),
	}
}

func getSystemInfoSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"hostname": map[string]interface{}{
				"type":        "string",
				"description": "System hostname",
				"example":     "unraid-server",
			},
			"kernel": map[string]interface{}{
				"type":        "string",
				"description": "Kernel version",
				"example":     "5.19.17-Unraid",
			},
			"uptime": map[string]interface{}{
				"type":        "integer",
				"description": "System uptime in seconds",
				"example":     86400,
				"minimum":     0,
			},
			"load_average": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "number",
				},
				"description": "Load average (1, 5, 15 minutes)",
				"example":     []float64{0.5, 0.7, 0.8},
				"minItems":    3,
				"maxItems":    3,
			},
			"architecture": map[string]interface{}{
				"type":        "string",
				"description": "System architecture",
				"example":     "x86_64",
			},
			"last_updated": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Last update timestamp",
				"example":     "2025-06-16T14:30:00Z",
			},
		},
		"required": []string{"hostname", "kernel", "uptime", "load_average"},
	}
}

func getCPUInfoSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"usage_percent": map[string]interface{}{
				"type":        "number",
				"description": "CPU usage percentage",
				"example":     25.5,
				"minimum":     0,
				"maximum":     100,
			},
			"cores": map[string]interface{}{
				"type":        "integer",
				"description": "Number of CPU cores",
				"example":     8,
				"minimum":     1,
			},
			"threads": map[string]interface{}{
				"type":        "integer",
				"description": "Number of CPU threads",
				"example":     16,
				"minimum":     1,
			},
			"model": map[string]interface{}{
				"type":        "string",
				"description": "CPU model name",
				"example":     "Intel(R) Core(TM) i7-8700K CPU @ 3.70GHz",
			},
			"frequency": map[string]interface{}{
				"type":        "number",
				"description": "Current CPU frequency in MHz",
				"example":     3700.0,
				"minimum":     0,
			},
			"temperature": map[string]interface{}{
				"type":        "number",
				"description": "CPU temperature in Celsius",
				"example":     45.5,
			},
			"last_updated": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Last update timestamp",
				"example":     "2025-06-16T14:30:00Z",
			},
		},
		"required": []string{"usage_percent", "cores", "last_updated"},
	}
}

func getMemoryInfoSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"total": map[string]interface{}{
				"type":        "integer",
				"description": "Total memory in bytes",
				"example":     34359738368,
				"minimum":     0,
			},
			"available": map[string]interface{}{
				"type":        "integer",
				"description": "Available memory in bytes",
				"example":     17179869184,
				"minimum":     0,
			},
			"used": map[string]interface{}{
				"type":        "integer",
				"description": "Used memory in bytes",
				"example":     17179869184,
				"minimum":     0,
			},
			"usage_percent": map[string]interface{}{
				"type":        "number",
				"description": "Memory usage percentage",
				"example":     50.0,
				"minimum":     0,
				"maximum":     100,
			},
			"buffers": map[string]interface{}{
				"type":        "integer",
				"description": "Buffer memory in bytes",
				"example":     1073741824,
				"minimum":     0,
			},
			"cached": map[string]interface{}{
				"type":        "integer",
				"description": "Cached memory in bytes",
				"example":     2147483648,
				"minimum":     0,
			},
			"swap_total": map[string]interface{}{
				"type":        "integer",
				"description": "Total swap in bytes",
				"example":     4294967296,
				"minimum":     0,
			},
			"swap_used": map[string]interface{}{
				"type":        "integer",
				"description": "Used swap in bytes",
				"example":     0,
				"minimum":     0,
			},
			"last_updated": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Last update timestamp",
				"example":     "2025-06-16T14:30:00Z",
			},
		},
		"required": []string{"total", "available", "used", "usage_percent", "last_updated"},
	}
}

func getTemperatureDataSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"sensors": map[string]interface{}{
				"type":        "object",
				"description": "Temperature sensors by chip",
				"additionalProperties": map[string]interface{}{
					"$ref": "#/components/schemas/SensorChip",
				},
			},
			"last_updated": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Last update timestamp",
				"example":     "2025-06-16T14:30:00Z",
			},
		},
		"required": []string{"sensors", "last_updated"},
	}
}

func getSensorChipSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Sensor chip name",
				"example":     "coretemp-isa-0000",
			},
			"temperatures": map[string]interface{}{
				"type": "object",
				"additionalProperties": map[string]interface{}{
					"$ref": "#/components/schemas/TemperatureInput",
				},
				"description": "Temperature inputs",
			},
		},
	}
}

func getTemperatureInputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"label": map[string]interface{}{
				"type":        "string",
				"description": "Temperature sensor label",
				"example":     "Core 0",
			},
			"current": map[string]interface{}{
				"type":        "number",
				"description": "Current temperature in Celsius",
				"example":     45.0,
			},
			"high": map[string]interface{}{
				"type":        "number",
				"description": "High temperature threshold",
				"example":     100.0,
			},
			"critical": map[string]interface{}{
				"type":        "number",
				"description": "Critical temperature threshold",
				"example":     105.0,
			},
		},
		"required": []string{"label", "current"},
	}
}

func getFanDataSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"fans": map[string]interface{}{
				"type":        "object",
				"description": "Fan sensors",
				"additionalProperties": map[string]interface{}{
					"$ref": "#/components/schemas/FanInput",
				},
			},
			"last_updated": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Last update timestamp",
				"example":     "2025-06-16T14:30:00Z",
			},
		},
		"required": []string{"fans", "last_updated"},
	}
}

func getFanInputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"label": map[string]interface{}{
				"type":        "string",
				"description": "Fan label",
				"example":     "CPU Fan",
			},
			"current": map[string]interface{}{
				"type":        "number",
				"description": "Current fan speed in RPM",
				"example":     1200.0,
				"minimum":     0,
			},
			"min": map[string]interface{}{
				"type":        "number",
				"description": "Minimum fan speed",
				"example":     0.0,
			},
			"max": map[string]interface{}{
				"type":        "number",
				"description": "Maximum fan speed",
				"example":     3000.0,
			},
		},
		"required": []string{"label", "current"},
	}
}

func getUPSInfoSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"status": map[string]interface{}{
				"type":        "string",
				"description": "UPS status",
				"enum":        []string{"online", "onbatt", "lowbatt", "unknown"},
				"example":     "online",
			},
			"battery_charge": map[string]interface{}{
				"type":        "number",
				"description": "Battery charge percentage",
				"example":     95.0,
				"minimum":     0,
				"maximum":     100,
			},
			"battery_runtime": map[string]interface{}{
				"type":        "integer",
				"description": "Estimated battery runtime in minutes",
				"example":     45,
				"minimum":     0,
			},
			"load_percent": map[string]interface{}{
				"type":        "number",
				"description": "UPS load percentage",
				"example":     25.5,
				"minimum":     0,
				"maximum":     100,
			},
			"input_voltage": map[string]interface{}{
				"type":        "number",
				"description": "Input voltage",
				"example":     230.0,
			},
			"output_voltage": map[string]interface{}{
				"type":        "number",
				"description": "Output voltage",
				"example":     230.0,
			},
			"model": map[string]interface{}{
				"type":        "string",
				"description": "UPS model",
				"example":     "APC Smart-UPS 1500",
			},
			"last_updated": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Last update timestamp",
				"example":     "2025-06-16T14:30:00Z",
			},
		},
		"required": []string{"status", "last_updated"},
	}
}

func getGPUInfoSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"gpus": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id": map[string]interface{}{
							"type":        "string",
							"description": "GPU ID",
							"example":     "0",
						},
						"name": map[string]interface{}{
							"type":        "string",
							"description": "GPU name",
							"example":     "NVIDIA GeForce RTX 3080",
						},
						"utilization": map[string]interface{}{
							"type":        "number",
							"description": "GPU utilization percentage",
							"example":     15.5,
							"minimum":     0,
							"maximum":     100,
						},
						"memory_used": map[string]interface{}{
							"type":        "integer",
							"description": "Used GPU memory in bytes",
							"example":     2147483648,
							"minimum":     0,
						},
						"memory_total": map[string]interface{}{
							"type":        "integer",
							"description": "Total GPU memory in bytes",
							"example":     10737418240,
							"minimum":     0,
						},
						"temperature": map[string]interface{}{
							"type":        "number",
							"description": "GPU temperature in Celsius",
							"example":     65.0,
						},
					},
				},
			},
			"last_updated": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Last update timestamp",
				"example":     "2025-06-16T14:30:00Z",
			},
		},
		"required": []string{"gpus", "last_updated"},
	}
}

func getNetworkInfoSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"interfaces": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{
							"type":        "string",
							"description": "Interface name",
							"example":     "eth0",
						},
						"ip_address": map[string]interface{}{
							"type":        "string",
							"description": "IP address",
							"example":     "192.168.1.100",
						},
						"mac_address": map[string]interface{}{
							"type":        "string",
							"description": "MAC address",
							"example":     "00:11:22:33:44:55",
						},
						"status": map[string]interface{}{
							"type":        "string",
							"description": "Interface status",
							"enum":        []string{"up", "down"},
							"example":     "up",
						},
						"speed": map[string]interface{}{
							"type":        "string",
							"description": "Interface speed",
							"example":     "1000Mbps",
						},
					},
				},
			},
			"last_updated": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Last update timestamp",
				"example":     "2025-06-16T14:30:00Z",
			},
		},
		"required": []string{"interfaces", "last_updated"},
	}
}

func getSystemResourcesSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"cpu": map[string]interface{}{
				"$ref": "#/components/schemas/CPUInfo",
			},
			"memory": map[string]interface{}{
				"$ref": "#/components/schemas/MemoryInfo",
			},
			"load_average": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "number",
				},
				"description": "Load average (1, 5, 15 minutes)",
				"example":     []float64{0.5, 0.7, 0.8},
			},
			"processes": map[string]interface{}{
				"type":        "integer",
				"description": "Number of running processes",
				"example":     150,
				"minimum":     0,
			},
			"last_updated": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Last update timestamp",
				"example":     "2025-06-16T14:30:00Z",
			},
		},
		"required": []string{"cpu", "memory", "last_updated"},
	}
}

func getFilesystemInfoSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"filesystems": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"device": map[string]interface{}{
							"type":        "string",
							"description": "Device name",
							"example":     "/dev/sda1",
						},
						"mountpoint": map[string]interface{}{
							"type":        "string",
							"description": "Mount point",
							"example":     "/mnt/disk1",
						},
						"fstype": map[string]interface{}{
							"type":        "string",
							"description": "Filesystem type",
							"example":     "xfs",
						},
						"size": map[string]interface{}{
							"type":        "integer",
							"description": "Total size in bytes",
							"example":     1099511627776,
							"minimum":     0,
						},
						"used": map[string]interface{}{
							"type":        "integer",
							"description": "Used space in bytes",
							"example":     549755813888,
							"minimum":     0,
						},
						"available": map[string]interface{}{
							"type":        "integer",
							"description": "Available space in bytes",
							"example":     549755813888,
							"minimum":     0,
						},
						"usage_percent": map[string]interface{}{
							"type":        "number",
							"description": "Usage percentage",
							"example":     50.0,
							"minimum":     0,
							"maximum":     100,
						},
					},
				},
			},
			"last_updated": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Last update timestamp",
				"example":     "2025-06-16T14:30:00Z",
			},
		},
		"required": []string{"filesystems", "last_updated"},
	}
}

func getSystemScriptSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Script name",
				"example":     "backup_script",
			},
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Script file path",
				"example":     "/boot/config/plugins/user.scripts/scripts/backup_script/script",
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "Script description",
				"example":     "Daily backup script",
			},
			"executable": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether the script is executable",
				"example":     true,
			},
		},
		"required": []string{"name", "path"},
	}
}

func getExecuteRequestSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"command": map[string]interface{}{
				"type":        "string",
				"description": "Command to execute",
				"example":     "ls -la /mnt/user",
				"maxLength":   1000,
			},
			"timeout": map[string]interface{}{
				"type":        "integer",
				"description": "Command timeout in seconds",
				"example":     30,
				"minimum":     1,
				"maximum":     300,
				"default":     30,
			},
		},
		"required": []string{"command"},
	}
}

func getExecuteResponseSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"success": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether the command executed successfully",
				"example":     true,
			},
			"exit_code": map[string]interface{}{
				"type":        "integer",
				"description": "Command exit code",
				"example":     0,
			},
			"stdout": map[string]interface{}{
				"type":        "string",
				"description": "Command standard output",
				"example":     "total 4\ndrwxrwxrwx 1 root root 28 Jun 16 14:30 .",
			},
			"stderr": map[string]interface{}{
				"type":        "string",
				"description": "Command standard error",
				"example":     "",
			},
			"duration": map[string]interface{}{
				"type":        "number",
				"description": "Command execution duration in seconds",
				"example":     0.125,
				"minimum":     0,
			},
		},
		"required": []string{"success", "exit_code", "stdout", "stderr", "duration"},
	}
}

func getLogEntrySchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"timestamp": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Log entry timestamp",
				"example":     "2025-06-16T14:30:00Z",
			},
			"level": map[string]interface{}{
				"type":        "string",
				"description": "Log level",
				"enum":        []string{"debug", "info", "warn", "error", "fatal"},
				"example":     "info",
			},
			"message": map[string]interface{}{
				"type":        "string",
				"description": "Log message",
				"example":     "System startup completed",
			},
			"source": map[string]interface{}{
				"type":        "string",
				"description": "Log source/component",
				"example":     "kernel",
			},
			"facility": map[string]interface{}{
				"type":        "string",
				"description": "Syslog facility",
				"example":     "daemon",
			},
		},
		"required": []string{"timestamp", "level", "message"},
	}
}

func getFanInfoSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"fans": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{
							"type":        "string",
							"description": "Fan name/label",
							"example":     "CPU Fan",
						},
						"speed": map[string]interface{}{
							"type":        "number",
							"description": "Current fan speed in RPM",
							"example":     1200.0,
							"minimum":     0,
						},
						"min_speed": map[string]interface{}{
							"type":        "number",
							"description": "Minimum fan speed",
							"example":     0.0,
						},
						"max_speed": map[string]interface{}{
							"type":        "number",
							"description": "Maximum fan speed",
							"example":     3000.0,
						},
						"status": map[string]interface{}{
							"type":        "string",
							"description": "Fan status",
							"enum":        []string{"normal", "warning", "critical", "unknown"},
							"example":     "normal",
						},
					},
					"required": []string{"name", "speed", "status"},
				},
			},
			"last_updated": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Last update timestamp",
				"example":     "2024-01-01T12:00:00Z",
			},
		},
		"required": []string{"fans", "last_updated"},
	}
}

func getSystemLogsSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"logs": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"$ref": "#/components/schemas/LogEntry",
				},
				"description": "System log entries",
			},
			"total_count": map[string]interface{}{
				"type":        "integer",
				"description": "Total number of log entries",
				"example":     1000,
				"minimum":     0,
			},
			"filtered_count": map[string]interface{}{
				"type":        "integer",
				"description": "Number of filtered log entries",
				"example":     50,
				"minimum":     0,
			},
			"log_sources": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
				"description": "Available log sources",
				"example":     []string{"kernel", "syslog", "auth", "daemon"},
			},
			"last_updated": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Last update timestamp",
				"example":     "2024-01-01T12:00:00Z",
			},
		},
		"required": []string{"logs", "total_count", "last_updated"},
	}
}

func getParityCheckStatusSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"status": map[string]interface{}{
				"type":        "string",
				"description": "Parity check status",
				"enum":        []string{"idle", "running", "paused", "cancelled", "completed"},
				"example":     "idle",
			},
			"progress": map[string]interface{}{
				"type":        "number",
				"description": "Check progress percentage",
				"example":     0.0,
				"minimum":     0,
				"maximum":     100,
			},
			"speed": map[string]interface{}{
				"type":        "integer",
				"description": "Check speed in bytes per second",
				"example":     150000000,
				"minimum":     0,
			},
			"eta": map[string]interface{}{
				"type":        "integer",
				"description": "Estimated time to completion in seconds",
				"example":     0,
				"minimum":     0,
			},
			"errors": map[string]interface{}{
				"type":        "integer",
				"description": "Number of errors found",
				"example":     0,
				"minimum":     0,
			},
			"last_check": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Last parity check timestamp",
				"example":     "2025-06-01T02:00:00Z",
			},
			"duration": map[string]interface{}{
				"type":        "integer",
				"description": "Last check duration in seconds",
				"example":     28800,
				"minimum":     0,
			},
			"type": map[string]interface{}{
				"type":        "string",
				"description": "Type of parity operation",
				"enum":        []string{"check", "correct"},
				"example":     "check",
			},
		},
		"required": []string{"status", "progress", "errors"},
	}
}

func getParityDiskInfoSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"parity1": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"device": map[string]interface{}{
						"type":        "string",
						"description": "Parity disk device path",
						"example":     "/dev/sdb",
					},
					"serial": map[string]interface{}{
						"type":        "string",
						"description": "Disk serial number",
						"example":     "WD-WCC4N7XXXXXX",
					},
					"model": map[string]interface{}{
						"type":        "string",
						"description": "Disk model",
						"example":     "WDC WD80EFAX-68LHPN0",
					},
					"size": map[string]interface{}{
						"type":        "integer",
						"description": "Disk size in bytes",
						"example":     8000000000000,
						"minimum":     0,
					},
					"temperature": map[string]interface{}{
						"type":        "number",
						"description": "Disk temperature in Celsius",
						"example":     35.0,
					},
					"status": map[string]interface{}{
						"type":        "string",
						"description": "Disk status",
						"enum":        []string{"active", "standby", "spun_down", "error", "missing"},
						"example":     "active",
					},
				},
				"required": []string{"device", "size", "status"},
			},
			"parity2": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"device": map[string]interface{}{
						"type":        "string",
						"description": "Parity disk device path",
						"example":     "/dev/sdc",
					},
					"serial": map[string]interface{}{
						"type":        "string",
						"description": "Disk serial number",
						"example":     "WD-WCC4N7YYYYYY",
					},
					"model": map[string]interface{}{
						"type":        "string",
						"description": "Disk model",
						"example":     "WDC WD80EFAX-68LHPN0",
					},
					"size": map[string]interface{}{
						"type":        "integer",
						"description": "Disk size in bytes",
						"example":     8000000000000,
						"minimum":     0,
					},
					"temperature": map[string]interface{}{
						"type":        "number",
						"description": "Disk temperature in Celsius",
						"example":     36.0,
					},
					"status": map[string]interface{}{
						"type":        "string",
						"description": "Disk status",
						"enum":        []string{"active", "standby", "spun_down", "error", "missing"},
						"example":     "active",
					},
				},
				"required": []string{"device", "size", "status"},
				"nullable": true,
			},
			"last_updated": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Last update timestamp",
				"example":     "2024-01-01T12:00:00Z",
			},
		},
		"required": []string{"parity1", "last_updated"},
	}
}

func getTemperatureInfoSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"sensors": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{
							"type":        "string",
							"description": "Temperature sensor name",
							"example":     "CPU Core 0",
						},
						"current": map[string]interface{}{
							"type":        "number",
							"description": "Current temperature in Celsius",
							"example":     45.0,
						},
						"high": map[string]interface{}{
							"type":        "number",
							"description": "High temperature threshold",
							"example":     80.0,
						},
						"critical": map[string]interface{}{
							"type":        "number",
							"description": "Critical temperature threshold",
							"example":     100.0,
						},
						"status": map[string]interface{}{
							"type":        "string",
							"description": "Temperature status",
							"enum":        []string{"normal", "warm", "hot", "critical"},
							"example":     "normal",
						},
						"chip": map[string]interface{}{
							"type":        "string",
							"description": "Sensor chip identifier",
							"example":     "coretemp-isa-0000",
						},
					},
					"required": []string{"name", "current", "status"},
				},
			},
			"overall_status": map[string]interface{}{
				"type":        "string",
				"description": "Overall temperature status",
				"enum":        []string{"normal", "warm", "hot", "critical"},
				"example":     "normal",
			},
			"max_temperature": map[string]interface{}{
				"type":        "number",
				"description": "Highest temperature reading",
				"example":     45.0,
			},
			"last_updated": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Last update timestamp",
				"example":     "2024-01-01T12:00:00Z",
			},
		},
		"required": []string{"sensors", "overall_status", "last_updated"},
	}
}
