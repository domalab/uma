package paths

import (
	"github.com/domalab/uma/daemon/services/api/openapi/responses"
	"github.com/domalab/uma/daemon/services/api/openapi/schemas"
)

// AsyncOperationsPaths returns OpenAPI path definitions for async operations endpoints
func AsyncOperationsPaths() map[string]interface{} {
	return map[string]interface{}{
		"/api/v1/operations": map[string]interface{}{
			"get": map[string]interface{}{
				"tags":        []string{"Async Operations"},
				"summary":     "List async operations",
				"description": "Retrieve a list of asynchronous operations with optional filtering",
				"parameters": []map[string]interface{}{
					{
						"name":        "status",
						"in":          "query",
						"description": "Filter operations by status",
						"required":    false,
						"schema": map[string]interface{}{
							"type": "string",
							"enum": []string{"pending", "running", "completed", "failed", "cancelled"},
						},
					},
					{
						"name":        "type",
						"in":          "query",
						"description": "Filter operations by operation type",
						"required":    false,
						"schema": map[string]interface{}{
							"type": "string",
							"enum": []string{
								"parity_check", "parity_correct", "array_start", "array_stop",
								"disk_scan", "smart_scan", "system_reboot", "system_shutdown",
								"bulk_container", "bulk_vm",
							},
						},
					},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "List of operations retrieved successfully",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": schemas.GetAsyncOperationListResponse(),
							},
						},
					},
					"429": responses.RateLimitExceeded(),
					"500": responses.InternalServerError(),
				},
			},
			"post": map[string]interface{}{
				"tags":        []string{"Async Operations"},
				"summary":     "Start async operation",
				"description": "Start a new asynchronous operation",
				"requestBody": map[string]interface{}{
					"required": true,
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": schemas.GetAsyncOperationRequest(),
							"examples": map[string]interface{}{
								"smart_scan": map[string]interface{}{
									"summary": "SMART data collection",
									"value": map[string]interface{}{
										"type":        "smart_scan",
										"description": "Comprehensive SMART data collection for all disks",
										"cancellable": true,
									},
								},
								"parity_check": map[string]interface{}{
									"summary": "Parity check operation",
									"value": map[string]interface{}{
										"type":        "parity_check",
										"description": "Full parity check of the array",
										"cancellable": true,
										"parameters": map[string]interface{}{
											"type":     "check",
											"priority": "normal",
										},
									},
								},
								"bulk_container": map[string]interface{}{
									"summary": "Bulk container operation",
									"value": map[string]interface{}{
										"type":        "bulk_container",
										"description": "Start multiple Docker containers",
										"cancellable": false,
										"parameters": map[string]interface{}{
											"container_ids": []string{"container1", "container2"},
											"operation":     "start",
										},
									},
								},
							},
						},
					},
				},
				"responses": map[string]interface{}{
					"201": map[string]interface{}{
						"description": "Operation started successfully",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": schemas.GetAsyncOperationResponse(),
							},
						},
					},
					"400": responses.BadRequest(),
					"409": responses.Conflict(),
					"429": responses.RateLimitExceeded(),
				},
			},
		},
		"/api/v1/operations/{operationId}": map[string]interface{}{
			"get": map[string]interface{}{
				"tags":        []string{"Async Operations"},
				"summary":     "Get operation details",
				"description": "Retrieve detailed information about a specific operation",
				"parameters": []map[string]interface{}{
					{
						"name":        "operationId",
						"in":          "path",
						"required":    true,
						"description": "Unique identifier of the operation",
						"schema": map[string]interface{}{
							"type":    "string",
							"format":  "uuid",
							"example": "4a799a05-bb59-42a1-ab3b-d4ccbfdc623f",
						},
					},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Operation details retrieved successfully",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": schemas.GetAsyncOperationDetailResponse(),
							},
						},
					},
					"404": responses.NotFound(),
					"429": responses.RateLimitExceeded(),
				},
			},
			"delete": map[string]interface{}{
				"tags":        []string{"Async Operations"},
				"summary":     "Cancel operation",
				"description": "Cancel a running or pending operation (if cancellable)",
				"parameters": []map[string]interface{}{
					{
						"name":        "operationId",
						"in":          "path",
						"required":    true,
						"description": "Unique identifier of the operation to cancel",
						"schema": map[string]interface{}{
							"type":    "string",
							"format":  "uuid",
							"example": "4a799a05-bb59-42a1-ab3b-d4ccbfdc623f",
						},
					},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Operation cancelled successfully",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": schemas.GetAsyncOperationCancelResponse(),
							},
						},
					},
					"400": responses.BadRequest(),
					"404": responses.NotFound(),
					"429": responses.RateLimitExceeded(),
				},
			},
		},
		"/api/v1/operations/stats": map[string]interface{}{
			"get": map[string]interface{}{
				"tags":        []string{"Async Operations"},
				"summary":     "Get operation statistics",
				"description": "Retrieve statistics about async operations",
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Operation statistics retrieved successfully",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": schemas.GetAsyncOperationStatsResponse(),
							},
						},
					},
					"429": responses.RateLimitExceeded(),
				},
			},
		},
	}
}

// RateLimitingPaths returns OpenAPI path definitions for rate limiting endpoints
func RateLimitingPaths() map[string]interface{} {
	return map[string]interface{}{
		"/api/v1/rate-limits/stats": map[string]interface{}{
			"get": map[string]interface{}{
				"tags":        []string{"Rate Limiting"},
				"summary":     "Get rate limiting statistics",
				"description": "Retrieve current rate limiting statistics for all operation types and clients",
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Rate limiting statistics retrieved successfully",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": schemas.GetRateLimitStatsResponse(),
								"examples": map[string]interface{}{
									"example_stats": map[string]interface{}{
										"summary": "Example rate limiting statistics",
										"value": map[string]interface{}{
											"data": map[string]interface{}{
												"general_rate_limiter": map[string]interface{}{
													"type": "general",
												},
												"operation_rate_limiter": map[string]interface{}{
													"type":          "operation_specific",
													"total_clients": 3,
													"operation_limits": map[string]interface{}{
														"smart_data": map[string]interface{}{
															"requests": 1,
															"window":   "1m0s",
														},
														"parity_check": map[string]interface{}{
															"requests": 1,
															"window":   "1h0m0s",
														},
													},
													"client_stats": map[string]interface{}{
														"192.168.1.100": map[string]interface{}{
															"smart_data": map[string]interface{}{
																"tokens_remaining": 0,
																"max_tokens":       1,
															},
														},
													},
												},
											},
											"meta": map[string]interface{}{
												"timestamp":   1640995200,
												"api_version": "v1",
											},
										},
									},
								},
							},
						},
					},
					"429": responses.RateLimitExceeded(),
				},
			},
		},
		"/api/v1/rate-limits/config": map[string]interface{}{
			"get": map[string]interface{}{
				"tags":        []string{"Rate Limiting"},
				"summary":     "Get rate limiting configuration",
				"description": "Retrieve current rate limiting configuration for all operation types",
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Rate limiting configuration retrieved successfully",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": schemas.GetRateLimitConfigResponse(),
							},
						},
					},
					"429": responses.RateLimitExceeded(),
				},
			},
			"put": map[string]interface{}{
				"tags":        []string{"Rate Limiting"},
				"summary":     "Update rate limiting configuration",
				"description": "Update rate limiting configuration for operation types (requires admin privileges)",
				"requestBody": map[string]interface{}{
					"required": true,
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": schemas.GetRateLimitConfigUpdate(),
							"examples": map[string]interface{}{
								"update_smart_data": map[string]interface{}{
									"summary": "Update SMART data rate limit",
									"value": map[string]interface{}{
										"smart_data": map[string]interface{}{
											"requests": 2,
											"window":   "2m0s",
										},
									},
								},
							},
						},
					},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Rate limiting configuration updated successfully",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": schemas.GetRateLimitConfigUpdateResponse(),
							},
						},
					},
					"400": responses.BadRequest(),
					"401": responses.Unauthorized(),
					"403": responses.Forbidden(),
					"501": responses.NotImplemented(),
				},
			},
		},
	}
}
