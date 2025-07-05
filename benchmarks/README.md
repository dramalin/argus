# Argus Performance Benchmarks

This directory contains benchmark tests for the Argus system monitoring application, designed to measure performance and identify optimization opportunities.

## Overview

The benchmarks are organized into several categories:

- **Metrics Collection**: Tests the performance of CPU, memory, network, and process metrics collection
- **Alert Evaluation**: Tests the alert evaluation system performance
- **Notification System**: Tests the notification processing, rate limiting, and email queuing performance
- **HTTP Server & Middleware**: Tests the HTTP server, middleware stack, and static file serving performance
- **Memory Pool Optimization**: Tests the performance improvements from object pooling for bytes.Buffer, strings.Builder, slices, and maps
- **Process Optimization**: Tests the performance of process metrics collection with pagination, filtering, and efficient sorting algorithms
- **HTTP Server**: Tests the HTTP server and middleware performance
- **Concurrent Operations**: Tests system behavior under concurrent load

## Running Benchmarks

### Prerequisites

1. Ensure the Argus application dependencies are installed:

   ```bash
   go mod download
   ```

2. Make sure you have the required system monitoring permissions for gopsutil to work properly.

### Basic Benchmark Execution

Run all benchmarks:

```bash
go test -bench=. ./benchmarks/
```

Run specific benchmark categories:

```bash
# Metrics collection benchmarks
go test -bench=BenchmarkCPU ./benchmarks/
go test -bench=BenchmarkMemory ./benchmarks/
go test -bench=BenchmarkNetwork ./benchmarks/
go test -bench=BenchmarkProcess ./benchmarks/

# Alert system benchmarks
go test -bench=BenchmarkAlert ./benchmarks/

# Notification system benchmarks
go test -bench=BenchmarkNotification ./benchmarks/
go test -bench=BenchmarkRateLimit ./benchmarks/
go test -bench=BenchmarkEmail ./benchmarks/

# HTTP server and middleware benchmarks
go test -bench=BenchmarkMiddleware ./benchmarks/
go test -bench=BenchmarkStatic ./benchmarks/
go test -bench=BenchmarkAPI ./benchmarks/
go test -bench=BenchmarkConcurrent ./benchmarks/

# Memory pool optimization benchmarks
go test -bench=BenchmarkBytesBuffer ./benchmarks/
go test -bench=BenchmarkStringsBuilder ./benchmarks/
go test -bench=BenchmarkStringSlice ./benchmarks/
go test -bench=BenchmarkMapStringString ./benchmarks/
go test -bench=BenchmarkTemplate ./benchmarks/
go test -bench=BenchmarkConcurrentPool ./benchmarks/
go test -bench=BenchmarkMemoryPool ./benchmarks/
go test -bench=BenchmarkPoolSize ./benchmarks/

# Process optimization benchmarks
go test -bench=BenchmarkProcessFiltering ./benchmarks/
go test -bench=BenchmarkProcessPagination ./benchmarks/
go test -bench=BenchmarkTopNSelection ./benchmarks/
go test -bench=BenchmarkProcessSorting ./benchmarks/
go test -bench=BenchmarkConcurrentProcessAccess ./benchmarks/
```

### Advanced Benchmark Options

Run benchmarks with memory profiling:

```bash
go test -bench=. -memprofile=mem.prof ./benchmarks/
```

Run benchmarks with CPU profiling:

```bash
go test -bench=. -cpuprofile=cpu.prof ./benchmarks/
```

Run benchmarks multiple times for statistical significance:

```bash
go test -bench=. -count=5 ./benchmarks/
```

Run benchmarks with specific duration:

```bash
go test -bench=. -benchtime=30s ./benchmarks/
```

## Load Testing

The `scripts/load_test.go` file provides a load testing utility to test the HTTP API under realistic conditions.

### Running Load Tests

1. Start the Argus server:

   ```bash
   go run cmd/argus/main.go
   ```

2. In another terminal, run the load test:

   ```bash
   go run scripts/load_test.go
   ```

### Load Test Configuration

You can modify the load test parameters in `scripts/load_test.go`:

- `ConcurrentUsers`: Number of concurrent users (default: 10)
- `RequestsPerUser`: Number of requests per user (default: 20)
- `RequestDelay`: Delay between requests (default: 100ms)
- `BaseURL`: Target server URL (default: <http://localhost:8080>)

## Profiling Integration

With debug mode enabled in the configuration, you can access profiling endpoints while the server is running:

### CPU Profiling

```bash
go tool pprof http://localhost:8080/debug/pprof/profile?seconds=30
```

### Memory Profiling

```bash
go tool pprof http://localhost:8080/debug/pprof/heap
```

### Goroutine Analysis

```bash
go tool pprof http://localhost:8080/debug/pprof/goroutine
```

### Trace Analysis

```bash
curl http://localhost:8080/debug/pprof/trace?seconds=5 > trace.out
go tool trace trace.out
```

## Interpreting Results

### Benchmark Output Format

```
BenchmarkCPUCollection-8    1000    1234567 ns/op    456 B/op    7 allocs/op
```

- `BenchmarkCPUCollection-8`: Benchmark name with GOMAXPROCS
- `1000`: Number of iterations
- `1234567 ns/op`: Nanoseconds per operation
- `456 B/op`: Bytes allocated per operation
- `7 allocs/op`: Number of allocations per operation

### Performance Targets

Based on the optimize-go.mdc guidelines, we target:

- **Metrics Collection**: < 10ms per collection cycle
- **Alert Evaluation**: < 5ms per alert evaluation
- **HTTP Response**: < 100ms for API endpoints
- **Memory Allocations**: Minimize allocations per operation
- **Goroutine Usage**: Stable goroutine count under load

### Optimization Indicators

Look for these patterns in benchmark results:

1. **High allocation counts**: Indicates potential for memory pooling
2. **Long response times**: Suggests need for caching or algorithm optimization
3. **High variance**: May indicate lock contention or resource competition
4. **Memory growth**: Could signal memory leaks or inefficient data structures

## Continuous Monitoring

For ongoing performance monitoring:

1. Run benchmarks before and after code changes
2. Set up automated benchmark runs in CI/CD pipeline
3. Track performance metrics over time
4. Use profiling data to guide optimization efforts

## Troubleshooting

### Common Issues

1. **Permission errors**: Ensure proper system monitoring permissions
2. **Resource exhaustion**: Reduce concurrent users or requests for resource-constrained systems
3. **Network timeouts**: Increase timeout values in load test configuration
4. **Inconsistent results**: Run multiple iterations and use statistical analysis

### Debug Mode Requirements

Ensure debug mode is enabled in your configuration:

```yaml
debug:
  enabled: true
  pprof_enabled: true
  pprof_path: "/debug/pprof"
  benchmark_enabled: true
```
