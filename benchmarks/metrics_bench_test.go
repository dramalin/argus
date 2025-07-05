package benchmarks

import (
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

// BenchmarkCPUCollection benchmarks CPU metrics collection
func BenchmarkCPUCollection(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		loadAvg, err := load.Avg()
		if err != nil {
			b.Fatal(err)
		}
		_ = loadAvg

		cpuPercent, err := cpu.Percent(time.Second, false)
		if err != nil {
			b.Fatal(err)
		}
		_ = cpuPercent
	}
}

// BenchmarkMemoryCollection benchmarks memory metrics collection
func BenchmarkMemoryCollection(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm, err := mem.VirtualMemory()
		if err != nil {
			b.Fatal(err)
		}
		_ = vm
	}
}

// BenchmarkNetworkCollection benchmarks network metrics collection
func BenchmarkNetworkCollection(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ioCounters, err := net.IOCounters(false)
		if err != nil {
			b.Fatal(err)
		}
		_ = ioCounters
	}
}

// BenchmarkProcessCollection benchmarks process metrics collection
func BenchmarkProcessCollection(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		procs, err := process.Processes()
		if err != nil {
			b.Fatal(err)
		}

		// Simulate the processing done in the actual endpoint
		type processInfo struct {
			pid        int32
			name       string
			cpuPercent float64
			memPercent float32
		}

		var processes []processInfo
		for _, p := range procs {
			name, err := p.Name()
			if err != nil {
				continue
			}

			cpuP, err := p.CPUPercent()
			if err != nil {
				cpuP = 0.0
			}

			memP, err := p.MemoryPercent()
			if err != nil {
				memP = 0.0
			}

			processes = append(processes, processInfo{
				pid:        p.Pid,
				name:       name,
				cpuPercent: cpuP,
				memPercent: memP,
			})
		}
		_ = processes
	}
}

// BenchmarkGinContextCreation benchmarks Gin context creation overhead
func BenchmarkGinContextCreation(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c, _ := gin.CreateTestContext(nil)
		_ = c
	}
	_ = router
}

// BenchmarkConcurrentMetricsCollection benchmarks concurrent metrics collection
func BenchmarkConcurrentMetricsCollection(b *testing.B) {
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Simulate concurrent access to different metrics
			go func() {
				vm, _ := mem.VirtualMemory()
				_ = vm
			}()

			go func() {
				loadAvg, _ := load.Avg()
				_ = loadAvg
			}()

			go func() {
				ioCounters, _ := net.IOCounters(false)
				_ = ioCounters
			}()
		}
	})
}
