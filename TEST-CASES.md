# TravelMate Test Cases Documentation

This document describes the test cases implemented for the TravelMate project and provides instructions on how to execute them.

## Test Structure

All tests are located in the `tests/` directory and belong to the `tests` package. This structure allows for external testing of the application's components.

| File | Type | Description |
| :--- | :--- | :--- |
| `user_service_test.go` | Unit (Mock) | Tests user registration logic and duplicate email prevention. |
| `trip_service_test.go` | Unit (Mock) | Tests trip creation validation (empty titles, invalid dates). |
| `user_repository_test.go` | Integration | Tests database CRUD operations using an **in-memory SQLite**. |
| `recommendation_server_test.go` | Logic (Mock) | Tests gRPC recommendation and budget analysis logic. |
| `tcp_server_test.go` | Integration | Tests TCP chat server connectivity and welcome message. |
| `grpc_integration_test.go` | E2E/Integration | Verifies full gRPC communication (requires running server). |

## How to Run Tests

### 1. Run All Tests
To run all tests in the suite, use the following command in the project root:
```bash
go test -v ./tests/
```

### 2. Run Specific Test Groups
If you want to run only a specific component's tests:
```bash
# User Service and Repo
go test -v tests/user_service_test.go tests/user_repository_test.go

# gRPC Logic entries
go test -v tests/recommendation_server_test.go
```

### 3. Running gRPC Integration Test
The `TestGRPCIntegration` (in `grpc_test.go`) connects to a live gRPC server. 
1. Start your server: `go run cmd/web/main.go`
2. Run the test: `go test -v -run TestGRPCIntegration ./tests/`

## Key Testing Technologies
- **Testify**: Used for assertions (`assert`) and mocking (`mock`).
- **GORM + SQLite (In-Memory)**: Used to test database interactions without persistent storage.
- **Context & httptest**: Used for testing time-sensitive and network operations.

## Adding New Tests
When adding new tests to the `tests/` directory:
1. Ensure the filename ends with `_test.go`.
2. Use `package tests`.
3. Import the necessary internal packages (e.g., `travel-platform/internal/services`).
4. Use the package prefix when calling functions (e.g., `services.NewUserService`).
