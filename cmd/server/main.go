package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpserver "github.com/eminel9311/freeapi-hub/internal/http"
	"github.com/eminel9311/freeapi-hub/internal/providers/crypto"
	"github.com/eminel9311/freeapi-hub/internal/providers/weather"
	"github.com/joho/godotenv"
)

// TODO Tuần 1: học từng bước thay vì copy hết một lần.
//
// Buổi 1-2: chỉ cần in "Hello Go" và chạy server đơn giản.
// Buổi 3-4: tách config, gọi API ngoài.
// Buổi 5-7: thêm chi router, handlers, response helpers.
//
// File này là "đích đến" của tuần 1. Đừng vội copy hết — hãy tự gõ lại.

func main() {
	// Load .env file (chỉ khi dev local)
	_ = godotenv.Load()

	// Setup structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// TODO: load config từ env vars
	// TODO: khởi tạo providers (weather, crypto, ...)
	// TODO: khởi tạo cache (in-memory hoặc Redis)
	// TODO: khởi tạo database pool
	// TODO: khởi tạo router với chi

	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}

	// mux := http.NewServeMux()
	// mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
	// 	w.WriteHeader(http.StatusOK)
	// 	_, _ = fmt.Fprintln(w, `{"status":"ok"}`)
	// })

	// Khởi tạo provider
	// Open-Meteo có 2 endpoint: geocoding và forecast, nên provider cần 2 URL.
	weatherProv := weather.New(
		"https://geocoding-api.open-meteo.com",
		"https://api.open-meteo.com",
	)
	// Crypto provider cũng khởi tạo tương tự, nhưng chỉ cần 1 base URL./
	cryptoProv := crypto.New(
		"https://api.coingecko.com/api/v3",
	)

	router := httpserver.NewRouter(weatherProv, cryptoProv)
	// Khởi tạo router với provider

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Graceful shutdown setup
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start server trong goroutine để main có thể chờ shutdown signal
	go func() {
		slog.Info("server starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Đợi tín hiệu shutdown
	<-ctx.Done()
	slog.Info("shutdown signal received, draining connections...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped cleanly")
}
