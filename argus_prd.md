# Linux System Performance Monitoring Web Application Design Document

## 1. Project Overview

- **Objective**: Implement a backend API based on GoLang + Gin and a React.js frontend UI to monitor Linux system CPU load, memory usage, network traffic, and process status.
- **Architecture**:
  - Backend: Go + Gin, using the `gopsutil` package to read system data and provide RESTful API.
  - Frontend: React.js, periodically calling APIs to display monitoring data with chart visualization.

---

## 2. System Architecture Design

| System Layer     | Technology         | Description                      |
|------------------|--------------------|----------------------------------|
| Backend API      | Go + Gin           | Collect Linux system resources and provide API |
| System Monitor Library | gopsutil      | Retrieve CPU, Memory, Network, Process usage data |
| Frontend UI      | React.js           | User interface, display monitoring information |
| Data Communication | REST API (JSON)   | Frontend-backend data interaction |

---

## 3. Backend Design

### 3.1 Required Packages

- [Gin](https://github.com/gin-gonic/gin) - Lightweight HTTP Web framework
- [gopsutil](https://github.com/shirou/gopsutil) - System resource reading package

### 3.2 API Specifications

| Route          | Method | Function              | Return Data Format (JSON)                                   |
|----------------|--------|-----------------------|-------------------------------------------------------------|
| `/api/cpu`     | GET    | Get CPU load and usage rate | `{ "load1": float, "load5": float, "load15": float, "usage_percent": float }` |
| `/api/memory`  | GET    | Get memory usage status | `{ "total": uint64, "used": uint64, "free": uint64, "used_percent": float }` |
| `/api/network` | GET    | Get network traffic statistics | `{ "bytes_sent": uint64, "bytes_recv": uint64, "packets_sent": uint64, "packets_recv": uint64 }` |
| `/api/process` | GET    | Get process resource usage status | `[ { "pid": int, "name": string, "cpu_percent": float, "mem_percent": float }, ... ]` |

### 3.3 Sample Backend Code (CPU API)

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/shirou/gopsutil/cpu"
    "github.com/shirou/gopsutil/load"
    "net/http"
)

func GetCpuLoad(c *gin.Context) {
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
    c.JSON(http.StatusOK, gin.H{
        "load1":         loadAvg.Load1,
        "load5":         loadAvg.Load5,
        "load15":        loadAvg.Load15,
        "usage_percent": cpuPercent[0],
    })
}

func main() {
    router := gin.Default()
    router.GET("/api/cpu", GetCpuLoad)
    // 其他 API 也可類似實作
    router.Run(":8080")
}
