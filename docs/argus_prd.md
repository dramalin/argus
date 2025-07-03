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
| Task Scheduler   | Go + cron          | Schedule and run system maintenance tasks |
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
```

## 4. Task Management System Design

### 4.1 Overview

The Task Management System allows scheduling and executing recurring system maintenance tasks using cron expressions. It follows the repository pattern for data persistence and provides a flexible API for task management.

### 4.2 Task Types

| Task Type | Purpose | Parameters |
|-----------|---------|------------|
| `log_rotation` | Rotate and archive log files based on size | `log_dir`, `pattern`, `max_size_mb`, `keep_count` |
| `metrics_aggregation` | Collect and aggregate system metrics | `metrics_dir`, `retention_days` |
| `health_check` | Monitor system and service health | `url`, `timeout` |
| `system_cleanup` | Purge old temporary files | `cleanup_dir`, `pattern`, `retention_days` |

### 4.3 System Components

1. **Task Models** - Define task configurations and execution records
2. **Task Repository** - Persist tasks and executions using file-based storage
3. **Task Runners** - Execute specific task types
4. **Task Scheduler** - Schedule and trigger tasks using cron expressions
5. **API Handlers** - Expose RESTful endpoints for task management

### 4.4 API Endpoints

| Endpoint | Method | Description | Request/Response |
|----------|--------|-------------|------------------|
| `/api/tasks` | GET | List all tasks | Response: Array of task configurations |
| `/api/tasks` | POST | Create a new task | Request: Task configuration, Response: Created task |
| `/api/tasks/:id` | GET | Get a specific task | Response: Task configuration |
| `/api/tasks/:id` | PUT | Update a task | Request: Updated task, Response: Updated task |
| `/api/tasks/:id` | DELETE | Delete a task | Response: Success message |
| `/api/tasks/:id/executions` | GET | Get execution history | Response: Array of executions |
| `/api/tasks/:id/run` | POST | Run task immediately | Response: Execution record |

### 4.5 Task Configuration Schema

```json
{
  "id": "string",
  "name": "string",
  "description": "string",
  "type": "string",
  "enabled": true,
  "schedule": {
    "cron_expression": "string",
    "one_time": false,
    "next_run_time": "2025-07-03T12:00:00Z"
  },
  "parameters": {
    "param1": "value1",
    "param2": "value2"
  },
  "created_at": "2025-07-03T10:00:00Z",
  "updated_at": "2025-07-03T10:00:00Z"
}
```

### 4.6 Task Execution Schema

```json
{
  "id": "string",
  "task_id": "string",
  "status": "pending|running|completed|failed",
  "start_time": "2025-07-03T12:00:00Z",
  "end_time": "2025-07-03T12:01:00Z",
  "error": "string",
  "output": "string"
}
```

### 4.7 Example: Creating a Log Rotation Task

```http
POST /api/tasks HTTP/1.1
Content-Type: application/json

{
  "name": "Daily Log Rotation",
  "description": "Rotate log files daily at midnight",
  "type": "log_rotation",
  "enabled": true,
  "schedule": {
    "cron_expression": "0 0 * * *"
  },
  "parameters": {
    "log_dir": "/var/log/argus",
    "pattern": "*.log",
    "max_size_mb": "10",
    "keep_count": "7"
  }
}
```

Response:

```json
{
  "id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
  "name": "Daily Log Rotation",
  "description": "Rotate log files daily at midnight",
  "type": "log_rotation",
  "enabled": true,
  "schedule": {
    "cron_expression": "0 0 * * *",
    "one_time": false,
    "next_run_time": "2025-07-04T00:00:00Z"
  },
  "parameters": {
    "log_dir": "/var/log/argus",
    "pattern": "*.log",
    "max_size_mb": "10",
    "keep_count": "7"
  },
  "created_at": "2025-07-03T15:30:00Z",
  "updated_at": "2025-07-03T15:30:00Z"
}
```
