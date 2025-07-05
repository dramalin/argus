package benchmarks

import (
	"bytes"
	"strings"
	"sync"
	"testing"
	"text/template"

	"argus/internal/models"
	"argus/internal/services"
	"argus/internal/utils"
)

// BenchmarkBytesBufferPool benchmarks the performance of pooled vs non-pooled bytes.Buffer
func BenchmarkBytesBufferPool(b *testing.B) {
	b.Run("WithPool", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := utils.GetBytesBuffer()
			buf.WriteString("test string for benchmarking")
			buf.WriteString(" additional content")
			_ = buf.String()
			utils.PutBytesBuffer(buf)
		}
	})

	b.Run("WithoutPool", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			buf.WriteString("test string for benchmarking")
			buf.WriteString(" additional content")
			_ = buf.String()
		}
	})
}

// BenchmarkStringsBuilderPool benchmarks the performance of pooled vs non-pooled strings.Builder
func BenchmarkStringsBuilderPool(b *testing.B) {
	b.Run("WithPool", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			sb := utils.GetStringsBuilder()
			sb.WriteString("test string for benchmarking")
			sb.WriteString(" additional content")
			_ = sb.String()
			utils.PutStringsBuilder(sb)
		}
	})

	b.Run("WithoutPool", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var sb strings.Builder
			sb.WriteString("test string for benchmarking")
			sb.WriteString(" additional content")
			_ = sb.String()
		}
	})
}

// BenchmarkStringSlicePool benchmarks the performance of pooled vs non-pooled string slices
func BenchmarkStringSlicePool(b *testing.B) {
	b.Run("WithPool", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			slice := utils.GetStringSlice()
			slice = append(slice, "item1", "item2", "item3", "item4", "item5")
			_ = len(slice)
			utils.PutStringSlice(slice)
		}
	})

	b.Run("WithoutPool", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			slice := make([]string, 0, 16)
			slice = append(slice, "item1", "item2", "item3", "item4", "item5")
			_ = len(slice)
		}
	})
}

// BenchmarkMapStringStringPool benchmarks the performance of pooled vs non-pooled maps
func BenchmarkMapStringStringPool(b *testing.B) {
	b.Run("WithPool", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			m := utils.GetMapStringString()
			m["key1"] = "value1"
			m["key2"] = "value2"
			m["key3"] = "value3"
			_ = len(m)
			utils.PutMapStringString(m)
		}
	})

	b.Run("WithoutPool", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			m := make(map[string]string, 16)
			m["key1"] = "value1"
			m["key2"] = "value2"
			m["key3"] = "value3"
			_ = len(m)
		}
	})
}

// BenchmarkTemplateRenderingWithPools benchmarks template rendering with and without pooled buffers
func BenchmarkTemplateRenderingWithPools(b *testing.B) {
	tmpl, err := template.New("test").Parse("Alert: {{.Alert.Name}} - Status: {{.NewState}} - Value: {{.CurrentValue}}")
	if err != nil {
		b.Fatal(err)
	}

	event := createBenchmarkAlertEvent()

	b.Run("WithPool", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := utils.GetBytesBuffer()
			err := tmpl.Execute(buf, event)
			if err != nil {
				b.Fatal(err)
			}
			_ = buf.String()
			utils.PutBytesBuffer(buf)
		}
	})

	b.Run("WithoutPool", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			err := tmpl.Execute(&buf, event)
			if err != nil {
				b.Fatal(err)
			}
			_ = buf.String()
		}
	})
}

// BenchmarkConcurrentPoolUsage benchmarks concurrent usage of memory pools
func BenchmarkConcurrentPoolUsage(b *testing.B) {
	b.Run("BytesBufferPool", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				buf := utils.GetBytesBuffer()
				buf.WriteString("concurrent test data")
				_ = buf.String()
				utils.PutBytesBuffer(buf)
			}
		})
	})

	b.Run("StringsBuilderPool", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				sb := utils.GetStringsBuilder()
				sb.WriteString("concurrent test data")
				_ = sb.String()
				utils.PutStringsBuilder(sb)
			}
		})
	})
}

// BenchmarkNotificationWithPools benchmarks notification processing with pooled buffers
func BenchmarkNotificationWithPools(b *testing.B) {
	// Create notifier with pre-compiled templates
	notifier := services.NewNotifier(nil)

	// Register mock channel
	mockChannel := &memoryPoolMockChannel{}
	notifier.RegisterChannel(mockChannel)

	event := createBenchmarkAlertEvent()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		notifier.ProcessEvent(event)
	}
}

// BenchmarkMemoryPoolOverhead benchmarks the overhead of pool operations
func BenchmarkMemoryPoolOverhead(b *testing.B) {
	b.Run("GetPutBytesBuffer", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := utils.GetBytesBuffer()
			utils.PutBytesBuffer(buf)
		}
	})

	b.Run("GetPutStringsBuilder", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			sb := utils.GetStringsBuilder()
			utils.PutStringsBuilder(sb)
		}
	})

	b.Run("DirectAllocation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = &bytes.Buffer{}
			_ = &strings.Builder{}
		}
	})
}

// BenchmarkPoolSizeLimit benchmarks the behavior when pools hit size limits
func BenchmarkPoolSizeLimit(b *testing.B) {
	// Create large buffers that exceed the pool size limit
	b.Run("LargeBytesBuffer", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := utils.GetBytesBuffer()
			// Write data that exceeds the 64KB limit
			for j := 0; j < 1000; j++ {
				buf.WriteString("This is a long string to exceed the pool size limit and test the behavior when buffers are too large to be pooled")
			}
			_ = buf.String()
			utils.PutBytesBuffer(buf) // Should not be pooled due to size
		}
	})

	b.Run("LargeStringSlice", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			slice := utils.GetStringSlice()
			// Add items that exceed the 1024 capacity limit
			for j := 0; j < 1200; j++ {
				slice = append(slice, "item")
			}
			_ = len(slice)
			utils.PutStringSlice(slice) // Should not be pooled due to size
		}
	})
}

// Utility functions for benchmarks

type memoryPoolMockChannel struct {
	mu        sync.Mutex
	sendCount int
}

func (m *memoryPoolMockChannel) Send(event models.AlertEvent, subject, body string) error {
	m.mu.Lock()
	m.sendCount++
	m.mu.Unlock()
	return nil
}

func (m *memoryPoolMockChannel) Type() models.NotificationType {
	return models.NotificationInApp
}

func (m *memoryPoolMockChannel) Name() string {
	return "Memory Pool Mock Channel"
}
