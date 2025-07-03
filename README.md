# Argus System Monitor

A real-time Linux system performance monitoring web application built with Go (Gin) backend and React.js frontend.

## Features

- **Real-time Monitoring**: CPU usage, memory statistics, network traffic, and process information
- **Interactive Charts**: Visual representation of system metrics using Chart.js
- **Responsive Design**: Modern, mobile-friendly interface
- **RESTful API**: Clean API endpoints for system data
- **Process Management**: Sortable process table with CPU and memory usage
- **Task Management**: Schedule and run system maintenance tasks with cron expressions
  - Log Rotation: Automatically rotate and archive log files
  - Metrics Aggregation: Collect and aggregate system metrics
  - Health Checks: Monitor system and service health
  - System Cleanup: Purge old temporary files

## Architecture

| Component | Technology | Description |
|-----------|------------|-------------|
| Backend API | Go + Gin | Collects system data and provides REST endpoints |
| System Library | gopsutil | Retrieves CPU, memory, network, and process data |
| Frontend UI | React.js | Interactive dashboard with real-time charts |
| Charts | Chart.js | Data visualization for metrics |

## API Endpoints

| Endpoint | Method | Description | Response Format |
|----------|--------|-------------|-----------------|
| `/api/cpu` | GET | CPU load and usage | `{"load1": float, "load5": float, "load15": float, "usage_percent": float}` |
| `/api/memory` | GET | Memory usage statistics | `{"total": uint64, "used": uint64, "free": uint64, "used_percent": float}` |
| `/api/network` | GET | Network traffic data | `{"bytes_sent": uint64, "bytes_recv": uint64, "packets_sent": uint64, "packets_recv": uint64}` |
| `/api/process` | GET | Process resource usage | `[{"pid": int, "name": string, "cpu_percent": float, "mem_percent": float}, ...]` |
| `/api/tasks` | GET | List all tasks | `[{"id": string, "name": string, "type": string, "enabled": bool, ...}, ...]` |
| `/api/tasks` | POST | Create a new task | Request: Task configuration, Response: Created task |
| `/api/tasks/:id` | GET | Get a specific task | `{"id": string, "name": string, "type": string, "enabled": bool, ...}` |
| `/api/tasks/:id` | PUT | Update a task | Request: Updated task, Response: Updated task |
| `/api/tasks/:id` | DELETE | Delete a task | `{"success": bool, "message": string}` |
| `/api/tasks/:id/executions` | GET | Get task execution history | `[{"id": string, "task_id": string, "status": string, "start_time": string, ...}, ...]` |
| `/api/tasks/:id/run` | POST | Run a task immediately | `{"id": string, "task_id": string, "status": string, ...}` |
| `/health` | GET | Health check | `{"status": "healthy"}` |

## Installation & Usage

### Prerequisites

- Go 1.19 or later
- Modern web browser

### Running the Application

#### Option 1: Using Makefile (Recommended)

1. **Clone and navigate to the project:**

   ```bash
   git clone <repository-url>
   cd argus
   ```

2. **View available commands:**

   ```bash
   make help
   ```

3. **Quick start (build and run):**

   ```bash
   make start
   ```

4. **Development mode (with auto-reload):**

   ```bash
   make dev
   ```

5. **Access the web interface:**
   Open your browser to [http://localhost:8080](http://localhost:8080)

#### Option 2: Using Go directly

1. **Install dependencies:**

   ```bash
   go mod tidy
   ```

2. **Run the application:**

   ```bash
   go run cmd/argus/main.go
   ```

3. **Access the web interface:**
   Open your browser to [http://localhost:8080](http://localhost:8080)

### Testing the API

You can test the API endpoints using the provided test script:

```bash
chmod +x test_api.sh
./test_api.sh
```

Or manually test endpoints:

```bash
# Health check
curl http://localhost:8080/health

# CPU information
curl http://localhost:8080/api/cpu

# Memory statistics
curl http://localhost:8080/api/memory

# Network statistics
curl http://localhost:8080/api/network

# Process list
curl http://localhost:8080/api/process
```

## Makefile Commands

The project includes a comprehensive Makefile with many useful targets:

### Development Commands

- `make run` - Run the application in development mode
- `make dev` - Run with auto-reload (requires air)
- `make start` - Quick start (build and run)
- `make build` - Build the binary
- `make clean` - Clean build artifacts

### Testing & Quality

- `make test` - Run all tests
- `make test-race` - Run tests with race detector
- `make test-cover` - Run tests with coverage
- `make bench` - Run benchmarks
- `make lint` - Run linters
- `make fmt` - Format Go code
- `make vet` - Run go vet
- `make api-test` - Test API endpoints

### Dependencies

- `make mod-tidy` - Tidy up module dependencies
- `make deps-check` - Check for outdated dependencies
- `make deps-update` - Update all dependencies

### Release & Deployment

- `make release-build` - Build release binaries for multiple platforms
- `make docker-build` - Build Docker image
- `make docker-run` - Run Docker container
- `make install` - Install the binary to GOPATH/bin

### Setup & Tools

- `make setup-dev` - Setup development environment
- `make git-hooks` - Setup git hooks
- `make security` - Run security checks
- `make all` - Run quality checks and build
- `make ci` - Continuous integration checks

## Project Structure

```
argus/
├── cmd/argus/main.go          # Main application entry point
├── internal/                  # Internal packages
│   ├── api/                  # API handlers
│   │   ├── tasks.go         # Task management API
│   ├── tasks/                # Task management system
│   │   ├── types.go         # Task types and models
│   │   ├── runner.go        # Task runners
│   │   ├── scheduler.go     # Task scheduler
│   │   └── repository/      # Task persistence
│   │       └── repository.go # File-based storage
├── webapp/                    # Frontend assets
│   ├── index.html            # Main HTML file
│   ├── app.js                # React application
│   └── alerts.js             # Alert handling
├── docs/argus_prd.md         # Product Requirements Document
├── test_api.sh               # API testing script
├── Makefile                  # Build automation and development tasks
├── go.mod                    # Go module dependencies
└── README.md                 # This file
```

## Dashboard Features

### CPU Monitoring

- Real-time CPU usage percentage
- Load averages (1, 5, 15 minutes)
- Historical CPU usage chart

### Memory Monitoring

- Total, used, and free memory
- Memory usage percentage
- Visual memory distribution chart

### Network Monitoring

- Bytes sent/received
- Packets sent/received
- Network traffic trends

### Process Monitoring

- Top 20 processes by resource usage
- Sortable by PID, name, CPU%, or memory%
- Real-time process statistics

## Development

### Backend Development

The backend is built with:

- **Gin**: HTTP web framework
- **gopsutil**: System and process utilities
- **CORS**: Cross-origin resource sharing support

Key features:

- RESTful API design
- Error handling and logging
- Static file serving for frontend
- Health check endpoint

### Frontend Development

The frontend uses:

- **React.js**: Component-based UI library
- **Chart.js**: Chart and graph visualization
- **CSS Grid & Flexbox**: Responsive layout
- **Fetch API**: HTTP client for backend communication

Key features:

- Real-time data updates (5-second intervals)
- Interactive charts and visualizations
- Responsive design for mobile/desktop
- Error handling and loading states

## Configuration

### Server Configuration

- Default port: 8080
- CORS enabled for development
- Static files served from `/webapp`

### Data Update Frequency

- Frontend polls backend every 5 seconds
- Charts maintain 20 data points (100 seconds of history)
- Process list limited to top 20 processes

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Task Management

The system includes a task management system for scheduling and running system maintenance tasks. Tasks are defined by their configuration and can run on a schedule or be triggered manually.

### Supported Task Types

1. **Log Rotation (`log_rotation`)**
   - Rotates and archives log files based on size
   - Parameters: 
     - `log_dir` - Directory containing log files
     - `pattern` - File pattern to match (e.g., "*.log")
     - `max_size_mb` - Maximum file size before rotation
     - `keep_count` - Number of rotated logs to retain

2. **Metrics Aggregation (`metrics_aggregation`)**
   - Collects and aggregates system metrics
   - Parameters:
     - `metrics_dir` - Directory for metrics storage
     - `retention_days` - Days to keep metrics data

3. **Health Check (`health_check`)**
   - Monitors system and service health
   - Parameters:
     - `url` - URL to check (HTTP/HTTPS)
     - `timeout` - Request timeout in seconds

4. **System Cleanup (`system_cleanup`)**
   - Purges old temporary files
   - Parameters:
     - `cleanup_dir` - Directory to clean
     - `pattern` - File pattern to match (e.g., "*.tmp")
     - `retention_days` - Days to keep files

### Task Configuration

Tasks use [cron expressions](https://en.wikipedia.org/wiki/Cron) to schedule recurring execution. For example:

- `*/15 * * * *` - Run every 15 minutes
- `0 2 * * *` - Run at 2:00 AM daily
- `0 0 * * 0` - Run at midnight on Sundays

### Sample Task Configuration

```json
{
  "name": "Daily Log Rotation",
  "description": "Rotate log files daily at midnight",
  "type": "log_rotation",
  "enabled": true,
  "schedule": {
    "cron_expression": "0 0 * * *",
    "one_time": false
  },
  "parameters": {
    "log_dir": "/var/log/argus",
    "pattern": "*.log",
    "max_size_mb": "10",
    "keep_count": "7"
  }
}
```

### Creating a Task

To create a new task, send a POST request to `/api/tasks`:

```bash
curl -X POST http://localhost:8080/api/tasks \
  -H "Content-Type: application/json" \
  -d '{
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
  }'
```

### Running a Task Immediately

To run a task immediately regardless of its schedule:

```bash
curl -X POST http://localhost:8080/api/tasks/task-id-here/run
```
