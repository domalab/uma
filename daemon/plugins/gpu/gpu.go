package gpu

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/domalab/uma/daemon/dto"
	"github.com/domalab/uma/daemon/lib"
	"github.com/domalab/uma/daemon/logger"
)

// GPUMonitor provides GPU monitoring capabilities
type GPUMonitor struct {
	gpus []GPUInfo
}

// GPUInfo represents information about a GPU
type GPUInfo struct {
	Index           int     `json:"index"`
	Name            string  `json:"name"`
	UUID            string  `json:"uuid"`
	Driver          string  `json:"driver"`
	Temperature     int     `json:"temperature"`
	PowerDraw       float64 `json:"power_draw_watts"`
	PowerLimit      float64 `json:"power_limit_watts"`
	UtilizationGPU  int     `json:"utilization_gpu_percent"`
	UtilizationMem  int     `json:"utilization_memory_percent"`
	MemoryTotal     uint64  `json:"memory_total_bytes"`
	MemoryUsed      uint64  `json:"memory_used_bytes"`
	MemoryFree      uint64  `json:"memory_free_bytes"`
	FanSpeed        int     `json:"fan_speed_percent"`
	ClockCore       int     `json:"clock_core_mhz"`
	ClockMemory     int     `json:"clock_memory_mhz"`
	Status          string  `json:"status"`
}

// NvidiaSMIOutput represents nvidia-smi JSON output structure
type NvidiaSMIOutput struct {
	GPUs []struct {
		Index       string `json:"index"`
		Name        string `json:"name"`
		UUID        string `json:"uuid"`
		Temperature struct {
			GPU string `json:"gpu_temp"`
		} `json:"temperature"`
		Power struct {
			Draw  string `json:"power_draw"`
			Limit string `json:"power_limit"`
		} `json:"power"`
		Utilization struct {
			GPU    string `json:"gpu_util"`
			Memory string `json:"memory_util"`
		} `json:"utilization"`
		Memory struct {
			Total string `json:"total"`
			Used  string `json:"used"`
			Free  string `json:"free"`
		} `json:"fb_memory_usage"`
		FanSpeed string `json:"fan_speed"`
		Clocks   struct {
			Core   string `json:"graphics_clock"`
			Memory string `json:"mem_clock"`
		} `json:"clocks"`
	} `json:"gpus"`
}

// NewGPUMonitor creates a new GPU monitor
func NewGPUMonitor() *GPUMonitor {
	return &GPUMonitor{
		gpus: make([]GPUInfo, 0),
	}
}

// GetGPUInfo returns information about all GPUs
func (g *GPUMonitor) GetGPUInfo() ([]GPUInfo, error) {
	gpus := make([]GPUInfo, 0)

	// Try NVIDIA GPUs first
	nvidiaGPUs, err := g.getNvidiaGPUs()
	if err == nil && len(nvidiaGPUs) > 0 {
		gpus = append(gpus, nvidiaGPUs...)
	}

	// Try AMD GPUs
	amdGPUs, err := g.getAMDGPUs()
	if err == nil && len(amdGPUs) > 0 {
		gpus = append(gpus, amdGPUs...)
	}

	// Try Intel GPUs
	intelGPUs, err := g.getIntelGPUs()
	if err == nil && len(intelGPUs) > 0 {
		gpus = append(gpus, intelGPUs...)
	}

	g.gpus = gpus
	return gpus, nil
}

// GetGPUSamples returns GPU information as DTO samples for compatibility
func (g *GPUMonitor) GetGPUSamples() []dto.Sample {
	samples := make([]dto.Sample, 0)

	gpus, err := g.GetGPUInfo()
	if err != nil {
		logger.Yellow("Failed to get GPU info: %v", err)
		return samples
	}

	for i, gpu := range gpus {
		prefix := fmt.Sprintf("GPU%d", i)

		samples = append(samples, dto.Sample{
			Key:       fmt.Sprintf("%s_NAME", prefix),
			Value:     gpu.Name,
			Unit:      "",
			Condition: "neutral",
		})

		if gpu.Temperature > 0 {
			samples = append(samples, dto.Sample{
				Key:       fmt.Sprintf("%s_TEMP", prefix),
				Value:     fmt.Sprintf("%d", gpu.Temperature),
				Unit:      "°C",
				Condition: g.getTempCondition(gpu.Temperature),
			})
		}

		if gpu.UtilizationGPU >= 0 {
			samples = append(samples, dto.Sample{
				Key:       fmt.Sprintf("%s_UTIL", prefix),
				Value:     fmt.Sprintf("%d", gpu.UtilizationGPU),
				Unit:      "%",
				Condition: g.getUtilCondition(gpu.UtilizationGPU),
			})
		}

		if gpu.PowerDraw > 0 {
			samples = append(samples, dto.Sample{
				Key:       fmt.Sprintf("%s_POWER", prefix),
				Value:     fmt.Sprintf("%.1f", gpu.PowerDraw),
				Unit:      "W",
				Condition: g.getPowerCondition(gpu.PowerDraw, gpu.PowerLimit),
			})
		}

		if gpu.MemoryTotal > 0 {
			memUsedPercent := float64(gpu.MemoryUsed) / float64(gpu.MemoryTotal) * 100
			samples = append(samples, dto.Sample{
				Key:       fmt.Sprintf("%s_MEM", prefix),
				Value:     fmt.Sprintf("%.1f", memUsedPercent),
				Unit:      "%",
				Condition: g.getMemCondition(memUsedPercent),
			})
		}
	}

	return samples
}

// getNvidiaGPUs gets information about NVIDIA GPUs using nvidia-smi
func (g *GPUMonitor) getNvidiaGPUs() ([]GPUInfo, error) {
	gpus := make([]GPUInfo, 0)

	// Check if nvidia-smi is available
	output := lib.GetCmdOutput("which", "nvidia-smi")
	if len(output) == 0 {
		return gpus, fmt.Errorf("nvidia-smi not found")
	}

	// Get GPU information in JSON format
	output = lib.GetCmdOutput("nvidia-smi", "--query-gpu=index,name,uuid,temperature.gpu,power.draw,power.limit,utilization.gpu,utilization.memory,memory.total,memory.used,memory.free,fan.speed,clocks.gr,clocks.mem", "--format=csv,noheader,nounits")

	for _, line := range output {
		if strings.TrimSpace(line) == "" {
			continue
		}

		gpu, err := g.parseNvidiaCSVLine(line)
		if err != nil {
			logger.Yellow("Failed to parse NVIDIA GPU line: %v", err)
			continue
		}

		gpu.Driver = "nvidia"
		gpu.Status = "active"
		gpus = append(gpus, gpu)
	}

	return gpus, nil
}

// parseNvidiaCSVLine parses a CSV line from nvidia-smi output
func (g *GPUMonitor) parseNvidiaCSVLine(line string) (GPUInfo, error) {
	var gpu GPUInfo

	fields := strings.Split(line, ",")
	if len(fields) < 14 {
		return gpu, fmt.Errorf("insufficient fields in CSV line")
	}

	// Trim whitespace from all fields
	for i := range fields {
		fields[i] = strings.TrimSpace(fields[i])
	}

	var err error

	// Parse index
	if gpu.Index, err = strconv.Atoi(fields[0]); err != nil {
		return gpu, err
	}

	// Parse name
	gpu.Name = fields[1]
	gpu.UUID = fields[2]

	// Parse temperature
	if fields[3] != "[Not Supported]" && fields[3] != "" {
		if gpu.Temperature, err = strconv.Atoi(fields[3]); err != nil {
			gpu.Temperature = 0
		}
	}

	// Parse power draw
	if fields[4] != "[Not Supported]" && fields[4] != "" {
		if gpu.PowerDraw, err = strconv.ParseFloat(fields[4], 64); err != nil {
			gpu.PowerDraw = 0
		}
	}

	// Parse power limit
	if fields[5] != "[Not Supported]" && fields[5] != "" {
		if gpu.PowerLimit, err = strconv.ParseFloat(fields[5], 64); err != nil {
			gpu.PowerLimit = 0
		}
	}

	// Parse GPU utilization
	if fields[6] != "[Not Supported]" && fields[6] != "" {
		if gpu.UtilizationGPU, err = strconv.Atoi(fields[6]); err != nil {
			gpu.UtilizationGPU = -1
		}
	}

	// Parse memory utilization
	if fields[7] != "[Not Supported]" && fields[7] != "" {
		if gpu.UtilizationMem, err = strconv.Atoi(fields[7]); err != nil {
			gpu.UtilizationMem = -1
		}
	}

	// Parse memory total (in MiB, convert to bytes)
	if fields[8] != "[Not Supported]" && fields[8] != "" {
		if memTotal, err := strconv.ParseUint(fields[8], 10, 64); err == nil {
			gpu.MemoryTotal = memTotal * 1024 * 1024
		}
	}

	// Parse memory used (in MiB, convert to bytes)
	if fields[9] != "[Not Supported]" && fields[9] != "" {
		if memUsed, err := strconv.ParseUint(fields[9], 10, 64); err == nil {
			gpu.MemoryUsed = memUsed * 1024 * 1024
		}
	}

	// Parse memory free (in MiB, convert to bytes)
	if fields[10] != "[Not Supported]" && fields[10] != "" {
		if memFree, err := strconv.ParseUint(fields[10], 10, 64); err == nil {
			gpu.MemoryFree = memFree * 1024 * 1024
		}
	}

	// Parse fan speed
	if fields[11] != "[Not Supported]" && fields[11] != "" {
		if gpu.FanSpeed, err = strconv.Atoi(fields[11]); err != nil {
			gpu.FanSpeed = 0
		}
	}

	// Parse core clock
	if fields[12] != "[Not Supported]" && fields[12] != "" {
		if gpu.ClockCore, err = strconv.Atoi(fields[12]); err != nil {
			gpu.ClockCore = 0
		}
	}

	// Parse memory clock
	if fields[13] != "[Not Supported]" && fields[13] != "" {
		if gpu.ClockMemory, err = strconv.Atoi(fields[13]); err != nil {
			gpu.ClockMemory = 0
		}
	}

	return gpu, nil
}

// getAMDGPUs gets information about AMD GPUs using rocm-smi
func (g *GPUMonitor) getAMDGPUs() ([]GPUInfo, error) {
	gpus := make([]GPUInfo, 0)

	// Check if rocm-smi is available
	output := lib.GetCmdOutput("which", "rocm-smi")
	if len(output) == 0 {
		return gpus, fmt.Errorf("rocm-smi not found")
	}

	// Get basic GPU information
	output = lib.GetCmdOutput("rocm-smi", "--showid", "--showproductname")
	
	// Parse AMD GPU output (simplified implementation)
	for i, line := range output {
		if strings.Contains(line, "GPU[") {
			gpu := GPUInfo{
				Index:  i,
				Driver: "amdgpu",
				Status: "active",
			}

			// Extract GPU name from the line
			if parts := strings.Split(line, ":"); len(parts) > 1 {
				gpu.Name = strings.TrimSpace(parts[1])
			}

			gpus = append(gpus, gpu)
		}
	}

	return gpus, nil
}

// getIntelGPUs gets information about Intel GPUs
func (g *GPUMonitor) getIntelGPUs() ([]GPUInfo, error) {
	gpus := make([]GPUInfo, 0)

	// Check for Intel GPU using lspci
	output := lib.GetCmdOutput("lspci", "-d", "8086:", "-v")
	
	index := 0
	for _, line := range output {
		if strings.Contains(line, "VGA compatible controller") && strings.Contains(line, "Intel") {
			gpu := GPUInfo{
				Index:  index,
				Driver: "i915",
				Status: "active",
			}

			// Extract GPU name
			if parts := strings.Split(line, ":"); len(parts) > 2 {
				gpu.Name = strings.TrimSpace(parts[2])
			}

			gpus = append(gpus, gpu)
			index++
		}
	}

	return gpus, nil
}

// Condition helper functions
func (g *GPUMonitor) getTempCondition(temp int) string {
	if temp >= 85 {
		return "critical"
	} else if temp >= 75 {
		return "warning"
	}
	return "normal"
}

func (g *GPUMonitor) getUtilCondition(util int) string {
	if util >= 95 {
		return "critical"
	} else if util >= 80 {
		return "warning"
	}
	return "normal"
}

func (g *GPUMonitor) getPowerCondition(draw, limit float64) string {
	if limit > 0 {
		percent := (draw / limit) * 100
		if percent >= 95 {
			return "critical"
		} else if percent >= 85 {
			return "warning"
		}
	}
	return "normal"
}

func (g *GPUMonitor) getMemCondition(percent float64) string {
	if percent >= 90 {
		return "critical"
	} else if percent >= 80 {
		return "warning"
	}
	return "normal"
}
