package schemas

// GetVMSchemas returns virtual machine management schemas
func GetVMSchemas() map[string]interface{} {
	return map[string]interface{}{
		"VMInfo":              getVMInfoSchema(),
		"VMState":             getVMStateSchema(),
		"VMOperation":         getVMOperationSchema(),
		"VMOperationResponse": getVMOperationResponseSchema(),
		"VMResources":         getVMResourcesSchema(),
		"VMDisk":              getVMDiskSchema(),
		"VMNetwork":           getVMNetworkSchema(),
		"VMConfig":            getVMConfigSchema(),
		"VMStats":             getVMStatsSchema(),
		"VMSnapshot":          getVMSnapshotSchema(),
		"BulkVMOperation":     getBulkVMOperationSchema(),
		"BulkVMResponse":      getBulkVMResponseSchema(),
		"VMList":              getVMListSchema(),
		"VMSnapshotList":      getVMSnapshotListSchema(),
		"VMSnapshotResponse":  getVMSnapshotResponseSchema(),
	}
}

func getVMInfoSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":        "string",
				"description": "VM identifier",
				"example":     "vm-001",
			},
			"name": map[string]interface{}{
				"type":        "string",
				"description": "VM name",
				"example":     "Windows-10-Gaming",
				"pattern":     "^[a-zA-Z0-9][a-zA-Z0-9_.-]+$",
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "VM description",
				"example":     "Windows 10 gaming virtual machine",
			},
			"state": map[string]interface{}{
				"$ref": "#/components/schemas/VMState",
			},
			"os_type": map[string]interface{}{
				"type":        "string",
				"description": "Operating system type",
				"enum":        []string{"windows", "linux", "macos", "other"},
				"example":     "windows",
			},
			"template": map[string]interface{}{
				"type":        "string",
				"description": "VM template used",
				"example":     "Windows 10",
			},
			"resources": map[string]interface{}{
				"$ref": "#/components/schemas/VMResources",
			},
			"disks": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"$ref": "#/components/schemas/VMDisk",
				},
				"description": "VM disk attachments",
			},
			"networks": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"$ref": "#/components/schemas/VMNetwork",
				},
				"description": "VM network interfaces",
			},
			"created": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "VM creation timestamp",
				"example":     "2025-06-16T14:30:00Z",
			},
			"last_updated": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Last update timestamp",
				"example":     "2025-06-16T14:30:00Z",
			},
		},
		"required": []string{"id", "name", "state", "os_type", "resources", "created", "last_updated"},
	}
}

func getVMStateSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"status": map[string]interface{}{
				"type":        "string",
				"description": "VM status",
				"enum":        []string{"running", "stopped", "paused", "suspended", "starting", "stopping", "error"},
				"example":     "running",
			},
			"uptime": map[string]interface{}{
				"type":        "integer",
				"description": "VM uptime in seconds",
				"example":     3600,
				"minimum":     0,
			},
			"cpu_usage": map[string]interface{}{
				"type":        "number",
				"description": "CPU usage percentage",
				"example":     25.5,
				"minimum":     0,
				"maximum":     100,
			},
			"memory_usage": map[string]interface{}{
				"type":        "number",
				"description": "Memory usage percentage",
				"example":     45.2,
				"minimum":     0,
				"maximum":     100,
			},
			"memory_used": map[string]interface{}{
				"type":        "integer",
				"description": "Used memory in bytes",
				"example":     4294967296,
				"minimum":     0,
			},
			"vnc_port": map[string]interface{}{
				"type":        "integer",
				"description": "VNC port number",
				"example":     5900,
				"minimum":     5900,
				"maximum":     6000,
			},
			"autostart": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether VM starts automatically",
				"example":     true,
			},
			"last_updated": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Last state update timestamp",
				"example":     "2025-06-16T14:30:00Z",
			},
		},
		"required": []string{"status", "last_updated"},
	}
}

func getVMResourcesSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"vcpus": map[string]interface{}{
				"type":        "integer",
				"description": "Number of virtual CPUs",
				"example":     4,
				"minimum":     1,
				"maximum":     64,
			},
			"memory": map[string]interface{}{
				"type":        "integer",
				"description": "Allocated memory in bytes",
				"example":     8589934592,
				"minimum":     134217728,
			},
			"memory_mb": map[string]interface{}{
				"type":        "integer",
				"description": "Allocated memory in MB",
				"example":     8192,
				"minimum":     128,
			},
			"cpu_mode": map[string]interface{}{
				"type":        "string",
				"description": "CPU mode",
				"enum":        []string{"host-passthrough", "host-model", "custom"},
				"example":     "host-passthrough",
			},
			"cpu_topology": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"sockets": map[string]interface{}{
						"type":    "integer",
						"example": 1,
						"minimum": 1,
					},
					"cores": map[string]interface{}{
						"type":    "integer",
						"example": 4,
						"minimum": 1,
					},
					"threads": map[string]interface{}{
						"type":    "integer",
						"example": 1,
						"minimum": 1,
					},
				},
			},
			"machine_type": map[string]interface{}{
				"type":        "string",
				"description": "Machine type",
				"example":     "pc-q35-6.2",
			},
			"bios": map[string]interface{}{
				"type":        "string",
				"description": "BIOS type",
				"enum":        []string{"seabios", "ovmf"},
				"example":     "ovmf",
			},
		},
		"required": []string{"vcpus", "memory", "memory_mb"},
	}
}

func getVMDiskSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"device": map[string]interface{}{
				"type":        "string",
				"description": "Disk device name",
				"example":     "vda",
			},
			"source": map[string]interface{}{
				"type":        "string",
				"description": "Disk source path",
				"example":     "/mnt/user/domains/Windows-10-Gaming/vdisk1.img",
			},
			"type": map[string]interface{}{
				"type":        "string",
				"description": "Disk type",
				"enum":        []string{"file", "block", "network"},
				"example":     "file",
			},
			"bus": map[string]interface{}{
				"type":        "string",
				"description": "Disk bus type",
				"enum":        []string{"virtio", "sata", "ide", "scsi"},
				"example":     "virtio",
			},
			"format": map[string]interface{}{
				"type":        "string",
				"description": "Disk format",
				"enum":        []string{"raw", "qcow2", "vmdk", "vdi"},
				"example":     "raw",
			},
			"size": map[string]interface{}{
				"type":        "integer",
				"description": "Disk size in bytes",
				"example":     107374182400,
				"minimum":     0,
			},
			"readonly": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether disk is read-only",
				"example":     false,
			},
			"bootable": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether disk is bootable",
				"example":     true,
			},
		},
		"required": []string{"device", "source", "type", "bus"},
	}
}

func getVMNetworkSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"interface": map[string]interface{}{
				"type":        "string",
				"description": "Network interface name",
				"example":     "vnet0",
			},
			"type": map[string]interface{}{
				"type":        "string",
				"description": "Network type",
				"enum":        []string{"bridge", "network", "direct"},
				"example":     "bridge",
			},
			"source": map[string]interface{}{
				"type":        "string",
				"description": "Network source",
				"example":     "br0",
			},
			"model": map[string]interface{}{
				"type":        "string",
				"description": "Network model",
				"enum":        []string{"virtio", "e1000", "rtl8139"},
				"example":     "virtio",
			},
			"mac_address": map[string]interface{}{
				"type":        "string",
				"description": "MAC address",
				"example":     "52:54:00:12:34:56",
				"pattern":     "^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$",
			},
			"link_state": map[string]interface{}{
				"type":        "string",
				"description": "Link state",
				"enum":        []string{"up", "down"},
				"example":     "up",
			},
		},
		"required": []string{"interface", "type", "source", "model"},
	}
}

func getVMOperationSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"operation": map[string]interface{}{
				"type":        "string",
				"description": "VM operation to perform",
				"enum":        []string{"start", "stop", "restart", "pause", "resume", "suspend", "reset"},
				"example":     "start",
			},
			"force": map[string]interface{}{
				"type":        "boolean",
				"description": "Force the operation",
				"example":     false,
				"default":     false,
			},
		},
		"required": []string{"operation"},
	}
}

func getVMOperationResponseSchema() map[string]interface{} {
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
				"example":     "VM started successfully",
			},
			"vm_id": map[string]interface{}{
				"type":        "string",
				"description": "VM identifier",
				"example":     "vm-001",
			},
			"operation": map[string]interface{}{
				"type":        "string",
				"description": "Operation that was performed",
				"example":     "start",
			},
			"state": map[string]interface{}{
				"type":        "string",
				"description": "Current VM state after operation",
				"example":     "running",
			},
		},
		"required": []string{"success", "message", "vm_id", "operation"},
	}
}

func getVMConfigSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "VM name",
				"example":     "Windows-10-Gaming",
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "VM description",
				"example":     "Windows 10 gaming virtual machine",
			},
			"os_type": map[string]interface{}{
				"type":        "string",
				"description": "Operating system type",
				"enum":        []string{"windows", "linux", "macos", "other"},
				"example":     "windows",
			},
			"autostart": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether VM starts automatically",
				"example":     true,
			},
			"resources": map[string]interface{}{
				"$ref": "#/components/schemas/VMResources",
			},
			"disks": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"$ref": "#/components/schemas/VMDisk",
				},
			},
			"networks": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"$ref": "#/components/schemas/VMNetwork",
				},
			},
		},
		"required": []string{"name", "os_type", "resources"},
	}
}

func getVMStatsSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"vm_id": map[string]interface{}{
				"type":        "string",
				"description": "VM identifier",
				"example":     "vm-001",
			},
			"cpu_usage": map[string]interface{}{
				"type":        "number",
				"description": "CPU usage percentage",
				"example":     25.5,
				"minimum":     0,
				"maximum":     100,
			},
			"memory_usage": map[string]interface{}{
				"type":        "number",
				"description": "Memory usage percentage",
				"example":     45.2,
				"minimum":     0,
				"maximum":     100,
			},
			"memory_used": map[string]interface{}{
				"type":        "integer",
				"description": "Used memory in bytes",
				"example":     4294967296,
				"minimum":     0,
			},
			"memory_total": map[string]interface{}{
				"type":        "integer",
				"description": "Total allocated memory in bytes",
				"example":     8589934592,
				"minimum":     0,
			},
			"disk_read": map[string]interface{}{
				"type":        "integer",
				"description": "Disk read bytes",
				"example":     1073741824,
				"minimum":     0,
			},
			"disk_write": map[string]interface{}{
				"type":        "integer",
				"description": "Disk write bytes",
				"example":     536870912,
				"minimum":     0,
			},
			"network_rx": map[string]interface{}{
				"type":        "integer",
				"description": "Network received bytes",
				"example":     268435456,
				"minimum":     0,
			},
			"network_tx": map[string]interface{}{
				"type":        "integer",
				"description": "Network transmitted bytes",
				"example":     134217728,
				"minimum":     0,
			},
			"uptime": map[string]interface{}{
				"type":        "integer",
				"description": "VM uptime in seconds",
				"example":     3600,
				"minimum":     0,
			},
			"last_updated": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Last update timestamp",
				"example":     "2025-06-16T14:30:00Z",
			},
		},
		"required": []string{"vm_id", "cpu_usage", "memory_usage", "last_updated"},
	}
}

func getVMSnapshotSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Snapshot name",
				"example":     "pre-update-snapshot",
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "Snapshot description",
				"example":     "Snapshot before system update",
			},
			"state": map[string]interface{}{
				"type":        "string",
				"description": "VM state when snapshot was taken",
				"enum":        []string{"running", "shutoff", "paused"},
				"example":     "shutoff",
			},
			"creation_time": map[string]interface{}{
				"type":        "string",
				"format":      "date-time",
				"description": "Snapshot creation timestamp",
				"example":     "2025-06-16T14:30:00Z",
			},
			"parent": map[string]interface{}{
				"type":        "string",
				"description": "Parent snapshot name",
				"example":     "base-snapshot",
			},
			"current": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether this is the current snapshot",
				"example":     true,
			},
		},
		"required": []string{"name", "state", "creation_time"},
	}
}

func getBulkVMOperationSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"vm_ids": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
				"description": "Array of VM IDs or names",
				"example":     []string{"vm-001", "vm-002", "Windows-10-Gaming"},
				"minItems":    1,
				"maxItems":    20,
				"uniqueItems": true,
			},
			"operation": map[string]interface{}{
				"type":        "string",
				"description": "Operation to perform on all VMs",
				"enum":        []string{"start", "stop", "restart", "pause", "resume"},
				"example":     "start",
			},
			"force": map[string]interface{}{
				"type":        "boolean",
				"description": "Force the operation",
				"example":     false,
				"default":     false,
			},
		},
		"required": []string{"vm_ids", "operation"},
	}
}

func getBulkVMResponseSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"summary": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"total": map[string]interface{}{
						"type":        "integer",
						"description": "Total number of VMs processed",
						"example":     3,
						"minimum":     0,
					},
					"successful": map[string]interface{}{
						"type":        "integer",
						"description": "Number of successful operations",
						"example":     2,
						"minimum":     0,
					},
					"failed": map[string]interface{}{
						"type":        "integer",
						"description": "Number of failed operations",
						"example":     1,
						"minimum":     0,
					},
					"operation": map[string]interface{}{
						"type":        "string",
						"description": "Operation performed",
						"example":     "start",
					},
				},
				"required": []string{"total", "successful", "failed", "operation"},
			},
			"results": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"vm_id": map[string]interface{}{
							"type":        "string",
							"description": "VM identifier",
							"example":     "vm-001",
						},
						"success": map[string]interface{}{
							"type":        "boolean",
							"description": "Whether the operation was successful",
							"example":     true,
						},
						"message": map[string]interface{}{
							"type":        "string",
							"description": "Operation result message",
							"example":     "VM started successfully",
						},
						"error": map[string]interface{}{
							"type":        "string",
							"description": "Error message if operation failed",
							"example":     "VM not found",
						},
						"state": map[string]interface{}{
							"type":        "string",
							"description": "Current VM state after operation",
							"example":     "running",
						},
					},
					"required": []string{"vm_id", "success", "message"},
				},
				"description": "Individual operation results",
			},
		},
		"required": []string{"summary", "results"},
	}
}

func getVMListSchema() map[string]interface{} {
	return map[string]interface{}{
		"type":        "array",
		"description": "List of virtual machines",
		"items": map[string]interface{}{
			"$ref": "#/components/schemas/VMInfo",
		},
		"example": []interface{}{
			map[string]interface{}{
				"id":       "vm-001",
				"name":     "Ubuntu-Server",
				"state":    "running",
				"cpu":      2,
				"memory":   4096,
				"template": "Ubuntu",
			},
		},
	}
}

func getVMSnapshotListSchema() map[string]interface{} {
	return map[string]interface{}{
		"type":        "array",
		"description": "List of VM snapshots",
		"items": map[string]interface{}{
			"$ref": "#/components/schemas/VMSnapshot",
		},
		"example": []interface{}{
			map[string]interface{}{
				"name":          "pre-update-snapshot",
				"description":   "Snapshot before system update",
				"state":         "shutoff",
				"creation_time": "2025-06-16T14:30:00Z",
				"current":       false,
			},
		},
	}
}

func getVMSnapshotResponseSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"success": map[string]interface{}{
				"type":        "boolean",
				"description": "Whether the snapshot operation was successful",
				"example":     true,
			},
			"message": map[string]interface{}{
				"type":        "string",
				"description": "Operation result message",
				"example":     "Snapshot created successfully",
			},
			"snapshot": map[string]interface{}{
				"$ref": "#/components/schemas/VMSnapshot",
			},
			"vm_id": map[string]interface{}{
				"type":        "string",
				"description": "VM identifier",
				"example":     "vm-001",
			},
			"operation": map[string]interface{}{
				"type":        "string",
				"description": "Snapshot operation performed",
				"enum":        []string{"create", "delete", "revert"},
				"example":     "create",
			},
		},
		"required": []string{"success", "message", "vm_id", "operation"},
	}
}
