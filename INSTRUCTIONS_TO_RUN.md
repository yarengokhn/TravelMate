# TravelMate - Project Instructions

This document provides comprehensive instructions on how to compile, run and deploy the **TravelMate** project.

## 📋 Prerequisites

Before you begin, ensure you have the following installed on your system:

- **Go (Golang)**: Version 1.25 or higher ([Download](https://golang.org/dl/))
- **Git**: For version control ([Download](https://git-scm.com/))
- **Protocol Buffers (protoc)**: If you plan to modify gRPC services ([Download](https://github.com/protocolbuffers/protobuf/releases))
- **SQLite**: The database is file-based, so no separate server is required.

---

## 🛠️ Environment Setup

1. **Clone the Repository**:
   ```bash
   git clone <repository-url>
   cd TravelMate_Web_App/TravelMate
   ```

2. **Download Dependencies**:
   ```bash
   go mod download
   ```
3. **Compile**:
    ```bash
    go build -o travelmate cmd/web/main.go
    ```
---

## 🚀 Running Locally

The project consists of multiple services that run concurrently within the main web server.

### 1. Start the Main Web Server
This command launches the HTTP web interface, the gRPC recommendation service, and the TCP chat server.
```bash
go run cmd/web/main.go
```
- **Web Interface**: `http://localhost:8080`
- **gRPC Service**: `localhost:50051`
- **TCP Chat Server**: `localhost:9090`

## 🧪 gRPC Service Generation

If you modify the `.proto` files in the `proto/` directory, you need to regenerate the Go code:

```bash
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/recomendation.proto
```

---

## 🚢 Deployment Guide

### 1. Database Management
- The application uses **SQLite**. By default, it creates/uses `travel-platform.db` in the root directory.
- For production, ensure the user running the application has read/write permissions for this file and its directory.

### 2. Production Build
Prepare the application for production by building the binary and collecting static assets.
1. Build the binary for your target OS (as shown in the [Compilation](#-compilation-building-binaries) section).
2. Ensure the `web/` directory (containing `templates` and `static`) is located in the same directory as the executable.

## 📝 Troubleshooting
- **Port Conflicts**: Ensure ports `8080`, `50051`, and `9090` are not being used by other applications.
- **Dependency Issues**: If you encounter import errors, run `go mod tidy`.
- **Database Access**: If you see "database is locked" errors, ensure only one instance of the application is accessing the `.db` file at a time.
