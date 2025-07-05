// Package utils provides centralized memory pools for reducing garbage collection pressure
package utils

import (
	"bytes"
	"strings"
	"sync"
)

// Global memory pools for reuse across the application

// BytesBufferPool provides reusable bytes.Buffer instances for template rendering and string building
var BytesBufferPool = sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{}
	},
}

// StringsBuilderPool provides reusable strings.Builder instances for string concatenation
var StringsBuilderPool = sync.Pool{
	New: func() interface{} {
		return &strings.Builder{}
	},
}

// StringSlicePool provides reusable string slices for temporary operations
var StringSlicePool = sync.Pool{
	New: func() interface{} {
		return make([]string, 0, 16) // Start with capacity of 16
	},
}

// IntSlicePool provides reusable int slices for temporary operations
var IntSlicePool = sync.Pool{
	New: func() interface{} {
		return make([]int, 0, 16) // Start with capacity of 16
	},
}

// MapStringStringPool provides reusable map[string]string instances
var MapStringStringPool = sync.Pool{
	New: func() interface{} {
		return make(map[string]string, 16) // Start with capacity of 16
	},
}

// GetBytesBuffer retrieves a buffer from the pool and resets it
func GetBytesBuffer() *bytes.Buffer {
	buf := BytesBufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

// PutBytesBuffer returns a buffer to the pool
func PutBytesBuffer(buf *bytes.Buffer) {
	// Only pool buffers that aren't too large to avoid memory leaks
	if buf.Cap() < 64*1024 { // 64KB limit
		BytesBufferPool.Put(buf)
	}
}

// GetStringsBuilder retrieves a strings.Builder from the pool and resets it
func GetStringsBuilder() *strings.Builder {
	sb := StringsBuilderPool.Get().(*strings.Builder)
	sb.Reset()
	return sb
}

// PutStringsBuilder returns a strings.Builder to the pool
func PutStringsBuilder(sb *strings.Builder) {
	// Only pool builders that aren't too large to avoid memory leaks
	if sb.Cap() < 64*1024 { // 64KB limit
		StringsBuilderPool.Put(sb)
	}
}

// GetStringSlice retrieves a string slice from the pool and resets it
func GetStringSlice() []string {
	slice := StringSlicePool.Get().([]string)
	return slice[:0] // Reset length but keep capacity
}

// PutStringSlice returns a string slice to the pool
func PutStringSlice(slice []string) {
	// Only pool slices that aren't too large to avoid memory leaks
	if cap(slice) < 1024 {
		StringSlicePool.Put(slice)
	}
}

// GetIntSlice retrieves an int slice from the pool and resets it
func GetIntSlice() []int {
	slice := IntSlicePool.Get().([]int)
	return slice[:0] // Reset length but keep capacity
}

// PutIntSlice returns an int slice to the pool
func PutIntSlice(slice []int) {
	// Only pool slices that aren't too large to avoid memory leaks
	if cap(slice) < 1024 {
		IntSlicePool.Put(slice)
	}
}

// GetMapStringString retrieves a map[string]string from the pool and clears it
func GetMapStringString() map[string]string {
	m := MapStringStringPool.Get().(map[string]string)
	// Clear the map
	for k := range m {
		delete(m, k)
	}
	return m
}

// PutMapStringString returns a map[string]string to the pool
func PutMapStringString(m map[string]string) {
	// Only pool maps that aren't too large to avoid memory leaks
	if len(m) < 256 {
		MapStringStringPool.Put(m)
	}
}
