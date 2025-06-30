package collectors

import (
	"fmt"
	"time"
)

// SystemReader provides direct system metrics reading
type SystemReader struct {
	cpuReader     *CPUReader
	memoryReader  *MemoryReader
	networkReader *NetworkReader
	storageReader *StorageReader
	dockerReader  *DockerReader
}

// NewSystemReader creates a new system reader
func NewSystemReader() *SystemReader {
	return &SystemReader{
		cpuReader:     NewCPUReader(),
		memoryReader:  NewMemoryReader(),
		networkReader: NewNetworkReader(),
		storageReader: NewStorageReader(),
		dockerReader:  NewDockerReader(),
	}
}

// CollectSystemMetrics collects comprehensive system metrics
func (sr *SystemReader) CollectSystemMetrics() (*SystemMetrics, error) {
	start := time.Now()

	// Collect CPU metrics
	cpuPercent, load1m, err := sr.cpuReader.ReadCPUStats()
	if err != nil {
		cpuPercent, load1m = 0, 0 // Use defaults on error
	}

	// Collect memory metrics
	memUsed, memTotal, err := sr.memoryReader.ReadMemoryStats()
	if err != nil {
		memUsed, memTotal = 0, 0 // Use defaults on error
	}

	// Collect network metrics
	netRx, netTx, err := sr.networkReader.ReadNetworkStats()
	if err != nil {
		netRx, netTx = 0, 0 // Use defaults on error
	}

	memPercent := float64(0)
	if memTotal > 0 {
		memPercent = float64(memUsed) / float64(memTotal) * 100
	}

	metrics := &SystemMetrics{
		Timestamp:     time.Now().Unix(),
		CPUPercent:    cpuPercent,
		Load1m:        load1m,
		MemoryPercent: memPercent,
		MemoryUsed:    memUsed,
		MemoryTotal:   memTotal,
		NetworkRx:     netRx,
		NetworkTx:     netTx,
	}

	duration := time.Since(start)
	if duration > 50*time.Millisecond {
		fmt.Printf("System metrics collection took %v\n", duration)
	}

	return metrics, nil
}

// CollectContainerMetrics collects Docker container metrics
func (sr *SystemReader) CollectContainerMetrics() (*ContainerMetrics, error) {
	start := time.Now()

	containers, summary, err := sr.dockerReader.ReadContainerStats()
	if err != nil {
		// Return empty metrics on error
		return &ContainerMetrics{
			Timestamp:  time.Now().Unix(),
			Containers: []ContainerStats{},
			Summary:    ContainerSummary{Total: 0, Running: 0, Stopped: 0},
		}, nil
	}

	metrics := &ContainerMetrics{
		Timestamp:  time.Now().Unix(),
		Containers: containers,
		Summary:    summary,
	}

	duration := time.Since(start)
	if duration > 100*time.Millisecond {
		fmt.Printf("Container metrics collection took %v\n", duration)
	}

	return metrics, nil
}

// CollectStorageMetrics collects storage usage metrics
func (sr *SystemReader) CollectStorageMetrics() (*StorageMetrics, error) {
	start := time.Now()

	arrayUsage, cacheUsage, dockerUsage, err := sr.storageReader.ReadStorageUsage()
	if err != nil {
		// Return default metrics on error
		defaultUsage := StorageUsage{Total: 0, Used: 0, Available: 0, Percent: 0}
		return &StorageMetrics{
			Timestamp:   time.Now().Unix(),
			ArrayStatus: "Unknown",
			ArrayUsage:  defaultUsage,
			CacheUsage:  defaultUsage,
			DockerUsage: defaultUsage,
		}, nil
	}

	metrics := &StorageMetrics{
		Timestamp:   time.Now().Unix(),
		ArrayStatus: "Started", // This would be read from actual array status
		ArrayUsage:  arrayUsage,
		CacheUsage:  cacheUsage,
		DockerUsage: dockerUsage,
	}

	duration := time.Since(start)
	if duration > 200*time.Millisecond {
		fmt.Printf("Storage metrics collection took %v\n", duration)
	}

	return metrics, nil
}
