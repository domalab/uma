package paths

// GetVMPaths returns all virtual machine management API paths
func GetVMPaths() map[string]interface{} {
	return map[string]interface{}{
		"/api/v1/vms":              getVMsListPath(),
		"/api/v1/vms/{id}":         getVMByIDPath(),
		"/api/v1/vms/{id}/start":   getVMStartPath(),
		"/api/v1/vms/{id}/stop":    getVMStopPath(),
		"/api/v1/vms/{id}/restart": getVMRestartPath(),
		"/api/v1/vms/{id}/pause":   getVMPausePath(),
		"/api/v1/vms/{id}/resume":  getVMResumePath(),
		"/api/v1/vms/{id}/stats":   getVMStatsPath(),
		"/api/v1/vms/bulk/start":   getBulkVMStartPath(),
		"/api/v1/vms/bulk/stop":    getBulkVMStopPath(),
		"/api/v1/vms/bulk/restart": getBulkVMRestartPath(),
	}
}

func getVMsListPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "List virtual machines",
			"description": "Retrieve a list of virtual machines with their current state and configuration",
			"operationId": "listVMs",
			"tags":        []string{"Virtual Machines"},
			"parameters": []interface{}{
				map[string]interface{}{"$ref": "#/components/parameters/PageParameter"},
				map[string]interface{}{"$ref": "#/components/parameters/LimitParameter"},
				map[string]interface{}{"$ref": "#/components/parameters/StatusFilterParameter"},
				map[string]interface{}{"$ref": "#/components/parameters/VerboseParameter"},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "List of VMs retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"allOf": []interface{}{
									map[string]interface{}{"$ref": "#/components/schemas/StandardResponse"},
									map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"data": map[string]interface{}{
												"type": "array",
												"items": map[string]interface{}{
													"$ref": "#/components/schemas/VMInfo",
												},
											},
										},
									},
								},
							},
						},
					},
				},
				"401": map[string]interface{}{"$ref": "#/components/responses/Unauthorized"},
				"500": map[string]interface{}{"$ref": "#/components/responses/InternalServerError"},
			},
			"security": []map[string][]string{
				{"BearerAuth": {}},
				{"ApiKeyAuth": {}},
			},
		},
	}
}

func getVMStartPath() map[string]interface{} {
	return map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Start virtual machine",
			"description": "Start a virtual machine",
			"operationId": "startVM",
			"tags":        []string{"Virtual Machines"},
			"parameters": []interface{}{
				map[string]interface{}{"$ref": "#/components/parameters/VMIDParameter"},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "VM started successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/VMOperationResponse",
							},
						},
					},
				},
				"400": map[string]interface{}{"$ref": "#/components/responses/BadRequest"},
				"401": map[string]interface{}{"$ref": "#/components/responses/Unauthorized"},
				"403": map[string]interface{}{"$ref": "#/components/responses/Forbidden"},
				"404": map[string]interface{}{"$ref": "#/components/responses/NotFound"},
				"409": map[string]interface{}{"$ref": "#/components/responses/Conflict"},
				"500": map[string]interface{}{"$ref": "#/components/responses/InternalServerError"},
			},
			"security": []map[string][]string{
				{"BearerAuth": {}},
				{"ApiKeyAuth": {}},
			},
		},
	}
}

func getBulkVMStartPath() map[string]interface{} {
	return map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Start multiple virtual machines",
			"description": "Start multiple virtual machines in a single operation",
			"operationId": "bulkStartVMs",
			"tags":        []string{"Virtual Machines"},
			"requestBody": map[string]interface{}{
				"required": true,
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"$ref": "#/components/schemas/BulkVMOperation",
						},
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Bulk VM start operation completed",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/BulkVMResponse",
							},
						},
					},
				},
				"400": map[string]interface{}{"$ref": "#/components/responses/BadRequest"},
				"401": map[string]interface{}{"$ref": "#/components/responses/Unauthorized"},
				"403": map[string]interface{}{"$ref": "#/components/responses/Forbidden"},
				"500": map[string]interface{}{"$ref": "#/components/responses/InternalServerError"},
			},
			"security": []map[string][]string{
				{"BearerAuth": {}},
				{"ApiKeyAuth": {}},
			},
		},
	}
}

func getVMByIDPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get virtual machine information",
			"description": "Retrieve detailed information about a specific virtual machine",
			"operationId": "getVM",
			"tags":        []string{"Virtual Machines"},
			"parameters": []interface{}{
				map[string]interface{}{"$ref": "#/components/parameters/VMIDParameter"},
				map[string]interface{}{"$ref": "#/components/parameters/VerboseParameter"},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "VM information retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"allOf": []interface{}{
									map[string]interface{}{"$ref": "#/components/schemas/StandardResponse"},
									map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"data": map[string]interface{}{
												"$ref": "#/components/schemas/VMInfo",
											},
										},
									},
								},
							},
						},
					},
				},
				"400": map[string]interface{}{"$ref": "#/components/responses/BadRequest"},
				"401": map[string]interface{}{"$ref": "#/components/responses/Unauthorized"},
				"404": map[string]interface{}{"$ref": "#/components/responses/NotFound"},
				"500": map[string]interface{}{"$ref": "#/components/responses/InternalServerError"},
			},
			"security": []map[string][]string{
				{"BearerAuth": {}},
				{"ApiKeyAuth": {}},
			},
		},
	}
}

func getVMStopPath() map[string]interface{} {
	return map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Stop virtual machine",
			"description": "Stop a virtual machine gracefully or forcefully",
			"operationId": "stopVM",
			"tags":        []string{"Virtual Machines"},
			"parameters": []interface{}{
				map[string]interface{}{"$ref": "#/components/parameters/VMIDParameter"},
				map[string]interface{}{"$ref": "#/components/parameters/ForceParameter"},
				map[string]interface{}{"$ref": "#/components/parameters/TimeoutParameter"},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "VM stopped successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/VMOperationResponse",
							},
						},
					},
				},
				"400": map[string]interface{}{"$ref": "#/components/responses/BadRequest"},
				"401": map[string]interface{}{"$ref": "#/components/responses/Unauthorized"},
				"403": map[string]interface{}{"$ref": "#/components/responses/Forbidden"},
				"404": map[string]interface{}{"$ref": "#/components/responses/NotFound"},
				"409": map[string]interface{}{"$ref": "#/components/responses/Conflict"},
				"500": map[string]interface{}{"$ref": "#/components/responses/InternalServerError"},
			},
			"security": []map[string][]string{
				{"BearerAuth": {}},
				{"ApiKeyAuth": {}},
			},
		},
	}
}

func getVMRestartPath() map[string]interface{} {
	return map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Restart virtual machine",
			"description": "Restart a virtual machine",
			"operationId": "restartVM",
			"tags":        []string{"Virtual Machines"},
			"parameters": []interface{}{
				map[string]interface{}{"$ref": "#/components/parameters/VMIDParameter"},
				map[string]interface{}{"$ref": "#/components/parameters/TimeoutParameter"},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "VM restarted successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/VMOperationResponse",
							},
						},
					},
				},
				"400": map[string]interface{}{"$ref": "#/components/responses/BadRequest"},
				"401": map[string]interface{}{"$ref": "#/components/responses/Unauthorized"},
				"403": map[string]interface{}{"$ref": "#/components/responses/Forbidden"},
				"404": map[string]interface{}{"$ref": "#/components/responses/NotFound"},
				"500": map[string]interface{}{"$ref": "#/components/responses/InternalServerError"},
			},
			"security": []map[string][]string{
				{"BearerAuth": {}},
				{"ApiKeyAuth": {}},
			},
		},
	}
}

func getVMPausePath() map[string]interface{} {
	return map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Pause virtual machine",
			"description": "Pause a running virtual machine",
			"operationId": "pauseVM",
			"tags":        []string{"Virtual Machines"},
			"parameters": []interface{}{
				map[string]interface{}{"$ref": "#/components/parameters/VMIDParameter"},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "VM paused successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/VMOperationResponse",
							},
						},
					},
				},
				"400": map[string]interface{}{"$ref": "#/components/responses/BadRequest"},
				"401": map[string]interface{}{"$ref": "#/components/responses/Unauthorized"},
				"403": map[string]interface{}{"$ref": "#/components/responses/Forbidden"},
				"404": map[string]interface{}{"$ref": "#/components/responses/NotFound"},
				"409": map[string]interface{}{"$ref": "#/components/responses/Conflict"},
				"500": map[string]interface{}{"$ref": "#/components/responses/InternalServerError"},
			},
			"security": []map[string][]string{
				{"BearerAuth": {}},
				{"ApiKeyAuth": {}},
			},
		},
	}
}

func getVMResumePath() map[string]interface{} {
	return map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Resume virtual machine",
			"description": "Resume a paused virtual machine",
			"operationId": "resumeVM",
			"tags":        []string{"Virtual Machines"},
			"parameters": []interface{}{
				map[string]interface{}{"$ref": "#/components/parameters/VMIDParameter"},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "VM resumed successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/VMOperationResponse",
							},
						},
					},
				},
				"400": map[string]interface{}{"$ref": "#/components/responses/BadRequest"},
				"401": map[string]interface{}{"$ref": "#/components/responses/Unauthorized"},
				"403": map[string]interface{}{"$ref": "#/components/responses/Forbidden"},
				"404": map[string]interface{}{"$ref": "#/components/responses/NotFound"},
				"409": map[string]interface{}{"$ref": "#/components/responses/Conflict"},
				"500": map[string]interface{}{"$ref": "#/components/responses/InternalServerError"},
			},
			"security": []map[string][]string{
				{"BearerAuth": {}},
				{"ApiKeyAuth": {}},
			},
		},
	}
}

func getVMStatsPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get virtual machine statistics",
			"description": "Retrieve performance statistics for a specific virtual machine",
			"operationId": "getVMStats",
			"tags":        []string{"Virtual Machines"},
			"parameters": []interface{}{
				map[string]interface{}{"$ref": "#/components/parameters/VMIDParameter"},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "VM statistics retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"allOf": []interface{}{
									map[string]interface{}{"$ref": "#/components/schemas/StandardResponse"},
									map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"data": map[string]interface{}{
												"$ref": "#/components/schemas/VMStats",
											},
										},
									},
								},
							},
						},
					},
				},
				"400": map[string]interface{}{"$ref": "#/components/responses/BadRequest"},
				"401": map[string]interface{}{"$ref": "#/components/responses/Unauthorized"},
				"404": map[string]interface{}{"$ref": "#/components/responses/NotFound"},
				"500": map[string]interface{}{"$ref": "#/components/responses/InternalServerError"},
			},
			"security": []map[string][]string{
				{"BearerAuth": {}},
				{"ApiKeyAuth": {}},
			},
		},
	}
}

func getBulkVMStopPath() map[string]interface{} {
	return map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Stop multiple virtual machines",
			"description": "Stop multiple virtual machines in a single operation",
			"operationId": "bulkStopVMs",
			"tags":        []string{"Virtual Machines"},
			"requestBody": map[string]interface{}{
				"required": true,
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"$ref": "#/components/schemas/BulkVMOperation",
						},
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Bulk VM stop operation completed",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/BulkVMResponse",
							},
						},
					},
				},
				"400": map[string]interface{}{"$ref": "#/components/responses/BadRequest"},
				"401": map[string]interface{}{"$ref": "#/components/responses/Unauthorized"},
				"403": map[string]interface{}{"$ref": "#/components/responses/Forbidden"},
				"500": map[string]interface{}{"$ref": "#/components/responses/InternalServerError"},
			},
			"security": []map[string][]string{
				{"BearerAuth": {}},
				{"ApiKeyAuth": {}},
			},
		},
	}
}

func getBulkVMRestartPath() map[string]interface{} {
	return map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Restart multiple virtual machines",
			"description": "Restart multiple virtual machines in a single operation",
			"operationId": "bulkRestartVMs",
			"tags":        []string{"Virtual Machines"},
			"requestBody": map[string]interface{}{
				"required": true,
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"$ref": "#/components/schemas/BulkVMOperation",
						},
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Bulk VM restart operation completed",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/BulkVMResponse",
							},
						},
					},
				},
				"400": map[string]interface{}{"$ref": "#/components/responses/BadRequest"},
				"401": map[string]interface{}{"$ref": "#/components/responses/Unauthorized"},
				"403": map[string]interface{}{"$ref": "#/components/responses/Forbidden"},
				"500": map[string]interface{}{"$ref": "#/components/responses/InternalServerError"},
			},
			"security": []map[string][]string{
				{"BearerAuth": {}},
				{"ApiKeyAuth": {}},
			},
		},
	}
}
