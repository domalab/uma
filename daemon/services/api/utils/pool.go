package utils

import (
	"sync"
	"time"

	"github.com/domalab/uma/daemon/services/api/types/responses"
)

// ResponsePool provides object pooling for frequently allocated response structures
type ResponsePool struct {
	standardResponsePool sync.Pool
	responseMetaPool     sync.Pool
	paginationInfoPool   sync.Pool
	operationResponsePool sync.Pool
	bulkOperationResponsePool sync.Pool
	healthResponsePool   sync.Pool
	healthCheckPool      sync.Pool
	mapPool              sync.Pool
	slicePool            sync.Pool
}

// Global response pool instance
var globalResponsePool = NewResponsePool()

// NewResponsePool creates a new response pool
func NewResponsePool() *ResponsePool {
	return &ResponsePool{
		standardResponsePool: sync.Pool{
			New: func() interface{} {
				return &responses.StandardResponse{}
			},
		},
		responseMetaPool: sync.Pool{
			New: func() interface{} {
				return &responses.ResponseMeta{}
			},
		},
		paginationInfoPool: sync.Pool{
			New: func() interface{} {
				return &responses.PaginationInfo{}
			},
		},
		operationResponsePool: sync.Pool{
			New: func() interface{} {
				return &responses.OperationResponse{}
			},
		},
		bulkOperationResponsePool: sync.Pool{
			New: func() interface{} {
				return &responses.BulkOperationResponse{}
			},
		},
		healthResponsePool: sync.Pool{
			New: func() interface{} {
				return &responses.HealthResponse{}
			},
		},
		healthCheckPool: sync.Pool{
			New: func() interface{} {
				return &responses.HealthCheck{}
			},
		},
		mapPool: sync.Pool{
			New: func() interface{} {
				return make(map[string]interface{}, 8) // Pre-allocate with capacity
			},
		},
		slicePool: sync.Pool{
			New: func() interface{} {
				return make([]interface{}, 0, 16) // Pre-allocate with capacity
			},
		},
	}
}

// GetStandardResponse gets a StandardResponse from the pool
func GetStandardResponse() *responses.StandardResponse {
	resp := globalResponsePool.standardResponsePool.Get().(*responses.StandardResponse)
	// Reset the response
	resp.Data = nil
	resp.Error = ""
	resp.Message = ""
	resp.Pagination = nil
	resp.Meta = nil
	return resp
}

// PutStandardResponse returns a StandardResponse to the pool
func PutStandardResponse(resp *responses.StandardResponse) {
	if resp != nil {
		globalResponsePool.standardResponsePool.Put(resp)
	}
}

// GetResponseMeta gets a ResponseMeta from the pool
func GetResponseMeta() *responses.ResponseMeta {
	meta := globalResponsePool.responseMetaPool.Get().(*responses.ResponseMeta)
	// Reset the meta
	meta.RequestID = ""
	meta.Version = ""
	meta.Timestamp = time.Time{}
	return meta
}

// PutResponseMeta returns a ResponseMeta to the pool
func PutResponseMeta(meta *responses.ResponseMeta) {
	if meta != nil {
		globalResponsePool.responseMetaPool.Put(meta)
	}
}

// GetPaginationInfo gets a PaginationInfo from the pool
func GetPaginationInfo() *responses.PaginationInfo {
	pagination := globalResponsePool.paginationInfoPool.Get().(*responses.PaginationInfo)
	// Reset the pagination
	pagination.Page = 0
	pagination.PageSize = 0
	pagination.TotalPages = 0
	pagination.TotalItems = 0
	pagination.HasNext = false
	pagination.HasPrev = false
	return pagination
}

// PutPaginationInfo returns a PaginationInfo to the pool
func PutPaginationInfo(pagination *responses.PaginationInfo) {
	if pagination != nil {
		globalResponsePool.paginationInfoPool.Put(pagination)
	}
}

// GetOperationResponse gets an OperationResponse from the pool
func GetOperationResponse() *responses.OperationResponse {
	resp := globalResponsePool.operationResponsePool.Get().(*responses.OperationResponse)
	// Reset the response
	resp.Success = false
	resp.Message = ""
	resp.OperationID = ""
	return resp
}

// PutOperationResponse returns an OperationResponse to the pool
func PutOperationResponse(resp *responses.OperationResponse) {
	if resp != nil {
		globalResponsePool.operationResponsePool.Put(resp)
	}
}

// GetBulkOperationResponse gets a BulkOperationResponse from the pool
func GetBulkOperationResponse() *responses.BulkOperationResponse {
	resp := globalResponsePool.bulkOperationResponsePool.Get().(*responses.BulkOperationResponse)
	// Reset the response
	resp.Success = false
	resp.Message = ""
	resp.Operation = ""
	resp.Results = resp.Results[:0] // Reset slice but keep capacity
	resp.Summary = responses.BulkOperationSummary{}
	return resp
}

// PutBulkOperationResponse returns a BulkOperationResponse to the pool
func PutBulkOperationResponse(resp *responses.BulkOperationResponse) {
	if resp != nil {
		globalResponsePool.bulkOperationResponsePool.Put(resp)
	}
}

// GetHealthResponse gets a HealthResponse from the pool
func GetHealthResponse() *responses.HealthResponse {
	resp := globalResponsePool.healthResponsePool.Get().(*responses.HealthResponse)
	// Reset the response
	resp.Status = ""
	resp.Version = ""
	resp.Uptime = 0
	resp.Timestamp = time.Time{}
	if resp.Checks == nil {
		resp.Checks = make(map[string]responses.HealthCheck, 4)
	} else {
		// Clear the map but keep the allocated memory
		for k := range resp.Checks {
			delete(resp.Checks, k)
		}
	}
	return resp
}

// PutHealthResponse returns a HealthResponse to the pool
func PutHealthResponse(resp *responses.HealthResponse) {
	if resp != nil {
		globalResponsePool.healthResponsePool.Put(resp)
	}
}

// GetHealthCheck gets a HealthCheck from the pool
func GetHealthCheck() *responses.HealthCheck {
	check := globalResponsePool.healthCheckPool.Get().(*responses.HealthCheck)
	// Reset the check
	check.Status = ""
	check.Message = ""
	check.Timestamp = time.Time{}
	check.Duration = ""
	return check
}

// PutHealthCheck returns a HealthCheck to the pool
func PutHealthCheck(check *responses.HealthCheck) {
	if check != nil {
		globalResponsePool.healthCheckPool.Put(check)
	}
}

// GetMap gets a map[string]interface{} from the pool
func GetMap() map[string]interface{} {
	m := globalResponsePool.mapPool.Get().(map[string]interface{})
	// Clear the map but keep the allocated memory
	for k := range m {
		delete(m, k)
	}
	return m
}

// PutMap returns a map[string]interface{} to the pool
func PutMap(m map[string]interface{}) {
	if m != nil && len(m) < 64 { // Only pool maps that aren't too large
		globalResponsePool.mapPool.Put(m)
	}
}

// GetSlice gets a []interface{} from the pool
func GetSlice() []interface{} {
	s := globalResponsePool.slicePool.Get().([]interface{})
	// Reset the slice but keep capacity
	return s[:0]
}

// PutSlice returns a []interface{} to the pool
func PutSlice(s []interface{}) {
	if s != nil && cap(s) < 128 { // Only pool slices that aren't too large
		globalResponsePool.slicePool.Put(s)
	}
}

// GetResponsePoolStats returns statistics about the response pool usage
func GetResponsePoolStats() map[string]interface{} {
	stats := GetMap()
	stats["standard_response_pool"] = "active"
	stats["response_meta_pool"] = "active"
	stats["pagination_info_pool"] = "active"
	stats["operation_response_pool"] = "active"
	stats["bulk_operation_response_pool"] = "active"
	stats["health_response_pool"] = "active"
	stats["health_check_pool"] = "active"
	stats["map_pool"] = "active"
	stats["slice_pool"] = "active"
	return stats
}
