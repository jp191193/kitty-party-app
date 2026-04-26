# Kitty Party App – REST API

A clean-architecture Go REST API for managing Kitty Party groups.

## Tech Stack

| Concern         | Library / Tool        |
| --------------- | --------------------- |
| HTTP Framework  | [Gin](https://github.com/gin-gonic/gin) |
| Structured Logs | [Uber Zap](https://github.com/uber-go/zap) |
| UUID Generation | [google/uuid](https://github.com/google/uuid) |
| Config / .env   | [godotenv](https://github.com/joho/godotenv) |

## Project Layout

```
kitty-party-app/
├── cmd/
│   └── api/
│       └── main.go          # Entry point & dependency wiring
├── internal/
│   ├── apperrors/           # Domain error types
│   ├── config/              # Environment-based config loader
│   ├── logger/              # Zap logger factory
│   ├── member/              # Member domain (model, repo, service, handler)
│   ├── middleware/           # Gin middleware (logger, recovery, CORS)
│   ├── response/            # Consistent JSON response helpers
│   ├── router/              # Route registration
│   └── server/              # Graceful HTTP server
├── .env.example
├── .gitignore
├── go.mod
├── go.sum
└── Makefile
```

## Quick Start

```bash
# 1. Copy environment config
cp .env.example .env

# 2. Run the server (default port 8080)
go run ./cmd/api

# Or using make
make run
```

## API Endpoints

All routes are prefixed with `/api/v1`.

| Method   | Path               | Description         |
| -------- | ------------------ | ------------------- |
| `GET`    | `/health`          | Health check        |
| `GET`    | `/api/v1/members`  | List all members    |
| `POST`   | `/api/v1/members`  | Create a member     |
| `GET`    | `/api/v1/members/:id` | Get member by ID |
| `PUT`    | `/api/v1/members/:id` | Update a member  |
| `DELETE` | `/api/v1/members/:id` | Delete a member  |

### Example – Create Member

```bash
curl -X POST http://localhost:8080/api/v1/members \
  -H "Content-Type: application/json" \
  -d '{"name":"Anjali Sharma","email":"anjali@example.com","phone":"9876543210"}'
```

### Example Response

```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Anjali Sharma",
    "email": "anjali@example.com",
    "phone": "9876543210",
    "created_at": "2026-04-13T10:00:00Z",
    "updated_at": "2026-04-13T10:00:00Z"
  }
}
```

## Running Tests

```bash
go test -race -v ./...
# or
make test
```

## Adding a New Domain

1. Create `internal/<domain>/model.go` – DTOs & entity
2. Create `internal/<domain>/repository.go` – `Repository` interface + implementation
3. Create `internal/<domain>/service.go` – `Service` interface + implementation
4. Create `internal/<domain>/handler.go` – `Handler` + `RegisterRoutes`
5. Wire handler in `internal/router/router.go`
6. Inject repo → service → handler in `cmd/api/main.go`
