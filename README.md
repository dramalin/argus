# Argus System Monitor

A real-time Linux system performance monitoring web application built with Go (Gin) backend and React (Vite) frontend.

## Features

- **Real-time Monitoring**: CPU, memory, network, and process metrics
- **Interactive Dashboard**: Modern React UI with live charts
- **RESTful API**: Clean endpoints for system and task management
- **Task Scheduling**: Cron-based system maintenance and health checks
- **Dockerized**: Production-ready multi-stage Docker build
- **Configurable**: YAML-based configuration with environment overrides

## Project Structure

```
argus/
├── cmd/argus/main.go          # Main application entry point
├── internal/                  # Internal packages (config, server, services, handlers, models, database)
├── web/                       # Frontend and static assets
│   ├── argus-react/           # React (Vite) SPA frontend
│   └── static/                # Static HTML/CSS/JS assets
├── Dockerfile                 # Multi-stage Docker build
├── docker-compose.yml         # Dev/prod orchestration
├── config.example.yaml        # Config template
├── Makefile                   # Build and workflow automation
├── go.mod                     # Go module
├── README.md                  # Project documentation
└── docs/                      # Architecture, PRD, and API docs
```

## Development Workflow

### Prerequisites

- Go 1.23+
- Node.js 18+ and npm
- Docker & docker-compose (for containerized workflow)

### Common Makefile Commands

| Command            | Description                                 |
|--------------------|---------------------------------------------|
| `make build`       | Build backend and frontend for production   |
| `make dev`         | Run backend (auto-reload) & frontend (Vite) |
| `make web-dev`     | Run frontend dev server only                |
| `make web-build`   | Build frontend for production               |
| `make clean`       | Clean all build artifacts                   |
| `make deps`        | Install Go and frontend dependencies        |
| `make lint`        | Lint Go code                                |
| `make web-lint`    | Lint frontend code                          |
| `make docker-up`   | Start all services with docker-compose      |
| `make docker-down` | Stop all docker-compose services            |
| `make docker-build`| Build Docker image                          |

### Local Development (Recommended)

1. **Install dependencies:**

   ```bash
   make deps
   ```

2. **Start dev servers (backend + frontend):**

   ```bash
   make dev
   ```

   - Backend: [http://localhost:8080](http://localhost:8080)
   - Frontend: [http://localhost:5173](http://localhost:5173)

3. **Access the app:**
   - Open [http://localhost:5173](http://localhost:5173) for the React dashboard.

### Production Build

1. **Build everything:**

   ```bash
   make build
   ```

2. **Run the backend:**

   ```bash
   ./bin/argus
   ```

   - Serves static assets from `web/static/` (built by frontend)

### Docker Workflow

1. **Build and start all services:**

   ```bash
   make docker-up
   ```

   - Backend: [http://localhost:8080](http://localhost:8080)
   - Frontend (dev): [http://localhost:5173](http://localhost:5173)

2. **Stop all services:**

   ```bash
   make docker-down
   ```

3. **View logs:**

   ```bash
   make docker-logs
   ```

### Testing & Linting

- **Go tests:**

  ```bash
  make test
  ```

- **Frontend lint:**

  ```bash
  make web-lint
  ```

- **Go lint:**

  ```bash
  make lint
  ```

### Configuration

- Copy `config.example.yaml` to `config.yaml` and edit as needed.
- Environment variables can override config values (see docs/framework_directory.md).

### Directory Conventions

- Backend code: `internal/`
- Frontend app: `web/argus-react/`
- Static assets: `web/static/`
- Config: `config.yaml`
- Docs: `docs/`

## API Endpoints

See [docs/api_documentation.md](docs/api_documentation.md) for full API reference.

## License

MIT
