package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
)

func getCPU(c *gin.Context) {
	loadAvg, err := load.Avg()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	cpuPercent, err := cpu.Percent(0, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
	if err != nil || len(ioCounters) == 0 {
		if err == nil {
			err = gin.Error{Err: err, Type: gin.ErrorTypePublic}
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	result := []gin.H{}
	for _, p := range procs {
		name, err := p.Name()
		if err != nil {
			continue
		}
		cpuP, err := p.CPUPercent()
		if err != nil {
			continue
		}
		memP, err := p.MemoryPercent()
		if err != nil {
			continue
		}
		result = append(result, gin.H{
			"pid":         p.Pid,
			"name":        name,
			"cpu_percent": cpuP,
			"mem_percent": memP,
		})
	}
	c.JSON(http.StatusOK, result)
}

func main() {
	router := gin.Default()
	router.GET("/api/cpu", getCPU)
	router.GET("/api/memory", getMemory)
	router.GET("/api/network", getNetwork)
	router.GET("/api/process", getProcess)

	router.Run(":8080")
}
