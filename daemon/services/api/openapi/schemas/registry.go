package schemas

import (
	"fmt"
)

// Registry manages all OpenAPI schemas
type Registry struct {
	schemas map[string]interface{}
}

// NewRegistry creates a new schema registry
func NewRegistry() *Registry {
	return &Registry{
		schemas: make(map[string]interface{}),
	}
}

// RegisterAll registers all schemas from different modules
func (r *Registry) RegisterAll() {
	// Register common schemas
	for name, schema := range GetCommonSchemas() {
		r.schemas[name] = schema
	}

	// Register Docker schemas
	for name, schema := range GetDockerSchemas() {
		r.schemas[name] = schema
	}

	// Register system schemas
	for name, schema := range GetSystemSchemas() {
		r.schemas[name] = schema
	}

	// Register storage schemas
	for name, schema := range GetStorageSchemas() {
		r.schemas[name] = schema
	}

	// Register VM schemas
	for name, schema := range GetVMSchemas() {
		r.schemas[name] = schema
	}

	// Register WebSocket schemas
	for name, schema := range GetWebSocketSchemas() {
		r.schemas[name] = schema
	}

	// Register authentication schemas
	for name, schema := range GetAuthSchemas() {
		r.schemas[name] = schema
	}

	// Register diagnostics schemas
	for name, schema := range GetDiagnosticsSchemas() {
		r.schemas[name] = schema
	}

	// Register notification schemas
	for name, schema := range GetNotificationSchemas() {
		r.schemas[name] = schema
	}

	// Register operation schemas
	for name, schema := range GetOperationSchemas() {
		r.schemas[name] = schema
	}

	// Register response schemas
	for name, schema := range GetResponseSchemas() {
		r.schemas[name] = schema
	}

	// Register async operations schemas
	r.schemas["AsyncOperationRequest"] = GetAsyncOperationRequest()
	r.schemas["AsyncOperationResponse"] = GetAsyncOperationResponse()
	r.schemas["AsyncOperationDetailResponse"] = GetAsyncOperationDetailResponse()
	r.schemas["AsyncOperationListResponse"] = GetAsyncOperationListResponse()
	r.schemas["AsyncOperationCancelResponse"] = GetAsyncOperationCancelResponse()
	r.schemas["AsyncOperationStatsResponse"] = GetAsyncOperationStatsResponse()

	// Register rate limiting schemas
	r.schemas["RateLimitStatsResponse"] = GetRateLimitStatsResponse()
	r.schemas["RateLimitConfigResponse"] = GetRateLimitConfigResponse()
	r.schemas["RateLimitConfigUpdate"] = GetRateLimitConfigUpdate()
	r.schemas["RateLimitConfigUpdateResponse"] = GetRateLimitConfigUpdateResponse()

	// Register enhanced error schemas
	r.schemas["APIError"] = GetAPIError()
	r.schemas["ValidationError"] = GetValidationError()
	r.schemas["ValidationErrorResponse"] = GetValidationErrorResponse()
	r.schemas["ResourceNotFoundError"] = GetResourceNotFoundError()
	r.schemas["ConflictError"] = GetConflictError()
	r.schemas["RateLimitError"] = GetRateLimitError()
}

// GetAllSchemas returns all registered schemas
func (r *Registry) GetAllSchemas() map[string]interface{} {
	return r.schemas
}

// GetSchema returns a specific schema by name
func (r *Registry) GetSchema(name string) (interface{}, error) {
	schema, exists := r.schemas[name]
	if !exists {
		return nil, fmt.Errorf("schema '%s' not found", name)
	}
	return schema, nil
}

// HasSchema checks if a schema exists
func (r *Registry) HasSchema(name string) bool {
	_, exists := r.schemas[name]
	return exists
}

// ListSchemas returns a list of all schema names
func (r *Registry) ListSchemas() []string {
	names := make([]string, 0, len(r.schemas))
	for name := range r.schemas {
		names = append(names, name)
	}
	return names
}

// GetSchemasByCategory returns schemas grouped by category
func (r *Registry) GetSchemasByCategory() map[string][]string {
	categories := map[string][]string{
		"Common":          {},
		"Docker":          {},
		"System":          {},
		"Storage":         {},
		"VM":              {},
		"WebSocket":       {},
		"Auth":            {},
		"Diagnostics":     {},
		"Notifications":   {},
		"Operations":      {},
		"AsyncOperations": {},
		"RateLimiting":    {},
		"Errors":          {},
		"Responses":       {},
	}

	for name := range r.schemas {
		switch {
		case isCommonSchema(name):
			categories["Common"] = append(categories["Common"], name)
		case isDockerSchema(name):
			categories["Docker"] = append(categories["Docker"], name)
		case isSystemSchema(name):
			categories["System"] = append(categories["System"], name)
		case isStorageSchema(name):
			categories["Storage"] = append(categories["Storage"], name)
		case isVMSchema(name):
			categories["VM"] = append(categories["VM"], name)
		case isWebSocketSchema(name):
			categories["WebSocket"] = append(categories["WebSocket"], name)
		case isAuthSchema(name):
			categories["Auth"] = append(categories["Auth"], name)
		case isDiagnosticsSchema(name):
			categories["Diagnostics"] = append(categories["Diagnostics"], name)
		case isNotificationSchema(name):
			categories["Notifications"] = append(categories["Notifications"], name)
		case isOperationSchema(name):
			categories["Operations"] = append(categories["Operations"], name)
		case isResponseSchema(name):
			categories["Responses"] = append(categories["Responses"], name)
		case isAsyncOperationSchema(name):
			categories["AsyncOperations"] = append(categories["AsyncOperations"], name)
		case isRateLimitingSchema(name):
			categories["RateLimiting"] = append(categories["RateLimiting"], name)
		case isErrorSchema(name):
			categories["Errors"] = append(categories["Errors"], name)
		default:
			categories["Responses"] = append(categories["Responses"], name)
		}
	}

	return categories
}

// Helper functions to categorize schemas
func isCommonSchema(name string) bool {
	commonSchemas := []string{
		"StandardResponse", "PaginationInfo", "ResponseMeta",
		"HealthResponse", "Error", "SuccessResponse",
	}
	for _, schema := range commonSchemas {
		if name == schema {
			return true
		}
	}
	return false
}

func isDockerSchema(name string) bool {
	dockerSchemas := []string{
		"ContainerInfo", "ContainerState", "ContainerOperationResult",
		"ContainerOperationResponse", "BulkOperationRequest", "BulkOperationResponse",
		"BulkOperationSummary", "DockerImage", "DockerNetwork", "DockerInfo",
		"ContainerPort", "DockerContainerList", "DockerContainerInfo",
		"DockerImageList", "DockerNetworkList",
	}
	for _, schema := range dockerSchemas {
		if name == schema {
			return true
		}
	}
	return false
}

func isSystemSchema(name string) bool {
	systemSchemas := []string{
		"SystemInfo", "CPUInfo", "MemoryInfo", "TemperatureData", "FanData",
		"GPUInfo", "UPSInfo", "NetworkInfo", "SystemResources", "FilesystemInfo",
		"SystemScript", "ExecuteRequest", "ExecuteResponse", "LogEntry",
		"SensorChip", "FanInput", "TemperatureInput", "FanInfo", "SystemLogs",
		"ParityCheckStatus", "ParityDiskInfo", "TemperatureInfo",
	}
	for _, schema := range systemSchemas {
		if name == schema {
			return true
		}
	}
	return false
}

func isStorageSchema(name string) bool {
	storageSchemas := []string{
		"ArrayInfo", "DiskInfo", "SMARTData", "ParityInfo", "ParityCheckInfo",
		"CacheInfo", "ZFSPoolInfo", "ZFSDatasetInfo", "ArrayOperation",
		"ArrayStatus", "DiskTemperature", "StorageOverview", "BootInfo",
		"DiskList", "StorageGeneral", "ZFSInfo",
	}
	for _, schema := range storageSchemas {
		if name == schema {
			return true
		}
	}
	return false
}

func isVMSchema(name string) bool {
	vmSchemas := []string{
		"VMInfo", "VMState", "VMOperation", "VMOperationResponse", "VMResources",
		"VMDisk", "VMNetwork", "VMConfig", "VMStats", "VMSnapshot",
		"BulkVMOperation", "BulkVMResponse", "VMList", "VMSnapshotList", "VMSnapshotResponse",
	}
	for _, schema := range vmSchemas {
		if name == schema {
			return true
		}
	}
	return false
}

func isWebSocketSchema(name string) bool {
	wsSchemas := []string{
		"WebSocketMessage", "WebSocketEvent", "WebSocketSubscription",
		"WebSocketError", "WebSocketStats", "WebSocketConnection",
		"DockerEventsStream", "SystemStatsStream", "StorageStatusStream",
	}
	for _, schema := range wsSchemas {
		if name == schema {
			return true
		}
	}
	return false
}

func isAuthSchema(name string) bool {
	authSchemas := []string{
		"LoginRequest", "LoginResponse", "TokenResponse", "RefreshRequest",
		"UserInfo", "APIKeyInfo", "AuthError",
	}
	for _, schema := range authSchemas {
		if name == schema {
			return true
		}
	}
	return false
}

func isAsyncOperationSchema(name string) bool {
	asyncSchemas := []string{
		"AsyncOperationRequest", "AsyncOperationResponse", "AsyncOperationDetailResponse",
		"AsyncOperationListResponse", "AsyncOperationCancelResponse", "AsyncOperationStatsResponse",
	}
	for _, schema := range asyncSchemas {
		if name == schema {
			return true
		}
	}
	return false
}

func isRateLimitingSchema(name string) bool {
	rateLimitSchemas := []string{
		"RateLimitStatsResponse", "RateLimitConfigResponse", "RateLimitConfigUpdate",
		"RateLimitConfigUpdateResponse",
	}
	for _, schema := range rateLimitSchemas {
		if name == schema {
			return true
		}
	}
	return false
}

func isErrorSchema(name string) bool {
	errorSchemas := []string{
		"APIError", "ValidationError", "ValidationErrorResponse",
		"ResourceNotFoundError", "ConflictError", "RateLimitError",
	}
	for _, schema := range errorSchemas {
		if name == schema {
			return true
		}
	}
	return false
}

func isDiagnosticsSchema(name string) bool {
	diagnosticsSchemas := []string{
		"DiagnosticsHealth", "DiagnosticsInfo", "DiagnosticsRepair",
	}
	for _, schema := range diagnosticsSchemas {
		if name == schema {
			return true
		}
	}
	return false
}

func isNotificationSchema(name string) bool {
	notificationSchemas := []string{
		"NotificationList", "NotificationStats", "NotificationInfo",
	}
	for _, schema := range notificationSchemas {
		if name == schema {
			return true
		}
	}
	return false
}

func isOperationSchema(name string) bool {
	operationSchemas := []string{
		"OperationList", "OperationStats", "OperationInfo",
	}
	for _, schema := range operationSchemas {
		if name == schema {
			return true
		}
	}
	return false
}

func isResponseSchema(name string) bool {
	responseSchemas := []string{
		"NotificationResponse", "ParityCheckResponse", "SystemOperationResponse",
		"ArrayOperationResponse", "DockerOperationResponse",
	}
	for _, schema := range responseSchemas {
		if name == schema {
			return true
		}
	}
	return false
}
