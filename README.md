# FreeAPI Hub

> Unified gateway aggregating multiple free public APIs. Built with Go to learn the language end-to-end.

A learning project covering Go fundamentals through production patterns: HTTP services, concurrency, JWT authentication, PostgreSQL, caching, and Docker deployment.

## Tech Stack

| Concern | Package |
|---|---|
| HTTP Router | `chi` |
| HTTP Client | `resty` |
| Database driver | `pgx/v5` |
| SQL code-gen | `sqlc` |
| Migrations | `golang-migrate` |
| JWT | `golang-jwt/jwt/v5` |
| Password hash | `bcrypt` |
| Validation | `go-playground/validator/v10` |
| Config | `envconfig` + `godotenv` |
| Logging | `slog` (stdlib) |
| Redis client | `go-redis/v9` |
| Rate limit | `golang.org/x/time/rate` |
| Concurrency | `golang.org/x/sync/errgroup` |
| Testing | `testify` |
| Mocking | `mockery` |
| Live reload | `air` |

## Project Structure

```
freeapi-hub/
├── cmd/server/              # Application entry point
├── internal/
│   ├── config/              # Env var loading
│   ├── domain/              # Pure business types & errors
│   ├── auth/                # JWT + password hashing
│   ├── middleware/          # Auth, rate limit, logging
│   ├── providers/           # External API wrappers
│   │   ├── provider.go      # Common interface
│   │   ├── weather/
│   │   ├── crypto/
│   │   ├── exchange/
│   │   └── news/
│   ├── aggregator/          # Concurrent multi-provider fetcher
│   ├── cache/               # Cache interface + memory/redis impls
│   ├── storage/             # Database layer (sqlc generated)
│   ├── http/                # Router & response helpers
│   └── user/                # User service & handlers
├── pkg/validator/           # Reusable validation utility
├── migrations/              # Database migrations
├── tests/integration/       # Integration tests
└── ...                      # Makefile, Dockerfile, configs
```

## Quick Start

### 1. Prerequisites

- Go 1.22+
- Docker & Docker Compose
- `make` command available

### 2. Install Development Tools

```bash
make install-tools
```

Installs: `air` (live reload), `migrate` (DB migrations), `golangci-lint`, `mockery`.

Also install `sqlc`:
```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

### 3. Setup Environment

```bash
cp .env.example .env
# Edit .env: generate a real JWT_SECRET
openssl rand -base64 32  # paste output to JWT_SECRET in .env
```

### 4. Start Postgres & Redis

```bash
make docker-up
```

### 5. Run Migrations

```bash
make migrate-up
```

### 6. Start the Server

For development with auto-reload:
```bash
make dev
```

Or plain run:
```bash
make run
```

Test it:
```bash
curl http://localhost:8080/health
```

## Useful Commands

```bash
make help              # List all available commands
make test              # Run unit tests
make test-race         # Run tests with race detector
make lint              # Run golangci-lint
make check             # Run fmt + vet + lint + test-race (pre-commit)
make docker-up         # Start Postgres & Redis
make migrate-up        # Apply migrations
make build             # Build production binary
```

## Learning Roadmap

This codebase is structured around a 6-week curriculum. Each file contains `TODO` markers pointing to the week/session where you implement it.

| Week | Focus | Key Go Concepts |
|---|---|---|
| 1 | Foundations, first endpoint | Modules, structs, JSON, HTTP server/client |
| 2 | Architecture, multiple providers | Interfaces, DI, package layout, slog |
| 3 | Concurrency, dashboard endpoint | Goroutines, channels, context, errgroup |
| 4 | Database, JWT auth | sqlc, pgx, bcrypt, middleware |
| 5 | Cache, rate limit, API keys | Generics, RWMutex, token bucket |
| 6 | Testing, Docker, deployment | testify, mockery, testcontainers, CI |

### Tài liệu học

- **`docs/WEEK1.md`** — Hướng dẫn step-by-step Tuần 1, mỗi buổi có bài tập
- **`docs/LEARNING_NOTES.md`** — Giải thích chi tiết từng đoạn code đã viết sẵn (đọc khi gặp đoạn chưa hiểu)
- **`docs/EXERCISES.md`** — Bài tập "rebuild from scratch" + self-check questions cho mỗi tuần

**Cách dùng đúng**: với mỗi tuần, làm theo flow:
1. Đọc TODO comments trong code stub
2. Implement theo hint
3. Đọc `LEARNING_NOTES.md` section liên quan để hiểu sâu các phần đã viết sẵn
4. Cuối tuần làm bài tập trong `EXERCISES.md` — đặc biệt phần "Rebuild challenge" (xóa code đã có sẵn, tự viết lại)

## API Endpoints (final state)

### Public
- `GET  /health` — health check
- `POST /auth/register` — create account
- `POST /auth/login` — get JWT tokens
- `POST /auth/refresh` — refresh access token

### Protected (require `Authorization: Bearer <token>`)
- `GET /v1/weather?city=Hanoi`
- `GET /v1/crypto?symbol=btc`
- `GET /v1/exchange?from=USD&to=VND&amount=100`
- `GET /v1/news?q=technology`
- `GET /v1/dashboard?city=Hanoi&coins=btc,eth&news=tech` — concurrent fan-out
- `GET /v1/me` — current user info
- `POST /v1/api-keys` — create personal API key

## License

MIT (for learning purposes).
