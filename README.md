# Argus

Argus is a simple Linux system performance monitoring web application. The backend is written in Go using Gin and gopsutil, while the frontend is a lightweight React app. It exposes several REST APIs and periodically fetches system metrics for display.

## Backend

### Building and Running

```bash
go run ./cmd/argus
```

The server listens on port `8080` and provides the following endpoints:

- `GET /api/cpu`
- `GET /api/memory`
- `GET /api/network`
- `GET /api/process`

## Frontend

A very small React application is provided in `webapp/`. Open `webapp/index.html` in a browser while the backend is running to view the metrics (the page uses CDN hosted scripts).

## Development

This repository requires Go 1.20+ and Node (for the frontend scripts). Dependencies are referenced but not vendored. Running `go mod tidy` or installing Node packages may require Internet access.

