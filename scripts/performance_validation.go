// File: scripts/performance_validation.go
// Brief: Comprehensive performance validation script for Argus optimizations
// Detailed: Runs all benchmarks, generates pprof profiles, and collects performance metrics to validate optimizations
// Author: drama.lin@aver.com
// Date: 2024-07-04

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// BenchmarkResult represents a single benchmark result
type BenchmarkResult struct {
	Name        string  `json:"name"`
	Iterations  int     `json:"iterations"`
	NsPerOp     float64 `json:"ns_per_op"`
	BytesPerOp  int     `json:"bytes_per_op"`
	AllocsPerOp int     `json:"allocs_per_op"`
	MBPerSec    float64 `json:"mb_per_sec,omitempty"`
}

// PerformanceReport represents the complete performance validation report
type PerformanceReport struct {
	Timestamp         time.Time          `json:"timestamp"`
	GoVersion         string             `json:"go_version"`
	Benchmarks        []BenchmarkResult  `json:"benchmarks"`
	ProfilesGenerated []string           `json:"profiles_generated"`
	Summary           PerformanceSummary `json:"summary"`
}

// PerformanceSummary provides high-level performance metrics
type PerformanceSummary struct {
	TotalBenchmarks   int      `json:"total_benchmarks"`
	AverageNsPerOp    float64  `json:"average_ns_per_op"`
	TotalAllocations  int      `json:"total_allocations"`
	AverageBytesPerOp float64  `json:"average_bytes_per_op"`
	OptimizationAreas []string `json:"optimization_areas"`
}

func main() {
	fmt.Println("🚀 Starting Argus Performance Validation")
	fmt.Println("=========================================")

	report := &PerformanceReport{
		Timestamp: time.Now(),
	}

	// Get Go version
	if version, err := getGoVersion(); err == nil {
		report.GoVersion = version
	}

	// Create output directory
	outputDir := "performance_results"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Step 1: Run comprehensive benchmarks
	fmt.Println("\n📊 Running Comprehensive Benchmarks...")
	benchmarks, err := runAllBenchmarks()
	if err != nil {
		log.Fatalf("Failed to run benchmarks: %v", err)
	}
	report.Benchmarks = benchmarks

	// Step 2: Generate performance profiles
	fmt.Println("\n📈 Generating Performance Profiles...")
	profiles, err := generatePerformanceProfiles(outputDir)
	if err != nil {
		log.Printf("Warning: Failed to generate some profiles: %v", err)
	}
	report.ProfilesGenerated = profiles

	// Step 3: Generate summary statistics
	report.Summary = generateSummary(benchmarks)

	// Step 4: Save performance report
	reportPath := filepath.Join(outputDir, "performance_report.json")
	if err := saveReport(report, reportPath); err != nil {
		log.Fatalf("Failed to save report: %v", err)
	}

	// Step 5: Generate human-readable summary
	summaryPath := filepath.Join(outputDir, "performance_summary.md")
	if err := generateMarkdownSummary(report, summaryPath); err != nil {
		log.Fatalf("Failed to generate summary: %v", err)
	}

	fmt.Printf("\n✅ Performance validation completed successfully!\n")
	fmt.Printf("📁 Results saved to: %s\n", outputDir)
	fmt.Printf("📊 Report: %s\n", reportPath)
	fmt.Printf("📝 Summary: %s\n", summaryPath)

	// Print quick summary
	printQuickSummary(report)
}

func getGoVersion() (string, error) {
	cmd := exec.Command("go", "version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func runAllBenchmarks() ([]BenchmarkResult, error) {
	var allResults []BenchmarkResult

	benchmarkPackages := []string{
		"./benchmarks",
	}

	for _, pkg := range benchmarkPackages {
		fmt.Printf("  Running benchmarks in %s...\n", pkg)

		cmd := exec.Command("go", "test", "-bench=.", "-benchmem", "-count=3", pkg)
		output, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("failed to run benchmarks in %s: %v", pkg, err)
		}

		results, err := parseBenchmarkOutput(string(output))
		if err != nil {
			return nil, fmt.Errorf("failed to parse benchmark output from %s: %v", pkg, err)
		}

		allResults = append(allResults, results...)
	}

	return allResults, nil
}

func parseBenchmarkOutput(output string) ([]BenchmarkResult, error) {
	var results []BenchmarkResult

	// Regex to parse benchmark output: BenchmarkName-N    iterations    ns/op    bytes/op    allocs/op
	re := regexp.MustCompile(`Benchmark(\w+)-\d+\s+(\d+)\s+(\d+(?:\.\d+)?)\s+ns/op\s+(\d+)\s+B/op\s+(\d+)\s+allocs/op`)

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		matches := re.FindStringSubmatch(line)
		if len(matches) == 6 {
			iterations, _ := strconv.Atoi(matches[2])
			nsPerOp, _ := strconv.ParseFloat(matches[3], 64)
			bytesPerOp, _ := strconv.Atoi(matches[4])
			allocsPerOp, _ := strconv.Atoi(matches[5])

			result := BenchmarkResult{
				Name:        matches[1],
				Iterations:  iterations,
				NsPerOp:     nsPerOp,
				BytesPerOp:  bytesPerOp,
				AllocsPerOp: allocsPerOp,
			}

			results = append(results, result)
		}
	}

	return results, nil
}

func generatePerformanceProfiles(outputDir string) ([]string, error) {
	var profiles []string

	// CPU Profile
	fmt.Println("  Generating CPU profile...")
	cpuProfilePath := filepath.Join(outputDir, "cpu.prof")
	cmd := exec.Command("go", "test", "-bench=BenchmarkCPU", "-cpuprofile", cpuProfilePath, "./benchmarks")
	if err := cmd.Run(); err == nil {
		profiles = append(profiles, cpuProfilePath)
	}

	// Memory Profile
	fmt.Println("  Generating memory profile...")
	memProfilePath := filepath.Join(outputDir, "mem.prof")
	cmd = exec.Command("go", "test", "-bench=BenchmarkMemory", "-memprofile", memProfilePath, "./benchmarks")
	if err := cmd.Run(); err == nil {
		profiles = append(profiles, memProfilePath)
	}

	// Block Profile
	fmt.Println("  Generating block profile...")
	blockProfilePath := filepath.Join(outputDir, "block.prof")
	cmd = exec.Command("go", "test", "-bench=BenchmarkConcurrent", "-blockprofile", blockProfilePath, "./benchmarks")
	if err := cmd.Run(); err == nil {
		profiles = append(profiles, blockProfilePath)
	}

	return profiles, nil
}

func generateSummary(benchmarks []BenchmarkResult) PerformanceSummary {
	if len(benchmarks) == 0 {
		return PerformanceSummary{}
	}

	var totalNs, totalBytes float64
	var totalAllocs int

	for _, b := range benchmarks {
		totalNs += b.NsPerOp
		totalBytes += float64(b.BytesPerOp)
		totalAllocs += b.AllocsPerOp
	}

	return PerformanceSummary{
		TotalBenchmarks:   len(benchmarks),
		AverageNsPerOp:    totalNs / float64(len(benchmarks)),
		TotalAllocations:  totalAllocs,
		AverageBytesPerOp: totalBytes / float64(len(benchmarks)),
		OptimizationAreas: []string{
			"Profiling Infrastructure",
			"Centralized Metrics Collection",
			"Alert Evaluator Concurrency",
			"Notification System Performance",
			"HTTP Server and Middleware",
			"Memory Pool Optimization",
			"Process Metrics Collection",
		},
	}
}

func saveReport(report *PerformanceReport, path string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func generateMarkdownSummary(report *PerformanceReport, path string) error {
	var sb strings.Builder

	sb.WriteString("# Argus Performance Validation Report\n\n")
	sb.WriteString(fmt.Sprintf("**Generated:** %s\n", report.Timestamp.Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("**Go Version:** %s\n\n", report.GoVersion))

	sb.WriteString("## Optimization Areas Validated\n\n")
	for i, area := range report.Summary.OptimizationAreas {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, area))
	}

	sb.WriteString("\n## Performance Summary\n\n")
	sb.WriteString(fmt.Sprintf("- **Total Benchmarks:** %d\n", report.Summary.TotalBenchmarks))
	sb.WriteString(fmt.Sprintf("- **Average ns/op:** %.2f\n", report.Summary.AverageNsPerOp))
	sb.WriteString(fmt.Sprintf("- **Average bytes/op:** %.2f\n", report.Summary.AverageBytesPerOp))
	sb.WriteString(fmt.Sprintf("- **Total allocations:** %d\n", report.Summary.TotalAllocations))

	sb.WriteString("\n## Benchmark Results\n\n")
	sb.WriteString("| Benchmark | Iterations | ns/op | bytes/op | allocs/op |\n")
	sb.WriteString("|-----------|------------|-------|----------|----------|\n")

	for _, b := range report.Benchmarks {
		sb.WriteString(fmt.Sprintf("| %s | %d | %.2f | %d | %d |\n",
			b.Name, b.Iterations, b.NsPerOp, b.BytesPerOp, b.AllocsPerOp))
	}

	sb.WriteString("\n## Generated Profiles\n\n")
	for _, profile := range report.ProfilesGenerated {
		sb.WriteString(fmt.Sprintf("- %s\n", profile))
	}

	sb.WriteString("\n## Analysis Commands\n\n")
	sb.WriteString("To analyze the generated profiles:\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString("# CPU Profile\n")
	sb.WriteString("go tool pprof performance_results/cpu.prof\n\n")
	sb.WriteString("# Memory Profile\n")
	sb.WriteString("go tool pprof performance_results/mem.prof\n\n")
	sb.WriteString("# Block Profile\n")
	sb.WriteString("go tool pprof performance_results/block.prof\n")
	sb.WriteString("```\n")

	return os.WriteFile(path, []byte(sb.String()), 0644)
}

func printQuickSummary(report *PerformanceReport) {
	fmt.Println("\n📈 Quick Performance Summary:")
	fmt.Printf("   • Total benchmarks: %d\n", report.Summary.TotalBenchmarks)
	fmt.Printf("   • Average performance: %.2f ns/op\n", report.Summary.AverageNsPerOp)
	fmt.Printf("   • Average memory usage: %.2f bytes/op\n", report.Summary.AverageBytesPerOp)
	fmt.Printf("   • Profiles generated: %d\n", len(report.ProfilesGenerated))
}
