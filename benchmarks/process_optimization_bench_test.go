package benchmarks

import (
	"reflect"
	"testing"
	"time"

	"argus/internal/metrics"
)

// BenchmarkProcessFiltering benchmarks the performance of process filtering
func BenchmarkProcessFiltering(b *testing.B) {
	// Create a collector with test data
	collector := metrics.NewCollector(metrics.DefaultConfig())

	// Simulate process metrics with various CPU and memory usage
	processes := make([]metrics.ProcessInfo, 1000)
	for i := 0; i < 1000; i++ {
		processes[i] = metrics.ProcessInfo{
			PID:        int32(i + 1),
			Name:       generateProcessName(i),
			CPUPercent: float64(i%100) + 0.5, // 0.5 to 99.5
			MemPercent: float32(i%50) + 0.1,  // 0.1 to 49.1
		}
	}

	// Set up test metrics
	testMetrics := &metrics.ProcessMetrics{
		Processes: processes,
		UpdatedAt: time.Now(),
	}

	// Use reflection or a test helper to set the metrics
	setTestProcessMetrics(collector, testMetrics)

	b.Run("NoFiltering", func(b *testing.B) {
		filter := metrics.ProcessFilter{
			Limit:     50,
			Offset:    0,
			SortBy:    "cpu",
			SortOrder: "desc",
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, err := collector.GetOptimizedProcessMetrics(filter)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("CPUFiltering", func(b *testing.B) {
		filter := metrics.ProcessFilter{
			Limit:     50,
			Offset:    0,
			SortBy:    "cpu",
			SortOrder: "desc",
			MinCPU:    50.0, // Filter processes with CPU > 50%
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, err := collector.GetOptimizedProcessMetrics(filter)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("MemoryFiltering", func(b *testing.B) {
		filter := metrics.ProcessFilter{
			Limit:     50,
			Offset:    0,
			SortBy:    "memory",
			SortOrder: "desc",
			MinMemory: 25.0, // Filter processes with memory > 25%
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, err := collector.GetOptimizedProcessMetrics(filter)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("NameFiltering", func(b *testing.B) {
		filter := metrics.ProcessFilter{
			Limit:        50,
			Offset:       0,
			SortBy:       "name",
			SortOrder:    "asc",
			NameContains: "process", // Filter processes containing "process"
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, err := collector.GetOptimizedProcessMetrics(filter)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("CombinedFiltering", func(b *testing.B) {
		filter := metrics.ProcessFilter{
			Limit:        50,
			Offset:       0,
			SortBy:       "cpu",
			SortOrder:    "desc",
			MinCPU:       30.0,
			MinMemory:    15.0,
			NameContains: "proc",
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, err := collector.GetOptimizedProcessMetrics(filter)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkProcessPagination benchmarks pagination performance
func BenchmarkProcessPagination(b *testing.B) {
	collector := metrics.NewCollector(metrics.DefaultConfig())

	// Create larger dataset for pagination testing
	processes := make([]metrics.ProcessInfo, 5000)
	for i := 0; i < 5000; i++ {
		processes[i] = metrics.ProcessInfo{
			PID:        int32(i + 1),
			Name:       generateProcessName(i),
			CPUPercent: float64(i%100) + 0.5,
			MemPercent: float32(i%50) + 0.1,
		}
	}

	testMetrics := &metrics.ProcessMetrics{
		Processes: processes,
		UpdatedAt: time.Now(),
	}
	setTestProcessMetrics(collector, testMetrics)

	b.Run("FirstPage", func(b *testing.B) {
		filter := metrics.ProcessFilter{
			Limit:     100,
			Offset:    0,
			SortBy:    "cpu",
			SortOrder: "desc",
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, err := collector.GetOptimizedProcessMetrics(filter)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("MiddlePage", func(b *testing.B) {
		filter := metrics.ProcessFilter{
			Limit:     100,
			Offset:    2000, // Middle of dataset
			SortBy:    "cpu",
			SortOrder: "desc",
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, err := collector.GetOptimizedProcessMetrics(filter)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("LastPage", func(b *testing.B) {
		filter := metrics.ProcessFilter{
			Limit:     100,
			Offset:    4900, // Near end of dataset
			SortBy:    "cpu",
			SortOrder: "desc",
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, err := collector.GetOptimizedProcessMetrics(filter)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkTopNSelection benchmarks heap-based top-N selection vs full sorting
func BenchmarkTopNSelection(b *testing.B) {
	collector := metrics.NewCollector(metrics.DefaultConfig())

	// Create large dataset
	processes := make([]metrics.ProcessInfo, 10000)
	for i := 0; i < 10000; i++ {
		processes[i] = metrics.ProcessInfo{
			PID:        int32(i + 1),
			Name:       generateProcessName(i),
			CPUPercent: float64(i%100) + 0.5,
			MemPercent: float32(i%50) + 0.1,
		}
	}

	testMetrics := &metrics.ProcessMetrics{
		Processes: processes,
		UpdatedAt: time.Now(),
	}
	setTestProcessMetrics(collector, testMetrics)

	b.Run("Top10_HeapBased", func(b *testing.B) {
		filter := metrics.ProcessFilter{
			TopN:      10,
			SortBy:    "cpu",
			SortOrder: "desc",
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, err := collector.GetOptimizedProcessMetrics(filter)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Top10_FullSort", func(b *testing.B) {
		filter := metrics.ProcessFilter{
			Limit:     10,
			Offset:    0,
			SortBy:    "cpu",
			SortOrder: "desc",
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, err := collector.GetOptimizedProcessMetrics(filter)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Top100_HeapBased", func(b *testing.B) {
		filter := metrics.ProcessFilter{
			TopN:      100,
			SortBy:    "cpu",
			SortOrder: "desc",
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, err := collector.GetOptimizedProcessMetrics(filter)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Top100_FullSort", func(b *testing.B) {
		filter := metrics.ProcessFilter{
			Limit:     100,
			Offset:    0,
			SortBy:    "cpu",
			SortOrder: "desc",
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, err := collector.GetOptimizedProcessMetrics(filter)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkProcessSorting benchmarks different sorting fields
func BenchmarkProcessSorting(b *testing.B) {
	collector := metrics.NewCollector(metrics.DefaultConfig())

	processes := make([]metrics.ProcessInfo, 1000)
	for i := 0; i < 1000; i++ {
		processes[i] = metrics.ProcessInfo{
			PID:        int32(i + 1),
			Name:       generateProcessName(i),
			CPUPercent: float64(i%100) + 0.5,
			MemPercent: float32(i%50) + 0.1,
		}
	}

	testMetrics := &metrics.ProcessMetrics{
		Processes: processes,
		UpdatedAt: time.Now(),
	}
	setTestProcessMetrics(collector, testMetrics)

	sortFields := []string{"cpu", "memory", "name", "pid"}
	sortOrders := []string{"asc", "desc"}

	for _, field := range sortFields {
		for _, order := range sortOrders {
			b.Run("Sort_"+field+"_"+order, func(b *testing.B) {
				filter := metrics.ProcessFilter{
					Limit:     100,
					Offset:    0,
					SortBy:    field,
					SortOrder: order,
				}

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					_, _, err := collector.GetOptimizedProcessMetrics(filter)
					if err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}

// BenchmarkConcurrentProcessAccess benchmarks concurrent access to optimized process metrics
func BenchmarkConcurrentProcessAccess(b *testing.B) {
	collector := metrics.NewCollector(metrics.DefaultConfig())

	processes := make([]metrics.ProcessInfo, 2000)
	for i := 0; i < 2000; i++ {
		processes[i] = metrics.ProcessInfo{
			PID:        int32(i + 1),
			Name:       generateProcessName(i),
			CPUPercent: float64(i%100) + 0.5,
			MemPercent: float32(i%50) + 0.1,
		}
	}

	testMetrics := &metrics.ProcessMetrics{
		Processes: processes,
		UpdatedAt: time.Now(),
	}
	setTestProcessMetrics(collector, testMetrics)

	b.Run("ConcurrentAccess", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				filter := metrics.ProcessFilter{
					Limit:     50,
					Offset:    0,
					SortBy:    "cpu",
					SortOrder: "desc",
					MinCPU:    10.0,
				}

				_, _, err := collector.GetOptimizedProcessMetrics(filter)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	})
}

// Helper functions

// generateProcessName creates a test process name
func generateProcessName(i int) string {
	names := []string{
		"systemd", "kthreadd", "ksoftirqd", "migration", "rcu_gp", "rcu_par_gp",
		"kworker", "mm_percpu_wq", "ksoftirqd", "migration", "rcu_gp", "rcu_par_gp",
		"netns", "kcompactd0", "khugepaged", "crypto", "kintegrityd", "kblockd",
		"ata_sff", "md", "edac-poller", "devfreq_wq", "watchdogd", "kswapd0",
		"kthrotld", "irq", "kmpath_rdacd", "kaluad", "kpsmoused", "ipv6_addrconf",
		"kstrp", "charger_manager", "scsi_eh", "scsi_tmf", "usb-storage",
		"process_worker", "background_task", "data_processor", "file_manager",
		"network_handler", "cache_cleaner", "log_rotator", "metric_collector",
	}

	return names[i%len(names)] + "_" + string(rune('a'+i%26))
}

// setTestProcessMetrics is a helper to set test metrics using reflection
func setTestProcessMetrics(collector *metrics.Collector, testMetrics *metrics.ProcessMetrics) {
	// Use reflection to access the private fields
	v := reflect.ValueOf(collector).Elem()

	// Set the process metrics
	processMetricsField := v.FieldByName("processMetrics")
	if processMetricsField.IsValid() && processMetricsField.CanSet() {
		processMetricsField.Set(reflect.ValueOf(testMetrics))
	}
}
