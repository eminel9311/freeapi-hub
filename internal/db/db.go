package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// New khởi tạo connection pool đến PostgreSQL với DSN đã cho.
// pgxpool quản lý nhiều connection, tái sử dụng - không phải mỗi query 1 connection mới.
func New(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)

	if err != nil {
		return nil, fmt.Errorf("create pgx pool: %w", err)
	}

	// Ping để verify connection hoạt động
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return pool, nil
}
