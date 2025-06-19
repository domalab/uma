package utils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/domalab/uma/daemon/dto"
	"github.com/domalab/uma/daemon/services/api/types/requests"
	"github.com/domalab/uma/daemon/services/api/types/responses"
)

// BenchmarkWriteJSON tests JSON response writing performance
func BenchmarkWriteJSON(b *testing.B) {
	data := map[string]interface{}{
		"message": "test response",
		"data": map[string]interface{}{
			"items": []string{"item1", "item2", "item3"},
			"count": 3,
		},
		"timestamp": time.Now().Unix(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		WriteJSON(w, http.StatusOK, data)
	}
}

// BenchmarkWriteJSONLarge tests performance with large JSON responses
func BenchmarkWriteJSONLarge(b *testing.B) {
	// Create a large data structure
	data := make(map[string]interface{})
	for i := 0; i < 1000; i++ {
		data[string(rune('a'+i%26))+string(rune('a'+(i/26)%26))] = map[string]interface{}{
			"id":          i,
			"name":        "Item " + string(rune('0'+i%10)),
			"description": "This is a test item with a longer description to simulate real data",
			"metadata": map[string]string{
				"category": "test",
				"type":     "benchmark",
				"status":   "active",
			},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		WriteJSON(w, http.StatusOK, data)
	}
}

// BenchmarkWriteError tests error response performance
func BenchmarkWriteError(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		WriteError(w, http.StatusBadRequest, "Test error message")
	}
}

// BenchmarkWritePaginatedResponse tests paginated response performance
func BenchmarkWritePaginatedResponse(b *testing.B) {
	data := make([]map[string]interface{}, 100)
	for i := 0; i < 100; i++ {
		data[i] = map[string]interface{}{
			"id":   i,
			"name": "Item " + string(rune('0'+i%10)),
		}
	}

	params := &dto.PaginationParams{
		Page:    1,
		PerPage: 20,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		WritePaginatedResponse(w, http.StatusOK, data, 500, params, "req-123", "v1")
	}
}

// BenchmarkGenerateRequestID tests request ID generation performance
func BenchmarkGenerateRequestID(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GenerateRequestID()
	}
}

// BenchmarkValidateShareCreateRequest tests validation performance
func BenchmarkValidateShareCreateRequest(b *testing.B) {
	request := &requests.ShareCreateRequest{
		Name:            "test-share",
		Comment:         "Test share for benchmarking",
		AllocatorMethod: "high-water",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateShareCreateRequest(request)
	}
}

// BenchmarkValidateVMCreateRequest tests VM validation performance
func BenchmarkValidateVMCreateRequest(b *testing.B) {
	request := &requests.VMCreateRequest{
		Name:   "test-vm",
		CPUs:   4,
		Memory: 8192,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateVMCreateRequest(request)
	}
}

// BenchmarkValidateDockerContainerCreateRequest tests Docker validation performance
func BenchmarkValidateDockerContainerCreateRequest(b *testing.B) {
	request := &requests.DockerContainerCreateRequest{
		Name:  "test-container",
		Image: "nginx:latest",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateDockerContainerCreateRequest(request)
	}
}

// BenchmarkJSONMarshalUnmarshal tests JSON processing performance
func BenchmarkJSONMarshalUnmarshal(b *testing.B) {
	data := responses.StandardResponse{
		Data: map[string]interface{}{
			"items": []string{"item1", "item2", "item3"},
			"count": 3,
		},
		Meta: &responses.ResponseMeta{
			RequestID: "req-123",
			Timestamp: time.Now(),
			Version:   "v1",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Marshal
		jsonData, err := json.Marshal(data)
		if err != nil {
			b.Fatal(err)
		}

		// Unmarshal
		var result responses.StandardResponse
		err = json.Unmarshal(jsonData, &result)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkConcurrentWriteJSON tests concurrent JSON writing performance
func BenchmarkConcurrentWriteJSON(b *testing.B) {
	data := map[string]interface{}{
		"message": "concurrent test",
		"id":      12345,
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			WriteJSON(w, http.StatusOK, data)
		}
	})
}

// BenchmarkConcurrentRequestIDGeneration tests concurrent request ID generation
func BenchmarkConcurrentRequestIDGeneration(b *testing.B) {
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = GenerateRequestID()
		}
	})
}

// BenchmarkResponseBufferPool tests response buffer performance
func BenchmarkResponseBufferPool(b *testing.B) {
	data := map[string]interface{}{
		"large_data": make([]string, 1000),
	}

	// Fill with test data
	for i := 0; i < 1000; i++ {
		data["large_data"].([]string)[i] = "test data item " + string(rune('0'+i%10))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		encoder := json.NewEncoder(&buf)
		err := encoder.Encode(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkHTTPResponseWriting tests complete HTTP response writing
func BenchmarkHTTPResponseWriting(b *testing.B) {
	data := map[string]interface{}{
		"status":  "success",
		"message": "Operation completed",
		"data": map[string]interface{}{
			"id":     123,
			"name":   "Test Item",
			"active": true,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		encoder := json.NewEncoder(w)
		err := encoder.Encode(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMemoryAllocation tests memory allocation patterns
func BenchmarkMemoryAllocation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate typical API response creation
		response := responses.StandardResponse{
			Data: map[string]interface{}{
				"items": make([]map[string]interface{}, 10),
			},
			Meta: &responses.ResponseMeta{
				RequestID: GenerateRequestID(),
				Timestamp: time.Now(),
				Version:   "v1",
			},
		}

		// Fill items
		items := response.Data.(map[string]interface{})["items"].([]map[string]interface{})
		for j := 0; j < 10; j++ {
			items[j] = map[string]interface{}{
				"id":   j,
				"name": "Item " + string(rune('0'+j%10)),
			}
		}
	}
}
