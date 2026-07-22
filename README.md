# DDD Second-Hand Marketplace

A pedagogical **second-hand marketplace** using **Domain-Driven Design (DDD)** and **Hexagonal Architecture** (Ports & Adapters) in Go.

## Prerequisites

- **Go 1.25.3** or higher
- **Docker** (required for integration tests using testcontainers)
- **Git**

## Installation

### 1. Install Go

#### Linux (Ubuntu/Debian)

```bash
# Download Go
wget https://go.dev/dl/go1.25.3.linux-amd64.tar.gz

# Remove previous installation and extract
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.25.3.linux-amd64.tar.gz

# Add to PATH (add to ~/.bashrc or ~/.zshrc for persistence)
export PATH=$PATH:/usr/local/go/bin

# Verify installation
go version
```

#### macOS

```bash
# Using Homebrew
brew install go

# Or download from https://go.dev/dl/
```

#### Windows

Download the installer from [https://go.dev/dl/](https://go.dev/dl/) and follow the installation wizard.

### 2. Install Docker

Docker is required for running integration tests with testcontainers (PostgreSQL, Kafka).

- **Linux**: [https://docs.docker.com/engine/install/](https://docs.docker.com/engine/install/)
- **macOS/Windows**: [https://docs.docker.com/desktop/](https://docs.docker.com/desktop/)

### 3. Install dependencies

```bash
go mod download
```

## Verify Installation

After installing Go and Docker, verify your environment is correctly set up:

```bash
# Build the project
go build -o bin/api ./cmd/api

# Run the server
./bin/api
```

In another terminal, test the health endpoint:

```bash
curl http://localhost:8080/health
```

Expected response:
```json
{"status":"ok"}
```

If you see this response, your environment is ready.

## Build

```bash
go build -o bin/api ./cmd/api
```

## Run

```bash
./bin/api
```

The server runs on http://localhost:8080

## Testing

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for a specific package
go test -v ./internal/search/...

# Run a specific test
go test -run TestName ./path/to/package/

# Run tests with coverage
go test -cover ./...
```

## Code Quality

```bash
# Format code
go fmt ./...

# Vet code (static analysis)
go vet ./...

# Tidy dependencies
go mod tidy
```

## Project Structure

```
internal/{bounded-context}/
├── domain/           # Pure business logic, repository interfaces (ports)
├── application/
│   ├── command/     # Write operations (CQRS)
│   └── query/       # Read operations (CQRS)
└── adapter/
    ├── driven/      # Outbound: repositories, external services
    └── driving/     # Inbound: HTTP handlers, event consumers

pkg/                  # Shared infrastructure (eventbus, mailer)
cmd/api/              # Application entry point
```

## Architecture

This project follows **Hexagonal Architecture** with the dependency rule:

```
Adapter → Application → Domain (never reverse)
```

- **Domain**: Pure business logic with no external dependencies
- **Application**: Use cases (commands/queries) orchestrating domain logic
- **Adapters**: Infrastructure implementations (HTTP, database, messaging)
