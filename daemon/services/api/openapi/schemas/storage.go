package schemas

// GetStorageSchemas returns storage management schemas
func GetStorageSchemas() map[string]interface{} {
	return map[string]interface{}{
		"ArrayInfo":       getArrayInfoSchema(),
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
			"status": map[string]interface{}{
				"type":        "string",
				"description": "Array status",
				"enum":        []string{"started", "stopped", "starting", "stopping"},
				"example":     "started",
			},
			"state": map[string]interface{}{
				"type":        "string",
				"description": "Array state",
				"enum":        []string{"normal", "degraded", "invalid", "emulated"},
				"example":     "normal",
			},
			"num_disks": map[string]interface{}{
				"type":        "integer",
				"description": "Number of data disks",
				"example":     6,
				"minimum":     0,
			},
			"num_parity": map[string]interface{}{
				"type":        "integer",
				"description": "Number of parity disks",
				"example":     2,
				"minimum":     0,
				"maximum":     2,
			},
			"size": map[string]interface{}{
				"type":        "integer",
				"description": "Total array size in bytes",
				"example":     48000000000000,
				"minimum":     0,
			},
			"free": map[string]interface{}{
				"type":        "integer",
				"description": "Free space in bytes",
				"example":     24000000000000,
				"minimum":     0,
			},
			"used": map[string]interface{}{
				"type":        "integer",
				"description": "Used space in bytes",
				"example":     24000000000000,
				"minimum":     0,
			},
			"usage_percent": map[string]interface{}{
				"type":        "number",
				"description": "Usage percentage",
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
		"required": []string{"status", "state", "num_disks", "num_parity", "last_updated"},
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
		"required": []string{"name", "device", "size", "status", "last_updated"},
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
				"$ref": "#/components/schemas/DiskInfo",
			},
			"parity2": map[string]interface{}{
				"$ref": "#/components/schemas/DiskInfo",
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
			"last_updated": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Last update timestamp",
				"example":     "2025-06-16T14:30:00Z",
			},
		},
		"required": []string{"disks", "pool_status", "last_updated"},
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
		"required": []string{"device", "filesystem", "size", "used", "free", "mount_point", "last_updated"},
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
				"enum":        []string{"started", "stopped", "starting", "stopping"},
				"example":     "started",
			},
			"parity_valid": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether parity is valid",
				"example":     true,
			},
			"last_updated": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Last update timestamp",
				"example":     "2024-01-01T12:00:00Z",
			},
		},
		"required": []string{"total_capacity", "total_used", "total_free", "usage_percent", "disk_count", "array_status", "last_updated"},
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
