package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
)

// CORS middleware
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func getCPU(c *gin.Context) {
	loadAvg, err := load.Avg()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get load average: " + err.Error()})
		return
	}
	
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get CPU usage: " + err.Error()})
		return
	}
	
	usage := 0.0
	if len(cpuPercent) > 0 {
		usage = cpuPercent[0]
	}

	c.JSON(http.StatusOK, gin.H{
		"load1":         loadAvg.Load1,
		"load5":         loadAvg.Load5,
		"load15":        loadAvg.Load15,
		"usage_percent": usage,
	})
}

func getMemory(c *gin.Context) {
	vm, err := mem.VirtualMemory()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get memory info: " + err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"total":        vm.Total,
		"used":         vm.Used,
		"free":         vm.Free,
		"used_percent": vm.UsedPercent,
	})
}

func getNetwork(c *gin.Context) {
	ioCounters, err := net.IOCounters(false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get network stats: " + err.Error()})
		return
	}
	
	if len(ioCounters) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No network interfaces found"})
		return
	}
	
	io := ioCounters[0]
	c.JSON(http.StatusOK, gin.H{
		"bytes_sent":   io.BytesSent,
		"bytes_recv":   io.BytesRecv,
		"packets_sent": io.PacketsSent,
		"packets_recv": io.PacketsRecv,
	})
}

func getProcess(c *gin.Context) {
	procs, err := process.Processes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get process list: " + err.Error()})
		return
	}
	
	result := []gin.H{}
	count := 0
	
	// Limit to top 20 processes to avoid overwhelming the frontend
	for _, p := range procs {
		if count >= 20 {
			break
		}
		
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
		
		result = append(result, gin.H{
			"pid":         p.Pid,
			"name":        name,
			"cpu_percent": cpuP,
			"mem_percent": memP,
		})
		count++
	}
	
	c.JSON(http.StatusOK, result)
}

func main() {
	// Set Gin to release mode for production
	gin.SetMode(gin.ReleaseMode)
	
	router := gin.Default()
	
	// Add CORS middleware
	router.Use(corsMiddleware())
	
	// Serve static files from webapp directory
	router.Static("/static", "./webapp")
	router.StaticFile("/", "./webapp/index.html")
	
	// API routes
	api := router.Group("/api")
	{
		api.GET("/cpu", getCPU)
		api.GET("/memory", getMemory)
		api.GET("/network", getNetwork)
		api.GET("/process", getProcess)
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	router.Run(":8080")
}
