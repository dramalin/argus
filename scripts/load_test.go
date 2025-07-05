package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// LoadTestConfig holds configuration for the load test
type LoadTestConfig struct {
	BaseURL         string
	ConcurrentUsers int
	RequestsPerUser int
	RequestDelay    time.Duration
}

// LoadTestResult holds the results of a load test
type LoadTestResult struct {
	TotalRequests   int
	SuccessRequests int
	FailedRequests  int
	TotalDuration   time.Duration
	AvgResponseTime time.Duration
	MinResponseTime time.Duration
	MaxResponseTime time.Duration
}

// LoadTester performs load testing on the Argus API
type LoadTester struct {
	config  LoadTestConfig
	client  *http.Client
	results []time.Duration
	mu      sync.Mutex
}

// NewLoadTester creates a new load tester instance
func NewLoadTester(config LoadTestConfig) *LoadTester {
	return &LoadTester{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		results: make([]time.Duration, 0),
	}
}

// RunTest executes the load test
func (lt *LoadTester) RunTest() *LoadTestResult {
	var wg sync.WaitGroup
	startTime := time.Now()

	// Endpoints to test
	endpoints := []string{
		"/api/cpu",
		"/api/memory",
		"/api/network",
		"/api/process",
		"/api/health",
	}

	totalRequests := lt.config.ConcurrentUsers * lt.config.RequestsPerUser * len(endpoints)
	successCount := 0
	failedCount := 0

	// Launch concurrent users
	for i := 0; i < lt.config.ConcurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()
			lt.simulateUser(userID, endpoints, &successCount, &failedCount)
		}(i)
	}

	wg.Wait()
	totalDuration := time.Since(startTime)

	// Calculate statistics
	result := &LoadTestResult{
		TotalRequests:   totalRequests,
		SuccessRequests: successCount,
		FailedRequests:  failedCount,
		TotalDuration:   totalDuration,
	}

	if len(lt.results) > 0 {
		result.AvgResponseTime = lt.calculateAverage()
		result.MinResponseTime = lt.calculateMin()
		result.MaxResponseTime = lt.calculateMax()
	}

	return result
}

// simulateUser simulates a single user making requests
func (lt *LoadTester) simulateUser(userID int, endpoints []string, successCount, failedCount *int) {
	for i := 0; i < lt.config.RequestsPerUser; i++ {
		for _, endpoint := range endpoints {
			url := lt.config.BaseURL + endpoint

			start := time.Now()
			resp, err := lt.client.Get(url)
			duration := time.Since(start)

			lt.mu.Lock()
			lt.results = append(lt.results, duration)
			lt.mu.Unlock()

			if err != nil {
				log.Printf("User %d: Request to %s failed: %v", userID, endpoint, err)
				*failedCount++
				continue
			}

			if resp.StatusCode == http.StatusOK {
				*successCount++
			} else {
				*failedCount++
				log.Printf("User %d: Request to %s returned status %d", userID, endpoint, resp.StatusCode)
			}

			// Read and discard response body
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()

			// Add delay between requests
			if lt.config.RequestDelay > 0 {
				time.Sleep(lt.config.RequestDelay)
			}
		}
	}
}

// calculateAverage calculates the average response time
func (lt *LoadTester) calculateAverage() time.Duration {
	var total time.Duration
	for _, duration := range lt.results {
		total += duration
	}
	return total / time.Duration(len(lt.results))
}

// calculateMin finds the minimum response time
func (lt *LoadTester) calculateMin() time.Duration {
	if len(lt.results) == 0 {
		return 0
	}
	min := lt.results[0]
	for _, duration := range lt.results {
		if duration < min {
			min = duration
		}
	}
	return min
}

// calculateMax finds the maximum response time
func (lt *LoadTester) calculateMax() time.Duration {
	if len(lt.results) == 0 {
		return 0
	}
	max := lt.results[0]
	for _, duration := range lt.results {
		if duration > max {
			max = duration
		}
	}
	return max
}

// PrintResults prints the load test results
func (result *LoadTestResult) PrintResults() {
	fmt.Println("=== Load Test Results ===")
	fmt.Printf("Total Requests: %d\n", result.TotalRequests)
	fmt.Printf("Successful Requests: %d\n", result.SuccessRequests)
	fmt.Printf("Failed Requests: %d\n", result.FailedRequests)
	fmt.Printf("Success Rate: %.2f%%\n", float64(result.SuccessRequests)/float64(result.TotalRequests)*100)
	fmt.Printf("Total Duration: %v\n", result.TotalDuration)
	fmt.Printf("Requests per Second: %.2f\n", float64(result.TotalRequests)/result.TotalDuration.Seconds())
	fmt.Printf("Average Response Time: %v\n", result.AvgResponseTime)
	fmt.Printf("Min Response Time: %v\n", result.MinResponseTime)
	fmt.Printf("Max Response Time: %v\n", result.MaxResponseTime)
}

func main() {
	// Default configuration
	config := LoadTestConfig{
		BaseURL:         "http://localhost:8080",
		ConcurrentUsers: 10,
		RequestsPerUser: 20,
		RequestDelay:    100 * time.Millisecond,
	}

	fmt.Println("Starting load test...")
	fmt.Printf("Target: %s\n", config.BaseURL)
	fmt.Printf("Concurrent Users: %d\n", config.ConcurrentUsers)
	fmt.Printf("Requests per User: %d\n", config.RequestsPerUser)
	fmt.Printf("Request Delay: %v\n", config.RequestDelay)
	fmt.Println()

	tester := NewLoadTester(config)
	result := tester.RunTest()

	fmt.Println()
	result.PrintResults()
}
