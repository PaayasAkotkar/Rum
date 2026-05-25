package dog

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jaypipes/ghw"
)

// SystemMetrics tracks CPU, GPU, memory, and thermal metrics
type SystemMetrics struct {
	mu sync.RWMutex

	// CPU metrics
	CPUName        string
	CPUUsage       float64 // Percentage 0-100
	CPUCores       int
	CPUTemp        float64 // Celsius
	GoroutineCount int

	// Memory metrics
	MemoryUsage   uint64 // Bytes
	MemoryPercent float64
	AllocMB       float64
	TotalAllocMB  float64
	HeapAllocMB   float64
	HeapSysMB     float64

	// GPU metrics (if available)
	GPUUsage       float64 // Percentage 0-100
	GPUMemoryUsage uint64  // Bytes
	GPUTemp        float64 // Celsius
	GPUName        string

	// Thermal metrics
	CPUThrottled bool
	ThermalLevel string // "normal", "warning", "critical"

	// Process metrics
	PID         int
	RID         int // Rank ID
	StartTime   time.Time
	UpTime      time.Duration
	ThreadCount int

	// Sampling
	LastSample      time.Time
	SampleCount     atomic.Int64
	MaxMemorySeenMB float64
}

// NewSystemMetrics creates a new metrics tracker
func NewSystemMetrics(rankID int) *SystemMetrics {
	m := &SystemMetrics{
		PID:          os.Getpid(),
		RID:          rankID,
		StartTime:    time.Now(),
		CPUCores:     runtime.NumCPU(),
		ThermalLevel: "normal",
		LastSample:   time.Now(),
	}
	m.updateMetrics()
	return m
}

// updateMetrics collects current system metrics
func (sm *SystemMetrics) updateMetrics() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.LastSample = time.Now()
	sm.UpTime = time.Since(sm.StartTime)
	sm.SampleCount.Add(1)

	// Goroutine metrics
	sm.GoroutineCount = runtime.NumGoroutine()

	// Memory metrics (from runtime)
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	sm.AllocMB = float64(m.Alloc) / 1024 / 1024
	sm.TotalAllocMB = float64(m.TotalAlloc) / 1024 / 1024
	sm.HeapAllocMB = float64(m.HeapAlloc) / 1024 / 1024
	sm.HeapSysMB = float64(m.HeapSys) / 1024 / 1024
	sm.MemoryUsage = m.Alloc

	// Track max memory
	if sm.AllocMB > sm.MaxMemorySeenMB {
		sm.MaxMemorySeenMB = sm.AllocMB
	}

	// Estimate memory percentage (assuming 8GB total system memory)
	const systemMemoryMB = 8192.0
	sm.MemoryPercent = (sm.AllocMB / systemMemoryMB) * 100
	if sm.MemoryPercent > 100 {
		sm.MemoryPercent = 100
	}

	// CPU metrics (basic estimate based on goroutines)
	sm.CPUUsage = float64(sm.GoroutineCount) / float64(sm.CPUCores) * 10 // Rough estimate
	if sm.CPUUsage > 100 {
		sm.CPUUsage = 100
	}

	// Thermal metrics
	sm.checkThermalHealth()
	sm.estimateGPUMetrics()
}

// checkThermalHealth evaluates thermal status
func (sm *SystemMetrics) checkThermalHealth() {
	// Simulate CPU temperature based on usage and memory
	baseTemp := 45.0 // Base CPU temp in Celsius
	sm.CPUTemp = baseTemp + (sm.CPUUsage * 0.4) + (sm.MemoryPercent * 0.2)

	// Determine thermal level
	switch {
	case sm.CPUTemp >= 85:
		sm.ThermalLevel = "critical"
		sm.CPUThrottled = true
	case sm.CPUTemp >= 75:
		sm.ThermalLevel = "warning"
		sm.CPUThrottled = false
	default:
		sm.ThermalLevel = "normal"
		sm.CPUThrottled = false
	}
}

// estimateGPUMetrics provides GPU metrics if available
func (sm *SystemMetrics) estimateGPUMetrics() {
	// In production, this would use NVIDIA/AMD/Intel GPU APIs
	// For now, we'll provide a simulation based on CPU load

	// sm.GPUName = "Integrated Graphics (Intel/AMD)"
	g, err := ghw.GPU()
	if err != nil {
		fmt.Println("error getting gpu", err)
	}
	name := []string{}
	for _, r := range g.GraphicsCards {
		name = append(name, r.DeviceInfo.Vendor.Name)
	}
	c, err := ghw.CPU()
	cname := []string{}
	if err != nil {
		fmt.Println("error getting cpu", err)
	}
	for _, r := range c.Processors {
		cname = append(cname, r.Vendor)
	}
	sm.CPUName = strings.Join(cname, ",")
	sm.GPUName = strings.Join(name, ",")

	// GPU tends to correlate with CPU usage for integrated graphics
	sm.GPUUsage = sm.CPUUsage * 0.8

	// Estimate GPU memory (usually less than system RAM)
	gpuMemoryMB := sm.AllocMB * 0.3 // Rough estimate
	sm.GPUMemoryUsage = uint64(gpuMemoryMB * 1024 * 1024)

	// GPU temperature
	baseGPUTemp := 40.0
	sm.GPUTemp = baseGPUTemp + (sm.GPUUsage * 0.5)
}

// Sample updates all metrics
func (sm *SystemMetrics) Sample() {
	sm.updateMetrics()
}

// GetCPUHealth returns CPU health status
func (sm *SystemMetrics) GetCPUHealth() string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	switch {
	case sm.CPUUsage >= 90:
		return "critical"
	case sm.CPUUsage >= 70:
		return "warning"
	default:
		return "healthy"
	}
}

// GetMemoryHealth returns memory health status
func (sm *SystemMetrics) GetMemoryHealth() string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	switch {
	case sm.MemoryPercent >= 85:
		return "critical"
	case sm.MemoryPercent >= 70:
		return "warning"
	default:
		return "healthy"
	}
}

// GetGPUHealth returns GPU health status
func (sm *SystemMetrics) GetGPUHealth() string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	switch {
	case sm.GPUUsage >= 95:
		return "critical"
	case sm.GPUUsage >= 80:
		return "warning"
	default:
		return "healthy"
	}
}

// GetThermalHealth returns thermal health status
func (sm *SystemMetrics) GetThermalHealth() string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.ThermalLevel
}

// IsThrottled returns if CPU is thermally throttled
func (sm *SystemMetrics) IsThrottled() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.CPUThrottled
}

// GetSnapshot returns a thread-safe pointer to a copy of metrics
// This avoids copying the sync.RWMutex by value.
func (sm *SystemMetrics) GetSnapshot() *SystemMetrics {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	snapshot := &SystemMetrics{
		CPUName:         sm.CPUName,
		CPUUsage:        sm.CPUUsage,
		CPUCores:        sm.CPUCores,
		CPUTemp:         sm.CPUTemp,
		GoroutineCount:  sm.GoroutineCount,
		MemoryUsage:     sm.MemoryUsage,
		MemoryPercent:   sm.MemoryPercent,
		AllocMB:         sm.AllocMB,
		TotalAllocMB:    sm.TotalAllocMB,
		HeapAllocMB:     sm.HeapAllocMB,
		HeapSysMB:       sm.HeapSysMB,
		GPUUsage:        sm.GPUUsage,
		GPUMemoryUsage:  sm.GPUMemoryUsage,
		GPUTemp:         sm.GPUTemp,
		GPUName:         sm.GPUName,
		CPUThrottled:    sm.CPUThrottled,
		ThermalLevel:    sm.ThermalLevel,
		PID:             sm.PID,
		RID:             sm.RID,
		StartTime:       sm.StartTime,
		UpTime:          sm.UpTime,
		ThreadCount:     sm.ThreadCount,
		LastSample:      sm.LastSample,
		MaxMemorySeenMB: sm.MaxMemorySeenMB,
	}

	snapshot.SampleCount.Store(sm.SampleCount.Load())

	return snapshot
}

// String returns formatted metrics
func (sm *SystemMetrics) String() string {
	snapshot := sm.GetSnapshot()
	return fmt.Sprintf(`
╔════ System Metrics (PID: %d) ════╗
├─ CPU:
│  ├─ Name:      %s
│  ├─ Usage:     %.1f%%
│  ├─ Cores:     %d
│  ├─ Temp:      %.1f°C
│  └─ Throttled: %v
├─ Memory:
│  ├─ Alloc:     %.2f MB
│  ├─ MaxSeen:   %.2f MB
│  ├─ Percent:   %.1f%%
│  └─ Goroutines:%d
├─ GPU:
│  ├─ Name:      %s
│  ├─ Usage:     %.1f%%
│  ├─ Memory:    %.2f MB
│  └─ Temp:      %.1f°C
├─ Thermal: %s
└─ UpTime:  %v
`, snapshot.PID, snapshot.CPUName, snapshot.CPUUsage, snapshot.CPUCores, snapshot.CPUTemp, snapshot.CPUThrottled,
		snapshot.AllocMB, snapshot.MaxMemorySeenMB, snapshot.MemoryPercent, snapshot.GoroutineCount,
		snapshot.GPUName, snapshot.GPUUsage, float64(snapshot.GPUMemoryUsage)/1024/1024, snapshot.GPUTemp,
		snapshot.ThermalLevel, snapshot.UpTime)
}
