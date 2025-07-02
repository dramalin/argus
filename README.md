# Argus System Monitor

A real-time Linux system performance monitoring web application built with Go (Gin) backend and React.js frontend.

## Features

- **Real-time Monitoring**: CPU usage, memory statistics, network traffic, and process information
- **Interactive Charts**: Visual representation of system metrics using Chart.js
- **Responsive Design**: Modern, mobile-friendly interface
- **RESTful API**: Clean API endpoints for system data
- **Process Management**: Sortable process table with CPU and memory usage

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
├── webapp/                    # Frontend assets
│   ├── index.html            # Main HTML file
│   └── app.js                # React application
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
