package paths

// GetStoragePaths returns all storage management API paths
func GetStoragePaths() map[string]interface{} {
	return map[string]interface{}{
		"/api/v1/storage/array":        getArrayInfoPath(),
		"/api/v1/storage/array/start":  getArrayStartPath(),
		"/api/v1/storage/array/stop":   getArrayStopPath(),
		"/api/v1/storage/disks":        getDisksListPath(),
		"/api/v1/storage/disks/{id}":   getDiskByIDPath(),
		"/api/v1/storage/parity":       getParityInfoPath(),
		"/api/v1/storage/parity/check": getParityCheckPath(),
		"/api/v1/storage/cache":        getCacheInfoPath(),
		"/api/v1/storage/zfs/pools":    getZFSPoolsPath(),
		"/api/v1/storage/zfs/datasets": getZFSDatasetsPath(),
		"/api/v1/storage/overview":     getStorageOverviewPath(),
		"/api/v1/storage/temperatures": getDiskTemperaturesPath(),
	}
}

func getArrayInfoPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get array information",
			"description": "Retrieve Unraid array status, state, disk count, and usage information",
			"operationId": "getArrayInfo",
			"tags":        []string{"Storage"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Array information retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"allOf": []interface{}{
									map[string]interface{}{"$ref": "#/components/schemas/StandardResponse"},
									map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"data": map[string]interface{}{
												"$ref": "#/components/schemas/ArrayInfo",
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

func getArrayStartPath() map[string]interface{} {
	return map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Start Unraid array",
			"description": "Start the Unraid array with proper orchestration sequence",
			"operationId": "startArray",
			"tags":        []string{"Storage"},
			"requestBody": map[string]interface{}{
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"$ref": "#/components/schemas/ArrayOperation",
						},
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Array start operation completed",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/ArrayStatus",
							},
						},
					},
				},
				"400": map[string]interface{}{"$ref": "#/components/responses/BadRequest"},
				"401": map[string]interface{}{"$ref": "#/components/responses/Unauthorized"},
				"403": map[string]interface{}{"$ref": "#/components/responses/Forbidden"},
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

func getArrayStopPath() map[string]interface{} {
	return map[string]interface{}{
		"post": map[string]interface{}{
			"summary":     "Stop Unraid array",
			"description": "Stop the Unraid array with proper orchestration sequence (Docker stop → VM stop → unmount shares → unmount disks → stop parity → mdadm stop)",
			"operationId": "stopArray",
			"tags":        []string{"Storage"},
			"requestBody": map[string]interface{}{
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"schema": map[string]interface{}{
							"$ref": "#/components/schemas/ArrayOperation",
						},
					},
				},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Array stop operation completed",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"$ref": "#/components/schemas/ArrayStatus",
							},
						},
					},
				},
				"400": map[string]interface{}{"$ref": "#/components/responses/BadRequest"},
				"401": map[string]interface{}{"$ref": "#/components/responses/Unauthorized"},
				"403": map[string]interface{}{"$ref": "#/components/responses/Forbidden"},
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

func getDisksListPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "List storage disks",
			"description": "Retrieve information about all storage disks including data disks, parity disks, and cache disks",
			"operationId": "listDisks",
			"tags":        []string{"Storage"},
			"parameters": []interface{}{
				map[string]interface{}{"$ref": "#/components/parameters/SMARTParameter"},
				map[string]interface{}{"$ref": "#/components/parameters/TemperatureParameter"},
				map[string]interface{}{"$ref": "#/components/parameters/VerboseParameter"},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Disk list retrieved successfully",
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
													"$ref": "#/components/schemas/DiskInfo",
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

func getDiskByIDPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get disk information",
			"description": "Retrieve detailed information about a specific disk including SMART data and temperature",
			"operationId": "getDisk",
			"tags":        []string{"Storage"},
			"parameters": []interface{}{
				map[string]interface{}{"$ref": "#/components/parameters/DiskIDParameter"},
				map[string]interface{}{"$ref": "#/components/parameters/SMARTParameter"},
				map[string]interface{}{"$ref": "#/components/parameters/TemperatureParameter"},
			},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Disk information retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"allOf": []interface{}{
									map[string]interface{}{"$ref": "#/components/schemas/StandardResponse"},
									map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"data": map[string]interface{}{
												"$ref": "#/components/schemas/DiskInfo",
											},
										},
									},
								},
							},
						},
					},
				},
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

func getStorageOverviewPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get storage overview",
			"description": "Retrieve comprehensive storage overview including array, parity, cache, disks, and ZFS pools",
			"operationId": "getStorageOverview",
			"tags":        []string{"Storage"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Storage overview retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"allOf": []interface{}{
									map[string]interface{}{"$ref": "#/components/schemas/StandardResponse"},
									map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"data": map[string]interface{}{
												"$ref": "#/components/schemas/StorageOverview",
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

func getParityInfoPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get parity information",
			"description": "Retrieve parity disk information and check status",
			"operationId": "getParityInfo",
			"tags":        []string{"Storage"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Parity information retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"allOf": []interface{}{
									map[string]interface{}{"$ref": "#/components/schemas/StandardResponse"},
									map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"data": map[string]interface{}{
												"$ref": "#/components/schemas/ParityInfo",
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

func getParityCheckPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get parity check status",
			"description": "Retrieve current parity check status and progress",
			"operationId": "getParityCheck",
			"tags":        []string{"Storage"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Parity check status retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"allOf": []interface{}{
									map[string]interface{}{"$ref": "#/components/schemas/StandardResponse"},
									map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"data": map[string]interface{}{
												"$ref": "#/components/schemas/ParityCheckInfo",
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

func getCacheInfoPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get cache information",
			"description": "Retrieve cache pool information and disk status",
			"operationId": "getCacheInfo",
			"tags":        []string{"Storage"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Cache information retrieved successfully",
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"allOf": []interface{}{
									map[string]interface{}{"$ref": "#/components/schemas/StandardResponse"},
									map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"data": map[string]interface{}{
												"$ref": "#/components/schemas/CacheInfo",
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

func getZFSPoolsPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "List ZFS pools",
			"description": "Retrieve a list of ZFS pools with status and health information",
			"operationId": "listZFSPools",
			"tags":        []string{"Storage"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "ZFS pools retrieved successfully",
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
													"$ref": "#/components/schemas/ZFSPoolInfo",
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

func getZFSDatasetsPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "List ZFS datasets",
			"description": "Retrieve a list of ZFS datasets with usage information",
			"operationId": "listZFSDatasets",
			"tags":        []string{"Storage"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "ZFS datasets retrieved successfully",
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
													"$ref": "#/components/schemas/ZFSDatasetInfo",
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

func getDiskTemperaturesPath() map[string]interface{} {
	return map[string]interface{}{
		"get": map[string]interface{}{
			"summary":     "Get disk temperatures",
			"description": "Retrieve temperature information for all storage disks",
			"operationId": "getDiskTemperatures",
			"tags":        []string{"Storage"},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Disk temperatures retrieved successfully",
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
													"$ref": "#/components/schemas/DiskTemperature",
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
