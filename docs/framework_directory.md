# Web Application File Storage Architecture Framework

## Overview

This framework defines a standardized file storage architecture for web applications using:

- **Backend**: Go with Gin framework
- **Frontend**: React with Vite framework
- **Integration**: WebSocket support, external API integrations
- **Deployment**: Docker containerization

## Core Architecture Principles

### 1. Separation of Concerns

- **Backend** (`/` root): Go-based server application
- **Frontend** (`/web/`): React-based client applications
- **Configuration** (`/config.*.yaml`): Environment-specific settings
- **Documentation** (`/docs/`): Project specifications and guides

### 2. Internal vs External Code

- **Internal** (`/internal/`): Private application code (Go modules)
- **External Integrations** (`/[servicename]/`): Third-party service clients
- **Stubs** (`/stubs/`): Testing mocks and development utilities

### 3. Multiple Frontend Support

- **Modern SPA** (`/web/[appname]-react/`): React applications with Vite
- **Static Assets** (`/web/static/`): Traditional HTML/CSS/JS files

## Directory Structure Template

```
project-root/
├── go.mod                           # Go module definition
├── go.sum                           # Go dependency lock file
├── Makefile                         # Build automation scripts
├── Dockerfile                       # Container definition
├── docker-compose.yml              # Multi-container orchestration
├── config.example.yaml             # Configuration template
├── LICENSE                          # License file
├── README.md                        # Project documentation
│
├── cmd/                             # web server command
│   └── [project]/                   # project
│       └── main.go                  # Application entry point
│
├── bin/                             # Compiled binaries (gitignored)
│
├── docs/                            # Project documentation
│   ├── [project]_prd.md            # Product Requirements Document
│   ├── [project]_ui_prd.md         # UI/UX Requirements Document
│   └── api_documentation.md        # API specifications (optional)
│
├── internal/                        # Private application code
│   ├── config/                      # Configuration management
│   │   ├── config.go               # Configuration structure definitions
│   │   ├── config_test.go          # Configuration tests
│   │   └── timezone.go             # Timezone handling utilities
│   │
│   ├── server/                     # HTTP server components
│   │   ├── server.go               # Main server setup and routing
│   │   ├── middleware.go           # HTTP middleware (auth, logging, etc.)
│   │   └── websocket.go            # WebSocket connection handling
│   │
│   ├── services/                   # Business logic services
│   │   ├── service.go              # Core business logic
│   │   ├── service_test.go         # Service tests
│   │   └── notifier.go             # Event notification system
│   │
│   ├── handlers/                    # HTTP request handlers (recommended)
│   │   ├── auth.go                 # Authentication handlers
│   │   ├── api.go                  # API endpoint handlers
│   │   └── health.go               # Health check endpoints
│   │
│   ├── models/                      # Data models and structures
│   │   ├── user.go                 # User data models
│   │   ├── event.go                # Event data models
│   │   └── response.go             # API response structures
│   │
│   └── database/                    # Database layer (if applicable)
│       ├── migrations/             # Database migration files
│       ├── connection.go           # Database connection management
│       └── queries.go              # Database query functions
│
├── [external-service]/              # External service integrations
│   ├── client.go                   # Service client implementation
│   ├── client_test.go              # Client tests
│   ├── client_benchmark_test.go    # Performance benchmarks
│   ├── config.go                   # Service-specific configuration
│   ├── oauth.go                    # OAuth authentication
│   ├── [resource1].go              # Resource-specific operations
│   ├── [resource1]_test.go         # Resource tests
│   ├── [resource2].go              # Additional resource operations
│   └── [resource2]_test.go         # Additional resource tests
│
├── stubs/                           # Testing and development utilities
│   └── [dependency]/               # Stubbed dependencies
│       ├── go.mod                  # Stub module definition
│       └── [stubfile].go           # Stub implementations
│
└── web/                             # Frontend applications and assets
    ├── [app-name]-react/           # React application with Vite
    │   ├── package.json            # Node.js dependencies
    │   ├── package-lock.json       # Dependency lock file
    │   ├── vite.config.ts          # Vite configuration
    │   ├── tsconfig.json           # TypeScript configuration
    │   ├── tsconfig.node.json      # Node-specific TypeScript config
    │   ├── eslint.config.js        # ESLint configuration
    │   ├── index.html              # HTML entry point
    │   ├── README.md               # Frontend-specific documentation
    │   ├── build_output.txt        # Build logs (gitignored)
    │   │
    │   ├── public/                 # Static public assets
    │   │   └── vite.svg           # Framework assets
    │   │
    │   └── src/                    # React source code
    │       ├── main.tsx           # React application entry point
    │       ├── App.tsx            # Main application component
    │       ├── Dashboard.tsx      # Dashboard page component
    │       ├── api.ts             # API client functions
    │       ├── theme.ts           # Theme configuration
    │       ├── ThemeContext.tsx   # Theme context provider
    │       │
    │       ├── assets/            # Static assets
    │       │   ├── [project]_logo.png
    │       │   └── react.svg
    │       │
    │       ├── components/        # Reusable React components
    │       │   ├── README.md      # Component documentation
    │       │   ├── Layout.tsx     # Layout wrapper component
    │       │   ├── LoadingOverlay.tsx
    │       │   ├── ConfirmationDialog.tsx
    │       │   ├── [Feature]Form.tsx
    │       │   ├── [Feature]NotificationBar.tsx
    │       │   └── GridItem.tsx
    │       │
    │       ├── hooks/             # Custom React hooks
    │       │   ├── use[Feature]Data.ts
    │       │   └── useWebSocket.ts
    │       │
    │       ├── types/             # TypeScript type definitions
    │       │   ├── api.ts         # API-related types
    │       │   └── declarations.d.ts
    │       │
    │       ├── theme/             # Theme configuration
    │       │   └── index.ts
    │       │
    │       ├── utils/             # Utility functions
    │       │
    │       ├── [FeatureView].tsx  # Feature-specific view components
    │       └── [FeatureView].d.ts # View component type definitions
    │
    └── static/                     # Traditional static web assets
        ├── index.html             # Static HTML entry point
        ├── favicon.ico            # Site favicon
        │
        ├── assets/                # Static images and media
        │   └── [project]_logo.png
        │
        ├── css/                   # Stylesheets
        │   ├── main.css           # Main stylesheet
        │   ├── [feature].css      # Feature-specific styles
        │   └── combined-view.css  # Layout-specific styles
        │
        └── js/                    # JavaScript files
            ├── api.js             # API client functions
            ├── websocket-test.js  # WebSocket testing utilities
            ├── [feature].js       # Feature-specific scripts
            └── [feature]-view.js  # View-specific scripts
```

## File Naming Conventions

### Backend (Go)

- **Entry Point**: `main.go`
- **Packages**: `lowercase` (e.g., `config`, `server`, `sync`)
- **Files**: `snake_case.go` (e.g., `user_service.go`, `auth_handler.go`)
- **Tests**: `*_test.go` (e.g., `config_test.go`, `service_test.go`)
- **Benchmarks**: `*_benchmark_test.go`

### Frontend (React/Vite)

- **Components**: `PascalCase.tsx` (e.g., `UserProfile.tsx`, `EventForm.tsx`)
- **Hooks**: `use[Name].ts` (e.g., `useAuth.ts`, `useCalendarData.ts`)
- **Types**: `camelCase.ts` or `declarations.d.ts`
- **Utilities**: `camelCase.ts` (e.g., `dateUtils.ts`, `apiClient.ts`)
- **Styles**: `kebab-case.css` (e.g., `user-profile.css`, `main-layout.css`)

### Configuration

- **Main Config**: `config.example.yaml` (template), `config.yaml` (actual, gitignored)
- **Docker**: `Dockerfile`, `docker-compose.yml`
- **Build**: `Makefile`, `package.json`, `go.mod`

## Technology Stack Integration

### Backend Stack

```go
// go.mod example
module project-name

go 1.21

require (
    github.com/gin-gonic/gin v1.9.1
    github.com/gorilla/websocket v1.5.0
    gopkg.in/yaml.v3 v3.0.1
    // Add other dependencies
)
```

### Frontend Stack

```json
// package.json example
{
  "name": "project-name-react",
  "private": true,
  "version": "0.0.0",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "tsc && vite build",
    "preview": "vite preview",
    "lint": "eslint . --ext ts,tsx --report-unused-disable-directives --max-warnings 0"
  },
  "dependencies": {
    "react": "^18.2.0",
    "react-dom": "^18.2.0"
  },
  "devDependencies": {
    "@types/react": "^18.2.15",
    "@types/react-dom": "^18.2.7",
    "@typescript-eslint/eslint-plugin": "^6.0.0",
    "@typescript-eslint/parser": "^6.0.0",
    "@vitejs/plugin-react": "^4.0.3",
    "eslint": "^8.45.0",
    "eslint-plugin-react-hooks": "^4.6.0",
    "eslint-plugin-react-refresh": "^0.4.3",
    "typescript": "^5.0.2",
    "vite": "^4.4.5"
  }
}
```

## Configuration Management

### Environment Configuration

```yaml
# config.example.yaml
server:
  port: 8080
  host: "localhost"
  
database:
  host: "localhost"
  port: 5432
  name: "project_db"
  
external_services:
  service_name:
    client_id: "your_client_id"
    tenant_id: "your_tenant_id"
    base_url: "https://api.service.com"

logging:
  level: "info"
  format: "json"
```

### Go Configuration Structure

```go
// internal/config/config.go
type Config struct {
    Server struct {
        Port int    `yaml:"port"`
        Host string `yaml:"host"`
    } `yaml:"server"`
    
    Database struct {
        Host string `yaml:"host"`
        Port int    `yaml:"port"`
        Name string `yaml:"name"`
    } `yaml:"database"`
    
    ExternalServices map[string]ServiceConfig `yaml:"external_services"`
    
    Logging struct {
        Level  string `yaml:"level"`
        Format string `yaml:"format"`
    } `yaml:"logging"`
}
```

## Development Workflow

### Development Commands

```makefile
# Makefile example
.PHONY: build test run clean dev

# Build the application
build:
 go build -o bin/app main.go

# Run tests
test:
 go test ./...

# Run the application in development mode
dev:
 go run main.go

# Run frontend development server
dev-frontend:
 cd web/calendar-react && npm run dev

# Build frontend for production
build-frontend:
 cd web/calendar-react && npm run build

# Clean build artifacts
clean:
 rm -rf bin/
 rm -rf web/calendar-react/dist/

# Install dependencies
deps:
 go mod download
 cd web/calendar-react && npm install
```

### Git Ignore Patterns

```gitignore
# Binaries
bin/
*.exe

# Dependencies
node_modules/
vendor/

# Build outputs
dist/
build/
*.log

# Configuration
config.yaml
*.env
tokens.json

# IDE
.vscode/
.idea/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db
```

## API Integration Patterns

### External Service Client Structure

```go
// [servicename]/client.go
package servicename

type Client struct {
    httpClient *http.Client
    config     *Config
    logger     *log.Logger
}

func NewClient(config *Config) *Client {
    return &Client{
        httpClient: &http.Client{Timeout: 30 * time.Second},
        config:     config,
        logger:     log.New(os.Stdout, "[SERVICENAME] ", log.LstdFlags),
    }
}

func (c *Client) GetResource(ctx context.Context, id string) (*Resource, error) {
    // Implementation
}
```

### API Response Structures

```go
// internal/models/response.go
type APIResponse struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
    Meta    *Meta       `json:"meta,omitempty"`
}

type Meta struct {
    Total      int    `json:"total"`
    Page       int    `json:"page"`
    PerPage    int    `json:"per_page"`
    Timestamp  string `json:"timestamp"`
}
```

## Frontend Architecture Patterns

### Component Organization

```typescript
// src/components/Layout.tsx
import React from 'react';
import { Outlet } from 'react-router-dom';

interface LayoutProps {
  children?: React.ReactNode;
}

export const Layout: React.FC<LayoutProps> = ({ children }) => {
  return (
    <div className="layout">
      <header className="layout-header">
        {/* Header content */}
      </header>
      <main className="layout-main">
        {children || <Outlet />}
      </main>
      <footer className="layout-footer">
        {/* Footer content */}
      </footer>
    </div>
  );
};
```

### Custom Hooks Pattern

```typescript
// src/hooks/useWebSocket.ts
import { useState, useEffect, useRef } from 'react';

interface UseWebSocketOptions {
  onMessage?: (data: any) => void;
  onError?: (error: Event) => void;
  reconnectInterval?: number;
}

export const useWebSocket = (url: string, options: UseWebSocketOptions = {}) => {
  const [isConnected, setIsConnected] = useState(false);
  const [lastMessage, setLastMessage] = useState(null);
  const ws = useRef<WebSocket | null>(null);

  useEffect(() => {
    // WebSocket connection logic
  }, [url]);

  const sendMessage = (message: any) => {
    if (ws.current && isConnected) {
      ws.current.send(JSON.stringify(message));
    }
  };

  return { isConnected, lastMessage, sendMessage };
};
```

### API Client Pattern

```typescript
// src/api.ts
const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: string;
}

class ApiClient {
  private baseURL: string;

  constructor(baseURL: string = API_BASE_URL) {
    this.baseURL = baseURL;
  }

  async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<ApiResponse<T>> {
    const url = `${this.baseURL}${endpoint}`;
    
    try {
      const response = await fetch(url, {
        headers: {
          'Content-Type': 'application/json',
          ...options.headers,
        },
        ...options,
      });

      const data = await response.json();
      return data;
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Unknown error',
      };
    }
  }

  async get<T>(endpoint: string): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, { method: 'GET' });
  }

  async post<T>(endpoint: string, body: any): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, {
      method: 'POST',
      body: JSON.stringify(body),
    });
  }
}

export const apiClient = new ApiClient();
```

## Testing Patterns

### Backend Testing

```go
// internal/config/config_test.go
package config

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestConfigLoad(t *testing.T) {
    config, err := LoadConfig("testdata/config.yaml")
    assert.NoError(t, err)
    assert.Equal(t, 8080, config.Server.Port)
}

func BenchmarkConfigLoad(b *testing.B) {
    for i := 0; i < b.N; i++ {
        LoadConfig("testdata/config.yaml")
    }
}
```

### Frontend Testing Structure

```typescript
// src/components/__tests__/Layout.test.tsx
import { render, screen } from '@testing-library/react';
import { Layout } from '../Layout';

describe('Layout', () => {
  it('renders children correctly', () => {
    render(
      <Layout>
        <div data-testid="content">Test content</div>
      </Layout>
    );
    
    expect(screen.getByTestId('content')).toBeInTheDocument();
  });
});
```

## Docker Configuration

### Multi-stage Dockerfile

```dockerfile
# Build stage for Go
FROM golang:1.21-alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .

# Build stage for Node.js
FROM node:18-alpine AS node-builder
WORKDIR /app/web/calendar-react
COPY web/calendar-react/package*.json ./
RUN npm ci
COPY web/calendar-react/ ./
RUN npm run build

# Final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=go-builder /app/main .
COPY --from=node-builder /app/web/calendar-react/dist ./web/static/
COPY config.example.yaml ./config.yaml
EXPOSE 8080
CMD ["./main"]
```

### Docker Compose

```yaml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - CONFIG_FILE=/app/config.yaml
    volumes:
      - ./config.yaml:/app/config.yaml
    depends_on:
      - database

  database:
    image: postgres:15
    environment:
      POSTGRES_DB: project_db
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

volumes:
  postgres_data:
```

## AI Agent Guidelines

### Project Initialization Checklist

1. **Create basic directory structure** following this framework
2. **Set up Go module** with `go mod init`
3. **Initialize React project** with Vite in `web/[appname]-react/`
4. **Configure build tools** (Makefile, Docker, CI/CD)
5. **Set up configuration management** with example files
6. **Implement basic server structure** with routing and middleware
7. **Create API client** for frontend-backend communication
8. **Set up WebSocket support** if real-time features are needed

### Code Organization Principles

1. **Separation of concerns**: Keep business logic, HTTP handlers, and external integrations separate
2. **Consistent naming**: Follow Go and React/TypeScript naming conventions
3. **Error handling**: Implement comprehensive error handling and logging
4. **Testing**: Write tests for all major components and business logic
5. **Documentation**: Maintain clear documentation for APIs and complex logic

### Integration Patterns

1. **External services**: Create dedicated packages for each external service
2. **Database layer**: Abstract database operations into dedicated packages
3. **Configuration**: Use structured configuration with environment overrides
4. **Logging**: Implement structured logging throughout the application
5. **Health checks**: Include health check endpoints for monitoring

### Performance Considerations

1. **Connection pooling**: Use connection pools for databases and HTTP clients
2. **Caching**: Implement appropriate caching strategies
3. **Asset optimization**: Optimize frontend assets with Vite build tools
4. **Concurrent processing**: Use Go routines for concurrent operations
5. **Resource management**: Properly manage file handles and network connections

## Security Best Practices

### Authentication & Authorization

```go
// internal/middleware/auth.go
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.JSON(401, gin.H{"error": "Authorization header required"})
            c.Abort()
            return
        }
        
        // Validate token
        if !validateToken(token) {
            c.JSON(401, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

### Environment Variables

```go
// internal/config/security.go
func LoadSecureConfig() (*Config, error) {
    config := &Config{}
    
    // Load from environment variables for sensitive data
    config.Database.Password = os.Getenv("DB_PASSWORD")
    config.ExternalServices.APIKey = os.Getenv("EXTERNAL_API_KEY")
    
    // Load non-sensitive data from config file
    if err := yaml.Unmarshal(configData, config); err != nil {
        return nil, err
    }
    
    return config, nil
}
```

This framework provides a comprehensive foundation for building scalable web applications with Go/Gin backend and React/Vite frontend, designed specifically for AI agents to understand and implement consistently.
