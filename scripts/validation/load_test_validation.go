// File: scripts/validation/load_test_validation.go
// Brief: Comprehensive load testing script for Argus performance validation
// Detailed: Validates performance under realistic conditions and measures improvements from optimizations
// Author: drama.lin@aver.com
// Date: 2024-07-04

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// LoadTestConfig defines the configuration for load testing
type LoadTestConfig struct {
	BaseURL         string        `json:"base_url"`
	ConcurrentUsers int           `json:"concurrent_users"`
	RequestsPerUser int           `json:"requests_per_user"`
	TestDuration    time.Duration `json:"test_duration"`
	RampUpDuration  time.Duration `json:"ramp_up_duration"`
	RequestTimeout  time.Duration `json:"request_timeout"`
}

// LoadTestResult represents the results of a load test
type LoadTestResult struct {
	TestName            string          `json:"test_name"`
	Config              LoadTestConfig  `json:"config"`
	StartTime           time.Time       `json:"start_time"`
	EndTime             time.Time       `json:"end_time"`
	TotalRequests       int             `json:"total_requests"`
	SuccessfulRequests  int             `json:"successful_requests"`
	FailedRequests      int             `json:"failed_requests"`
	AverageResponseTime time.Duration   `json:"average_response_time"`
	MinResponseTime     time.Duration   `json:"min_response_time"`
	MaxResponseTime     time.Duration   `json:"max_response_time"`
	RequestsPerSecond   float64         `json:"requests_per_second"`
	ErrorRate           float64         `json:"error_rate"`
	ResponseTimes       []time.Duration `json:"response_times"`
	StatusCodes         map[int]int     `json:"status_codes"`
}

// LoadTestSuite represents a collection of load tests
type LoadTestSuite struct {
	Results []LoadTestResult `json:"results"`
	Summary LoadTestSummary  `json:"summary"`
}

// LoadTestSummary provides overall performance metrics
type LoadTestSummary struct {
	TotalTests          int           `json:"total_tests"`
	TotalRequests       int           `json:"total_requests"`
	OverallSuccessRate  float64       `json:"overall_success_rate"`
	AverageRPS          float64       `json:"average_rps"`
	AverageResponseTime time.Duration `json:"average_response_time"`
}

func main() {
	fmt.Println("🚀 Starting Argus Load Test Validation")
	fmt.Println("======================================")

	// Default configuration
	config := LoadTestConfig{
		BaseURL:         "http://localhost:8080",
		ConcurrentUsers: 50,
		RequestsPerUser: 100,
		TestDuration:    5 * time.Minute,
		RampUpDuration:  30 * time.Second,
		RequestTimeout:  10 * time.Second,
	}

	// Check if server is running
	if !isServerRunning(config.BaseURL) {
		fmt.Printf("⚠️  Server is not running at %s\n", config.BaseURL)
		fmt.Println("Please start the Argus server before running load tests.")
		fmt.Println("Example: go run cmd/argus/main.go")
		return
	}

	suite := &LoadTestSuite{}

	// Test 1: Metrics Collection Performance
	fmt.Println("\n📊 Testing Metrics Collection Performance...")
	metricsResult := runLoadTest("Metrics Collection", config, []string{
		"/api/metrics/system",
		"/api/metrics/process",
		"/api/metrics/alerts",
	})
	suite.Results = append(suite.Results, metricsResult)

	// Test 2: Process Metrics with Pagination
	fmt.Println("\n🔍 Testing Process Metrics with Pagination...")
	processConfig := config
	processConfig.ConcurrentUsers = 30
	processResult := runLoadTest("Process Metrics Pagination", processConfig, []string{
		"/api/process?limit=50&offset=0",
		"/api/process?limit=20&offset=20&sort_by=cpu&sort_order=desc",
		"/api/process?limit=10&min_cpu=5.0&top_n=10",
	})
	suite.Results = append(suite.Results, processResult)

	// Test 3: Alert System Performance
	fmt.Println("\n🚨 Testing Alert System Performance...")
	alertConfig := config
	alertConfig.ConcurrentUsers = 25
	alertResult := runLoadTest("Alert System", alertConfig, []string{
		"/api/alerts",
		"/api/alerts/active",
		"/api/alerts/history",
	})
	suite.Results = append(suite.Results, alertResult)

	// Test 4: HTTP Server and Middleware Performance
	fmt.Println("\n🌐 Testing HTTP Server and Middleware Performance...")
	serverConfig := config
	serverConfig.ConcurrentUsers = 100
	serverResult := runLoadTest("HTTP Server", serverConfig, []string{
		"/api/health",
		"/api/status",
		"/api/metrics/system",
	})
	suite.Results = append(suite.Results, serverResult)

	// Generate summary
	suite.Summary = generateLoadTestSummary(suite.Results)

	// Save results
	outputDir := "performance_results"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Save JSON report
	reportPath := fmt.Sprintf("%s/load_test_report.json", outputDir)
	if err := saveLoadTestReport(suite, reportPath); err != nil {
		log.Fatalf("Failed to save load test report: %v", err)
	}

	// Generate markdown summary
	summaryPath := fmt.Sprintf("%s/load_test_summary.md", outputDir)
	if err := generateLoadTestMarkdown(suite, summaryPath); err != nil {
		log.Fatalf("Failed to generate load test summary: %v", err)
	}

	fmt.Printf("\n✅ Load test validation completed successfully!\n")
	fmt.Printf("📁 Results saved to: %s\n", outputDir)
	fmt.Printf("📊 Report: %s\n", reportPath)
	fmt.Printf("📝 Summary: %s\n", summaryPath)

	// Print quick summary
	printLoadTestSummary(suite)
}

func isServerRunning(baseURL string) bool {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(baseURL + "/api/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func runLoadTest(testName string, config LoadTestConfig, endpoints []string) LoadTestResult {
	result := LoadTestResult{
		TestName:      testName,
		Config:        config,
		StartTime:     time.Now(),
		StatusCodes:   make(map[int]int),
		ResponseTimes: make([]time.Duration, 0),
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	client := &http.Client{Timeout: config.RequestTimeout}

	ctx, cancel := context.WithTimeout(context.Background(), config.TestDuration)
	defer cancel()

	// Run concurrent requests
	for i := 0; i < config.ConcurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()

			// Ramp up delay
			rampUpDelay := time.Duration(userID) * config.RampUpDuration / time.Duration(config.ConcurrentUsers)
			time.Sleep(rampUpDelay)

			for j := 0; j < config.RequestsPerUser; j++ {
				select {
				case <-ctx.Done():
					return
				default:
					endpoint := endpoints[j%len(endpoints)]
					url := config.BaseURL + endpoint

					start := time.Now()
					resp, err := client.Get(url)
					responseTime := time.Since(start)

					mu.Lock()
					result.TotalRequests++
					result.ResponseTimes = append(result.ResponseTimes, responseTime)

					if err != nil {
						result.FailedRequests++
					} else {
						result.SuccessfulRequests++
						result.StatusCodes[resp.StatusCode]++
						resp.Body.Close()
					}
					mu.Unlock()
				}
			}
		}(i)
	}

	wg.Wait()
	result.EndTime = time.Now()

	// Calculate metrics
	if len(result.ResponseTimes) > 0 {
		var totalTime time.Duration
		result.MinResponseTime = result.ResponseTimes[0]
		result.MaxResponseTime = result.ResponseTimes[0]

		for _, rt := range result.ResponseTimes {
			totalTime += rt
			if rt < result.MinResponseTime {
				result.MinResponseTime = rt
			}
			if rt > result.MaxResponseTime {
				result.MaxResponseTime = rt
			}
		}

		result.AverageResponseTime = totalTime / time.Duration(len(result.ResponseTimes))
		testDuration := result.EndTime.Sub(result.StartTime)
		result.RequestsPerSecond = float64(result.TotalRequests) / testDuration.Seconds()
		result.ErrorRate = float64(result.FailedRequests) / float64(result.TotalRequests) * 100
	}

	return result
}

func generateLoadTestSummary(results []LoadTestResult) LoadTestSummary {
	summary := LoadTestSummary{
		TotalTests: len(results),
	}

	var totalRequests int
	var totalSuccessful int
	var totalResponseTime time.Duration
	var totalRPS float64

	for _, result := range results {
		totalRequests += result.TotalRequests
		totalSuccessful += result.SuccessfulRequests
		totalResponseTime += result.AverageResponseTime
		totalRPS += result.RequestsPerSecond
	}

	summary.TotalRequests = totalRequests
	if totalRequests > 0 {
		summary.OverallSuccessRate = float64(totalSuccessful) / float64(totalRequests) * 100
	}
	if len(results) > 0 {
		summary.AverageRPS = totalRPS / float64(len(results))
		summary.AverageResponseTime = totalResponseTime / time.Duration(len(results))
	}

	return summary
}

func saveLoadTestReport(suite *LoadTestSuite, path string) error {
	data, err := json.MarshalIndent(suite, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func generateLoadTestMarkdown(suite *LoadTestSuite, path string) error {
	var sb bytes.Buffer

	sb.WriteString("# Argus Load Test Validation Report\n\n")
	sb.WriteString(fmt.Sprintf("**Generated:** %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	sb.WriteString("## Summary\n\n")
	sb.WriteString(fmt.Sprintf("- **Total Tests:** %d\n", suite.Summary.TotalTests))
	sb.WriteString(fmt.Sprintf("- **Total Requests:** %d\n", suite.Summary.TotalRequests))
	sb.WriteString(fmt.Sprintf("- **Overall Success Rate:** %.2f%%\n", suite.Summary.OverallSuccessRate))
	sb.WriteString(fmt.Sprintf("- **Average RPS:** %.2f\n", suite.Summary.AverageRPS))
	sb.WriteString(fmt.Sprintf("- **Average Response Time:** %v\n", suite.Summary.AverageResponseTime))

	sb.WriteString("\n## Test Results\n\n")
	sb.WriteString("| Test Name | Requests | Success Rate | Avg Response Time | RPS | Error Rate |\n")
	sb.WriteString("|-----------|----------|--------------|-------------------|-----|------------|\n")

	for _, result := range suite.Results {
		successRate := float64(result.SuccessfulRequests) / float64(result.TotalRequests) * 100
		sb.WriteString(fmt.Sprintf("| %s | %d | %.2f%% | %v | %.2f | %.2f%% |\n",
			result.TestName, result.TotalRequests, successRate, result.AverageResponseTime, result.RequestsPerSecond, result.ErrorRate))
	}

	sb.WriteString("\n## Detailed Results\n\n")
	for _, result := range suite.Results {
		sb.WriteString(fmt.Sprintf("### %s\n\n", result.TestName))
		sb.WriteString(fmt.Sprintf("- **Concurrent Users:** %d\n", result.Config.ConcurrentUsers))
		sb.WriteString(fmt.Sprintf("- **Requests per User:** %d\n", result.Config.RequestsPerUser))
		sb.WriteString(fmt.Sprintf("- **Total Requests:** %d\n", result.TotalRequests))
		sb.WriteString(fmt.Sprintf("- **Successful Requests:** %d\n", result.SuccessfulRequests))
		sb.WriteString(fmt.Sprintf("- **Failed Requests:** %d\n", result.FailedRequests))
		sb.WriteString(fmt.Sprintf("- **Average Response Time:** %v\n", result.AverageResponseTime))
		sb.WriteString(fmt.Sprintf("- **Min Response Time:** %v\n", result.MinResponseTime))
		sb.WriteString(fmt.Sprintf("- **Max Response Time:** %v\n", result.MaxResponseTime))
		sb.WriteString(fmt.Sprintf("- **Requests per Second:** %.2f\n", result.RequestsPerSecond))
		sb.WriteString(fmt.Sprintf("- **Error Rate:** %.2f%%\n", result.ErrorRate))

		sb.WriteString("\n**Status Codes:**\n")
		for code, count := range result.StatusCodes {
			sb.WriteString(fmt.Sprintf("- %d: %d requests\n", code, count))
		}
		sb.WriteString("\n")
	}

	return os.WriteFile(path, sb.Bytes(), 0644)
}

func printLoadTestSummary(suite *LoadTestSuite) {
	fmt.Println("\n📈 Load Test Summary:")
	fmt.Printf("   • Total tests: %d\n", suite.Summary.TotalTests)
	fmt.Printf("   • Total requests: %d\n", suite.Summary.TotalRequests)
	fmt.Printf("   • Overall success rate: %.2f%%\n", suite.Summary.OverallSuccessRate)
	fmt.Printf("   • Average RPS: %.2f\n", suite.Summary.AverageRPS)
	fmt.Printf("   • Average response time: %v\n", suite.Summary.AverageResponseTime)
}
