.PHONY: help run dev build test test-race lint fmt vet tidy \
        docker-up docker-down docker-logs \
        migrate-up migrate-down migrate-create \
        install-tools clean

# Tự generate help từ comments
help: ## Hiển thị danh sách lệnh
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ========== Development ==========

run: ## Chạy server (build rồi run)
	go run ./cmd/server

dev: ## Chạy với air (auto-reload khi save code)
	air

build: ## Build binary
	@mkdir -p bin
	go build -o bin/server ./cmd/server
	@echo "Built: bin/server"

clean: ## Xóa binaries & caches
	rm -rf bin/ tmp/ coverage.out coverage.html

# ========== Testing & Quality ==========

test: ## Chạy unit tests
	go test ./... -v

test-race: ## Chạy tests với race detector (BẮT BUỘC trước khi commit code dùng goroutine)
	go test -race ./...

test-coverage: ## Coverage report HTML
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Open coverage.html"

lint: ## Chạy golangci-lint
	golangci-lint run ./...

fmt: ## Format code
	go fmt ./...

vet: ## Run go vet
	go vet ./...

tidy: ## Cleanup go.mod
	go mod tidy

check: fmt vet lint test-race ## Run mọi check (chạy trước khi commit)

# ========== Docker ==========

docker-up: ## Start Postgres + Redis containers
	docker compose up -d

docker-down: ## Stop containers
	docker compose down

docker-logs: ## Tail container logs
	docker compose logs -f

docker-reset: ## Reset DB (XÓA HẾT DATA)
	docker compose down -v
	docker compose up -d

# ========== Migrations ==========
# Cần install: brew install golang-migrate (hoặc: go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest)

DB_URL ?= postgres://freeapi:freeapi@localhost:5432/freeapi?sslmode=disable

migrate-up: ## Apply tất cả migrations
	migrate -path ./migrations -database "$(DB_URL)" up

migrate-down: ## Rollback 1 migration
	migrate -path ./migrations -database "$(DB_URL)" down 1

migrate-create: ## Tạo migration mới: make migrate-create name=add_xxx
	migrate create -ext sql -dir migrations -seq $(name)

migrate-version: ## Xem version hiện tại
	migrate -path ./migrations -database "$(DB_URL)" version

migrate-force: ## Force version (dùng khi migration bị dirty): make migrate-force v=1
	migrate -path ./migrations -database "$(DB_URL)" force $(v)

# ========== Tools ==========

install-tools: ## Cài đặt tools dev (air, migrate, golangci-lint, mockery)
	@echo "Installing development tools..."
	go install github.com/air-verse/air@latest
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/vektra/mockery/v2@latest
	@echo "Done. Đảm bảo $$GOPATH/bin (hoặc $$HOME/go/bin) có trong PATH."
