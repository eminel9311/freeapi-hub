# Multi-stage build: image cuối cùng chỉ ~15MB.
# Tuần 6 bạn sẽ hiểu vì sao Go cực kỳ thân thiện với Docker.

# Stage 1: build
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go.mod/go.sum trước để cache layer dependencies
COPY go.mod go.sum* ./
RUN go mod download

# Copy source rồi build
COPY . .

# CGO_ENABLED=0: build static binary, không phụ thuộc libc
# -ldflags strip debug info → binary nhỏ hơn
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /app/bin/server \
    ./cmd/server

# Stage 2: runtime (siêu nhỏ)
FROM alpine:3.20

# ca-certificates cần thiết để gọi HTTPS APIs
RUN apk --no-cache add ca-certificates tzdata && \
    update-ca-certificates

WORKDIR /app

# Tạo user non-root vì security
RUN addgroup -S app && adduser -S app -G app
USER app

COPY --from=builder /app/bin/server /app/server

EXPOSE 8080

ENTRYPOINT ["/app/server"]
