package schemas

// GetStorageSchemas returns storage management schemas
func GetStorageSchemas() map[string]interface{} {
	return map[string]interface{}{
		"ArrayInfo":       getArrayInfoSchema(),
		"ArrayDisk":       getArrayDiskSchema(),
		"ParityDisk":      getParityDiskSchema(),
		"DiskInfo":        getDiskInfoSchema(),
		"SMARTData":       getSMARTDataSchema(),
		"ParityInfo":      getParityInfoSchema(),
		"ParityCheckInfo": getParityCheckInfoSchema(),
		"CacheInfo":       getCacheInfoSchema(),
		"ZFSPoolInfo":     getZFSPoolInfoSchema(),
		"ZFSDatasetInfo":  getZFSDatasetInfoSchema(),
		"ArrayOperation":  getArrayOperationSchema(),
		"ArrayStatus":     getArrayStatusSchema(),
		"DiskTemperature": getDiskTemperatureSchema(),
		"StorageOverview": getStorageOverviewSchema(),
		"BootInfo":        getBootInfoSchema(),
		"DiskList":        getDiskListSchema(),
		"StorageGeneral":  getStorageGeneralSchema(),
		"ZFSInfo":         getZFSInfoSchema(),
	}
}

func getArrayInfoSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"disks": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"$ref": "#/components/schemas/ArrayDisk",
				},
				"description": "Array data disks",
			},
			"parity": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"$ref": "#/components/schemas/ParityDisk",
				},
				"description": "Parity disks",
			},
			"protection": map[string]interface{}{
				"type":        "string",
				"description": "Array protection level",
				"enum":        []string{"parity", "dual-parity", "none"},
				"example":     "parity",
			},
			"state": map[string]interface{}{
				"type":        "string",
				"description": "Array state",
				"enum":        []string{"started", "stopped", "starting", "stopping"},
				"example":     "started",
			},
			"sync_action": map[string]interface{}{
				"type":        "string",
				"description": "Current synchronization action",
				"enum":        []string{"check", "check P", "resync", "none", "idle"},
				"example":     "check P",
			},
			"sync_progress": map[string]interface{}{
				"type":        "number",
				"description": "Synchronization progress percentage",
				"example":     45.2,
				"minimum":     0,
				"maximum":     100,
			},
			"last_updated": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Last update timestamp",
				"example":     "2025-06-19T14:30:00Z",
			},
		},
		"required": []string{"disks", "parity", "protection", "state", "last_updated"},
	}
}

func getArrayDiskSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"device": map[string]interface{}{
				"type":        "string",
				"description": "Device path",
				"example":     "/dev/sda",
			},
			"health": map[string]interface{}{
				"type":        "string",
				"description": "Disk health status",
				"enum":        []string{"PASSED", "FAILED", "UNKNOWN"},
				"example":     "PASSED",
			},
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Disk name",
				"example":     "disk1",
			},
			"serial": map[string]interface{}{
				"type":        "string",
				"description": "Disk serial number",
				"example":     "WD-WCC4N7XXXXXX",
			},
			"size": map[string]interface{}{
				"type":        "string",
				"description": "Disk size (human readable)",
				"example":     "8.0 TB",
			},
			"smart_data": map[string]interface{}{
				"type":                 "object",
				"description":          "SMART data attributes",
				"additionalProperties": true,
			},
			"status": map[string]interface{}{
				"type":        "string",
				"description": "Disk status",
				"enum":        []string{"active", "standby", "spun_down", "error", "missing"},
				"example":     "active",
			},
			"temperature": map[string]interface{}{
				"type":        "number",
				"description": "Disk temperature in Celsius",
				"example":     35.0,
			},
			"type": map[string]interface{}{
				"type":        "string",
				"description": "Disk type",
				"enum":        []string{"data", "cache", "pool"},
				"example":     "data",
			},
		},
		"required": []string{"device", "name", "status", "type"},
	}
}

func getParityDiskSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"device": map[string]interface{}{
				"type":        "string",
				"description": "Device path",
				"example":     "/dev/sdb",
			},
			"health": map[string]interface{}{
				"type":        "string",
				"description": "Disk health status",
				"enum":        []string{"PASSED", "FAILED", "UNKNOWN"},
				"example":     "PASSED",
			},
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Parity disk name",
				"example":     "parity",
			},
			"serial": map[string]interface{}{
				"type":        "string",
				"description": "Disk serial number",
				"example":     "WD-WCC4N7YYYYYY",
			},
			"size": map[string]interface{}{
				"type":        "string",
				"description": "Disk size (human readable)",
				"example":     "8.0 TB",
			},
			"smart_data": map[string]interface{}{
				"type":                 "object",
				"description":          "SMART data attributes",
				"additionalProperties": true,
			},
			"status": map[string]interface{}{
				"type":        "string",
				"description": "Disk status",
				"enum":        []string{"active", "standby", "spun_down", "error", "missing"},
				"example":     "active",
			},
			"temperature": map[string]interface{}{
				"type":        "number",
				"description": "Disk temperature in Celsius",
				"example":     34.0,
			},
			"type": map[string]interface{}{
				"type":        "string",
				"description": "Disk type",
				"enum":        []string{"parity", "parity2"},
				"example":     "parity",
			},
		},
		"required": []string{"device", "name", "status", "type"},
	}
}

func getDiskInfoSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Disk name",
				"example":     "disk1",
				"pattern":     "^(disk|parity|cache)\\d*$",
			},
			"device": map[string]interface{}{
				"type":        "string",
				"description": "Device path",
				"example":     "/dev/sda",
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
			"used": map[string]interface{}{
				"type":        "integer",
				"description": "Used space in bytes",
				"example":     4000000000000,
				"minimum":     0,
			},
			"free": map[string]interface{}{
				"type":        "integer",
				"description": "Free space in bytes",
				"example":     4000000000000,
				"minimum":     0,
			},
			"usage_percent": map[string]interface{}{
				"type":        "number",
				"description": "Usage percentage",
				"example":     50.0,
				"minimum":     0,
				"maximum":     100,
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
			"filesystem": map[string]interface{}{
				"type":        "string",
				"description": "Filesystem type",
				"example":     "xfs",
			},
			"health": map[string]interface{}{
				"type":        "string",
				"description": "Disk health status",
				"enum":        []string{"healthy", "unknown", "warning", "critical"},
				"example":     "healthy",
			},
			"type": map[string]interface{}{
				"type":        "string",
				"description": "Disk type",
				"enum":        []string{"disk", "parity", "cache", "pool"},
				"example":     "disk",
			},
			"smart_data": map[string]interface{}{
				"type":        "object",
				"description": "SMART monitoring data",
				"properties": map[string]interface{}{
					"available": map[string]interface{}{
						"type":        "boolean",
						"description": "Whether SMART data is available",
						"example":     true,
					},
					"status": map[string]interface{}{
						"type":        "string",
						"description": "SMART status",
						"enum":        []string{"passed", "failed", "unknown"},
						"example":     "passed",
					},
					"attributes": map[string]interface{}{
						"type":        "object",
						"description": "SMART attributes",
						"additionalProperties": map[string]interface{}{
							"type": "number",
						},
						"example": map[string]interface{}{
							"power_on_hours":    18762,
							"power_cycle_count": 241,
						},
					},
				},
				"required": []string{"available", "status"},
			},
			"smart": map[string]interface{}{
				"$ref": "#/components/schemas/SMARTData",
			},
			"last_updated": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Last update timestamp",
				"example":     "2025-06-16T14:30:00Z",
			},
		},
		"required": []string{"name", "device", "size", "status", "health", "type", "smart_data", "last_updated"},
	}
}

func getSMARTDataSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"overall_health": map[string]interface{}{
				"type":        "string",
				"description": "Overall SMART health status",
				"enum":        []string{"PASSED", "FAILED", "UNKNOWN"},
				"example":     "PASSED",
			},
			"temperature": map[string]interface{}{
				"type":        "number",
				"description": "Current temperature in Celsius",
				"example":     35.0,
			},
			"power_on_hours": map[string]interface{}{
				"type":        "integer",
				"description": "Total power-on hours",
				"example":     8760,
				"minimum":     0,
			},
			"power_cycle_count": map[string]interface{}{
				"type":        "integer",
				"description": "Power cycle count",
				"example":     100,
				"minimum":     0,
			},
			"reallocated_sectors": map[string]interface{}{
				"type":        "integer",
				"description": "Reallocated sector count",
				"example":     0,
				"minimum":     0,
			},
			"pending_sectors": map[string]interface{}{
				"type":        "integer",
				"description": "Current pending sector count",
				"example":     0,
				"minimum":     0,
			},
			"uncorrectable_errors": map[string]interface{}{
				"type":        "integer",
				"description": "Offline uncorrectable error count",
				"example":     0,
				"minimum":     0,
			},
			"last_test_result": map[string]interface{}{
				"type":        "string",
				"description": "Last self-test result",
				"example":     "Completed without error",
			},
			"last_updated": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Last SMART data update",
				"example":     "2025-06-16T14:30:00Z",
			},
		},
		"required": []string{"overall_health", "last_updated"},
	}
}

func getParityInfoSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"parity1": map[string]interface{}{
				"anyOf": []interface{}{
					map[string]interface{}{"$ref": "#/components/schemas/DiskInfo"},
					map[string]interface{}{"type": "null"},
				},
				"description": "First parity disk (null if not present)",
			},
			"parity2": map[string]interface{}{
				"anyOf": []interface{}{
					map[string]interface{}{"$ref": "#/components/schemas/DiskInfo"},
					map[string]interface{}{"type": "null"},
				},
				"description": "Second parity disk (null if not present)",
			},
			"check_status": map[string]interface{}{
				"$ref": "#/components/schemas/ParityCheckInfo",
			},
			"last_updated": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Last update timestamp",
				"example":     "2025-06-16T14:30:00Z",
			},
		},
		"required": []string{"last_updated"},
	}
}

func getParityCheckInfoSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"status": map[string]interface{}{
				"type":        "string",
				"description": "Parity check status",
				"enum":        []string{"idle", "running", "paused", "cancelled"},
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
			"scheduled": map[string]interface{}{
				"type":        "string",
				"description": "Next scheduled check",
				"example":     "Monthly on 1st at 02:00",
			},
		},
		"required": []string{"status", "progress", "errors"},
	}
}

func getCacheInfoSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"disks": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"$ref": "#/components/schemas/DiskInfo",
				},
				"description": "Cache disks",
			},
			"pool_status": map[string]interface{}{
				"type":        "string",
				"description": "Cache pool status",
				"enum":        []string{"online", "degraded", "offline", "faulted"},
				"example":     "online",
			},
			"total_size": map[string]interface{}{
				"type":        "integer",
				"description": "Total cache size in bytes",
				"example":     1000000000000,
				"minimum":     0,
			},
			"used": map[string]interface{}{
				"type":        "integer",
				"description": "Used cache space in bytes",
				"example":     500000000000,
				"minimum":     0,
			},
			"free": map[string]interface{}{
				"type":        "integer",
				"description": "Free cache space in bytes",
				"example":     500000000000,
				"minimum":     0,
			},
			"usage_percent": map[string]interface{}{
				"type":        "number",
				"description": "Cache usage percentage",
				"example":     50.0,
				"minimum":     0,
				"maximum":     100,
			},
			"pools": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{
							"type":        "string",
							"description": "Pool name",
							"example":     "cache",
						},
						"type": map[string]interface{}{
							"type":        "string",
							"description": "Pool type",
							"enum":        []string{"cache", "zfs_cache"},
							"example":     "cache",
						},
						"device": map[string]interface{}{
							"type":        "string",
							"description": "Pool device",
							"example":     "/dev/nvme0n1p1",
						},
						"mountpoint": map[string]interface{}{
							"type":        "string",
							"description": "Pool mountpoint",
							"example":     "/mnt/cache",
						},
						"size": map[string]interface{}{
							"type":        "string",
							"description": "Pool size",
							"example":     "477G",
						},
						"used": map[string]interface{}{
							"type":        "string",
							"description": "Used space",
							"example":     "67G",
						},
						"available": map[string]interface{}{
							"type":        "string",
							"description": "Available space",
							"example":     "408G",
						},
						"usage": map[string]interface{}{
							"type":        "string",
							"description": "Usage percentage",
							"example":     "14%",
						},
						"health": map[string]interface{}{
							"type":        "string",
							"description": "Pool health",
							"enum":        []string{"healthy", "ONLINE", "DEGRADED", "FAULTED"},
							"example":     "healthy",
						},
						"temperature": map[string]interface{}{
							"type":        "number",
							"description": "Pool temperature",
							"example":     0,
						},
						"smart_data": map[string]interface{}{
							"type":                 "object",
							"description":          "SMART data for the pool",
							"additionalProperties": true,
						},
					},
					"required": []string{"name", "type", "size", "health"},
				},
				"description": "Cache pools information",
			},
			"last_updated": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Last update timestamp",
				"example":     "2025-06-16T14:30:00Z",
			},
		},
		"required": []string{"disks", "pool_status", "pools", "last_updated"},
	}
}

func getZFSPoolInfoSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "ZFS pool name",
				"example":     "tank",
			},
			"status": map[string]interface{}{
				"type":        "string",
				"description": "Pool status",
				"enum":        []string{"ONLINE", "DEGRADED", "FAULTED", "OFFLINE", "UNAVAIL", "REMOVED"},
				"example":     "ONLINE",
			},
			"health": map[string]interface{}{
				"type":        "string",
				"description": "Pool health",
				"enum":        []string{"ONLINE", "DEGRADED", "FAULTED", "OFFLINE", "UNAVAIL", "REMOVED"},
				"example":     "ONLINE",
			},
			"size": map[string]interface{}{
				"type":        "integer",
				"description": "Total pool size in bytes",
				"example":     2000000000000,
				"minimum":     0,
			},
			"allocated": map[string]interface{}{
				"type":        "integer",
				"description": "Allocated space in bytes",
				"example":     1000000000000,
				"minimum":     0,
			},
			"free": map[string]interface{}{
				"type":        "integer",
				"description": "Free space in bytes",
				"example":     1000000000000,
				"minimum":     0,
			},
			"fragmentation": map[string]interface{}{
				"type":        "number",
				"description": "Pool fragmentation percentage",
				"example":     15.5,
				"minimum":     0,
				"maximum":     100,
			},
			"capacity": map[string]interface{}{
				"type":        "number",
				"description": "Pool capacity percentage",
				"example":     50.0,
				"minimum":     0,
				"maximum":     100,
			},
			"dedup_ratio": map[string]interface{}{
				"type":        "number",
				"description": "Deduplication ratio",
				"example":     1.0,
				"minimum":     1.0,
			},
			"last_updated": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Last update timestamp",
				"example":     "2025-06-16T14:30:00Z",
			},
		},
		"required": []string{"name", "status", "health", "size", "last_updated"},
	}
}

func getZFSDatasetInfoSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Dataset name",
				"example":     "tank/data",
			},
			"type": map[string]interface{}{
				"type":        "string",
				"description": "Dataset type",
				"enum":        []string{"filesystem", "volume", "snapshot"},
				"example":     "filesystem",
			},
			"used": map[string]interface{}{
				"type":        "integer",
				"description": "Used space in bytes",
				"example":     500000000000,
				"minimum":     0,
			},
			"available": map[string]interface{}{
				"type":        "integer",
				"description": "Available space in bytes",
				"example":     1500000000000,
				"minimum":     0,
			},
			"referenced": map[string]interface{}{
				"type":        "integer",
				"description": "Referenced space in bytes",
				"example":     500000000000,
				"minimum":     0,
			},
			"compression": map[string]interface{}{
				"type":        "string",
				"description": "Compression algorithm",
				"example":     "lz4",
			},
			"mountpoint": map[string]interface{}{
				"type":        "string",
				"description": "Dataset mountpoint",
				"example":     "/mnt/tank/data",
			},
			"last_updated": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Last update timestamp",
				"example":     "2025-06-16T14:30:00Z",
			},
		},
		"required": []string{"name", "type", "used", "available", "last_updated"},
	}
}

func getArrayOperationSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"operation": map[string]interface{}{
				"type":        "string",
				"description": "Array operation to perform",
				"enum":        []string{"start", "stop"},
				"example":     "start",
			},
			"force": map[string]interface{}{
				"type":        "boolean",
				"description": "Force the operation (use with caution)",
				"example":     false,
				"default":     false,
			},
		},
		"required": []string{"operation"},
	}
}

func getArrayStatusSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"success": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether the operation was successful",
				"example":     true,
			},
			"message": map[string]interface{}{
				"type":        "string",
				"description": "Operation result message",
				"example":     "Array started successfully",
			},
			"status": map[string]interface{}{
				"type":        "string",
				"description": "Current array status",
				"enum":        []string{"started", "stopped", "starting", "stopping"},
				"example":     "started",
			},
			"warnings": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
				"description": "Any warnings from the operation",
				"example":     []string{"Disk temperature high"},
			},
		},
		"required": []string{"success", "message", "status"},
	}
}

func getDiskTemperatureSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"disk": map[string]interface{}{
				"type":        "string",
				"description": "Disk identifier",
				"example":     "disk1",
			},
			"device": map[string]interface{}{
				"type":        "string",
				"description": "Device path",
				"example":     "/dev/sda",
			},
			"temperature": map[string]interface{}{
				"type":        "number",
				"description": "Current temperature in Celsius",
				"example":     35.0,
			},
			"max_temperature": map[string]interface{}{
				"type":        "number",
				"description": "Maximum safe temperature",
				"example":     60.0,
			},
			"status": map[string]interface{}{
				"type":        "string",
				"description": "Temperature status",
				"enum":        []string{"normal", "warm", "hot", "critical"},
				"example":     "normal",
			},
			"last_updated": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Last update timestamp",
				"example":     "2025-06-16T14:30:00Z",
			},
		},
		"required": []string{"disk", "device", "temperature", "status", "last_updated"},
	}
}

func getStorageOverviewSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"array": map[string]interface{}{
				"$ref": "#/components/schemas/ArrayInfo",
			},
			"parity": map[string]interface{}{
				"$ref": "#/components/schemas/ParityInfo",
			},
			"cache": map[string]interface{}{
				"$ref": "#/components/schemas/CacheInfo",
			},
			"disks": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"$ref": "#/components/schemas/DiskInfo",
				},
				"description": "All disks in the system",
			},
			"zfs_pools": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"$ref": "#/components/schemas/ZFSPoolInfo",
				},
				"description": "ZFS pools (if available)",
			},
			"total_capacity": map[string]interface{}{
				"type":        "integer",
				"description": "Total storage capacity in bytes",
				"example":     50000000000000,
				"minimum":     0,
			},
			"total_used": map[string]interface{}{
				"type":        "integer",
				"description": "Total used storage in bytes",
				"example":     25000000000000,
				"minimum":     0,
			},
			"total_free": map[string]interface{}{
				"type":        "integer",
				"description": "Total free storage in bytes",
				"example":     25000000000000,
				"minimum":     0,
			},
			"overall_usage_percent": map[string]interface{}{
				"type":        "number",
				"description": "Overall storage usage percentage",
				"example":     50.0,
				"minimum":     0,
				"maximum":     100,
			},
			"last_updated": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Last update timestamp",
				"example":     "2025-06-16T14:30:00Z",
			},
		},
		"required": []string{"array", "disks", "total_capacity", "total_used", "total_free", "last_updated"},
	}
}

func getBootInfoSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"device": map[string]interface{}{
				"type":        "string",
				"description": "Boot device path",
				"example":     "/dev/sdb1",
			},
			"filesystem": map[string]interface{}{
				"type":        "string",
				"description": "Boot filesystem type",
				"example":     "vfat",
			},
			"size": map[string]interface{}{
				"type":        "integer",
				"description": "Boot device size in bytes",
				"example":     1073741824,
				"minimum":     0,
			},
			"used": map[string]interface{}{
				"type":        "integer",
				"description": "Used space in bytes",
				"example":     536870912,
				"minimum":     0,
			},
			"free": map[string]interface{}{
				"type":        "integer",
				"description": "Free space in bytes",
				"example":     536870912,
				"minimum":     0,
			},
			"usage_percent": map[string]interface{}{
				"type":        "number",
				"description": "Usage percentage",
				"example":     50.0,
				"minimum":     0,
				"maximum":     100,
			},
			"usage": map[string]interface{}{
				"type":        "number",
				"description": "Usage percentage (alternative field)",
				"example":     6.6,
				"minimum":     0,
				"maximum":     100,
			},
			"available": map[string]interface{}{
				"type":        "string",
				"description": "Available space (human readable)",
				"example":     "29.9GB",
			},
			"mount_point": map[string]interface{}{
				"type":        "string",
				"description": "Mount point",
				"example":     "/boot",
			},
			"last_updated": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Last update timestamp",
				"example":     "2024-01-01T12:00:00Z",
			},
		},
		"required": []string{"device", "filesystem", "size", "used", "free", "usage", "available", "mount_point", "last_updated"},
	}
}

func getDiskListSchema() map[string]interface{} {
	return map[string]interface{}{
		"type":        "array",
		"description": "List of storage disks",
		"items": map[string]interface{}{
			"$ref": "#/components/schemas/DiskInfo",
		},
		"example": []interface{}{
			map[string]interface{}{
				"name":   "disk1",
				"device": "/dev/sda",
				"size":   8000000000000,
				"status": "active",
			},
		},
	}
}

func getStorageGeneralSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"total_capacity": map[string]interface{}{
				"type":        "integer",
				"description": "Total storage capacity in bytes",
				"example":     50000000000000,
				"minimum":     0,
			},
			"total_used": map[string]interface{}{
				"type":        "integer",
				"description": "Total used storage in bytes",
				"example":     25000000000000,
				"minimum":     0,
			},
			"total_free": map[string]interface{}{
				"type":        "integer",
				"description": "Total free storage in bytes",
				"example":     25000000000000,
				"minimum":     0,
			},
			"usage_percent": map[string]interface{}{
				"type":        "number",
				"description": "Overall usage percentage",
				"example":     50.0,
				"minimum":     0,
				"maximum":     100,
			},
			"disk_count": map[string]interface{}{
				"type":        "integer",
				"description": "Total number of disks",
				"example":     8,
				"minimum":     0,
			},
			"array_status": map[string]interface{}{
				"type":        "string",
				"description": "Array status",
				"enum":        []string{"started", "stopped", "starting", "stopping", "unknown"},
				"example":     "started",
			},
			"parity_valid": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether parity is valid",
				"example":     true,
			},
			"log_usage": map[string]interface{}{
				"type":        "object",
				"description": "Log directory usage information",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Log directory path",
						"example":     "/var/log",
					},
					"total": map[string]interface{}{
						"type":        "number",
						"description": "Total log space in bytes",
						"example":     134217728,
						"minimum":     0,
					},
					"used": map[string]interface{}{
						"type":        "number",
						"description": "Used log space in bytes",
						"example":     4771840,
						"minimum":     0,
					},
					"free": map[string]interface{}{
						"type":        "number",
						"description": "Free log space in bytes",
						"example":     129445888,
						"minimum":     0,
					},
					"usage": map[string]interface{}{
						"type":        "number",
						"description": "Log usage percentage",
						"example":     3.56,
						"minimum":     0,
						"maximum":     100,
					},
					"last_updated": map[string]interface{}{
						"type":        "string",
						"format":      "date-time",
						"description": "Last update timestamp",
						"example":     "2025-06-20T00:56:59Z",
					},
				},
				"required": []string{"path", "total", "used", "free", "usage", "last_updated"},
			},
			"boot_usage": map[string]interface{}{
				"type":        "object",
				"description": "Boot device usage information",
				"properties": map[string]interface{}{
					"device": map[string]interface{}{
						"type":        "string",
						"description": "Boot device path",
						"example":     "/dev/sda1",
					},
					"filesystem": map[string]interface{}{
						"type":        "string",
						"description": "Boot filesystem type",
						"example":     "vfat",
					},
					"size": map[string]interface{}{
						"type":        "string",
						"description": "Boot device size",
						"example":     "32GB",
					},
					"used": map[string]interface{}{
						"type":        "string",
						"description": "Used boot space",
						"example":     "2.1GB",
					},
					"available": map[string]interface{}{
						"type":        "string",
						"description": "Available boot space",
						"example":     "29.9GB",
					},
					"usage": map[string]interface{}{
						"type":        "number",
						"description": "Boot usage percentage",
						"example":     6.6,
						"minimum":     0,
						"maximum":     100,
					},
					"last_updated": map[string]interface{}{
						"type":        "string",
						"format":      "date-time",
						"description": "Last update timestamp",
						"example":     "2025-06-20T00:56:59Z",
					},
				},
				"required": []string{"device", "filesystem", "size", "used", "available", "usage", "last_updated"},
			},
			"docker_vdisk": map[string]interface{}{
				"type":        "object",
				"description": "Docker virtual disk usage information",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Docker vdisk path",
						"example":     "/var/lib/docker",
					},
					"total": map[string]interface{}{
						"type":        "number",
						"description": "Total docker vdisk space in bytes",
						"example":     161061273600,
						"minimum":     0,
					},
					"used": map[string]interface{}{
						"type":        "number",
						"description": "Used docker vdisk space in bytes",
						"example":     12305375232,
						"minimum":     0,
					},
					"free": map[string]interface{}{
						"type":        "number",
						"description": "Free docker vdisk space in bytes",
						"example":     148755898368,
						"minimum":     0,
					},
					"usage": map[string]interface{}{
						"type":        "number",
						"description": "Docker vdisk usage percentage",
						"example":     7.64,
						"minimum":     0,
						"maximum":     100,
					},
					"last_updated": map[string]interface{}{
						"type":        "string",
						"format":      "date-time",
						"description": "Last update timestamp",
						"example":     "2025-06-20T00:56:59Z",
					},
				},
				"required": []string{"path", "total", "used", "free", "usage", "last_updated"},
			},
			"last_updated": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Last update timestamp",
				"example":     "2024-01-01T12:00:00Z",
			},
		},
		"required": []string{"total_capacity", "total_used", "total_free", "usage_percent", "disk_count", "array_status", "log_usage", "boot_usage", "docker_vdisk", "last_updated"},
	}
}

func getZFSInfoSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"pools": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"$ref": "#/components/schemas/ZFSPoolInfo",
				},
				"description": "ZFS pools",
			},
			"datasets": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"$ref": "#/components/schemas/ZFSDatasetInfo",
				},
				"description": "ZFS datasets",
			},
			"total_capacity": map[string]interface{}{
				"type":        "integer",
				"description": "Total ZFS capacity in bytes",
				"example":     2000000000000,
				"minimum":     0,
			},
			"total_used": map[string]interface{}{
				"type":        "integer",
				"description": "Total ZFS used space in bytes",
				"example":     1000000000000,
				"minimum":     0,
			},
			"total_free": map[string]interface{}{
				"type":        "integer",
				"description": "Total ZFS free space in bytes",
				"example":     1000000000000,
				"minimum":     0,
			},
			"overall_health": map[string]interface{}{
				"type":        "string",
				"description": "Overall ZFS health status",
				"enum":        []string{"ONLINE", "DEGRADED", "FAULTED", "OFFLINE", "UNAVAIL"},
				"example":     "ONLINE",
			},
			"version": map[string]interface{}{
				"type":        "string",
				"description": "ZFS version",
				"example":     "2.1.5",
			},
			"last_updated": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Last update timestamp",
				"example":     "2024-01-01T12:00:00Z",
			},
		},
		"required": []string{"pools", "datasets", "total_capacity", "total_used", "total_free", "overall_health", "last_updated"},
	}
}
