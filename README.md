# Infra Library

The `infra` library provides foundational utilities and middleware for building Go services. It includes logging, observability, HTTP middleware, and configuration management.

**Module**: `github.com/creastat/infra`

## 📦 Components

### Telemetry (`telemetry/`)
Observability and telemetry utilities with structured logging.

**Features**:
- Structured logging with zerolog
- Context-aware logging with correlation IDs
- Asynq queue logger integration
- Configurable log levels and formats
- Service instrumentation

**Usage**:
```go
import "github.com/creastat/infra/telemetry"

logger := telemetry.New(telemetry.Config{
    Level: "debug",
    Format: "text",
    ServiceName: "my-service",
    Environment: "development",
})

logger.Info("Application started", 
    telemetry.String("version", "1.0.0"),
    telemetry.Int("port", 8080),
)

// With context
ctx := telemetry.ContextWithRequestID(context.Background(), "req-123")
logger.WithContext(ctx).Info("Processing request")
```

### Middleware (`middleware/`)
HTTP middleware utilities.

**Included Middleware**:
- **CORS**: Cross-Origin Resource Sharing configuration
- **Logging**: HTTP request/response logging with correlation IDs
- **Recovery**: Panic recovery with logging

**Usage**:
```go
import "github.com/creastat/infra/middleware"

// CORS middleware
corsMiddleware := middleware.NewCORSMiddleware(middleware.CORSConfig{
    AllowedOrigins: []string{"*"},
    AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
})

// Logging middleware
loggingMiddleware := middleware.NewLoggingMiddleware(logger)

// Recovery middleware
recoveryMiddleware := middleware.NewRecoveryMiddleware(logger)
```

### HTTP (`http/`)
HTTP utilities for error handling and responses.

**Features**:
- Standardized error responses
- Response formatting
- HTTP status code mapping

**Usage**:
```go
import "github.com/creastat/infra/http"

// Error response
http.Error(w, "Not found", http.StatusNotFound)

// Success response
http.JSON(w, http.StatusOK, data)
```

### Config (`config/`)
Configuration management utilities.

**Features**:
- YAML/environment variable configuration
- Configuration validation
- Configuration watching for hot-reload

**Usage**:
```go
import "github.com/creastat/infra/config"

cfg := config.Load("config.yaml")
```

### Types (`types/`)
Shared type definitions.

**Includes**:
- Logger types and interfaces
- Configuration types

## 🚀 Quick Start

### Installation

```bash
go get github.com/creastat/infra@latest
```

### Basic Setup

```go
package main

import (
    "github.com/creastat/infra/telemetry"
    "github.com/creastat/infra/middleware"
    "net/http"
)

func main() {
    // Initialize logger
    logger := telemetry.New(telemetry.Config{
        Level: "info",
        ServiceName: "my-service",
    })

    // Create HTTP server with middleware
    mux := http.NewServeMux()
    
    // Apply middleware
    handler := middleware.NewRecoveryMiddleware(logger)(
        middleware.NewLoggingMiddleware(logger)(mux),
    )

    logger.Info("Starting server on :8080")
    http.ListenAndServe(":8080", handler)
}
```

## 📋 Directory Structure

```
services/libraries/base/
├── telemetry/           # Observability & logging
│   ├── logger.go        # Structured logging with zerolog, Logger interface, NoOpLogger
│   └── writer.go        # Telemetry writer
├── middleware/          # HTTP middleware
│   ├── cors.go          # CORS configuration
│   ├── logging.go       # Request/response logging
│   └── recovery.go      # Panic recovery
├── http/                # HTTP utilities
│   ├── errors.go        # Error handling
│   └── response.go      # Response formatting
├── config/              # Configuration management
│   ├── base.go
│   ├── config.go
│   └── loader.go
├── go.mod
├── go.sum
└── README.md
```

## 🔧 Configuration

### Logger Configuration

```go
type Config struct {
    Level        string // "debug", "info", "warn", "error"
    Format       string // "json", "text"
    EnableCaller bool
    ServiceName  string
    Environment  string
}
```

### CORS Configuration

```go
type CORSConfig struct {
    AllowedOrigins   []string
    AllowedMethods   []string
    AllowedHeaders   []string
    ExposedHeaders   []string
    AllowCredentials bool
    MaxAge           int
}
```

## 🎯 Use Cases

- **Logging**: Structured logging across all services
- **Middleware**: Common HTTP middleware (CORS, logging, recovery)
- **Configuration**: Centralized configuration management
- **Telemetry**: Service instrumentation and observability
- **Error Handling**: Standardized HTTP error responses

## 📚 Related Libraries

- **providers**: AI provider integration library
- **platform**: Service-specific platform utilities

## 🤝 Contributing

When adding new utilities to the base library:

1. Ensure they are not service-specific
2. Add comprehensive documentation
3. Include usage examples
4. Update this README
5. Run `go mod tidy` and `go build ./...`

## 📝 Notes

- This library contains only reusable, non-service-specific utilities
- Service-specific code should remain in individual services
- Database clients and queue integration are in separate libraries
- Supabase-specific code remains in the platform library
