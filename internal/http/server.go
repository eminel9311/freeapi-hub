package httpserver

import (
	"net/http"
	"time"

	"github.com/eminel9311/freeapi-hub/internal/httputil"
	"github.com/eminel9311/freeapi-hub/internal/providers/crypto"
	"github.com/eminel9311/freeapi-hub/internal/providers/weather"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// NewRouter trả về router chính với middleware đã setup.
// TUẦN 1 - BUỔI 5: bạn sẽ extend file này.
//
// Đây là chỗ "lắp ráp" toàn bộ routes. Mỗi giai đoạn bạn thêm dần:
//   - Tuần 1: GET /v1/weather
//   - Tuần 2: + crypto, exchange, news
//   - Tuần 3: + /v1/dashboard
//   - Tuần 4: + /auth/register, /auth/login + protect /v1/*
//   - Tuần 5: + rate limit middleware
func NewRouter(weatherProv *weather.Provider, cryptoProv *crypto.Provider) *chi.Mux {
	r := chi.NewRouter()

	// Built-in chi middleware. Đọc docs để biết mỗi cái làm gì.
	r.Use(chimiddleware.RequestID)                 // gắn X-Request-ID cho mỗi request
	r.Use(chimiddleware.RealIP)                    // lấy IP thật từ X-Forwarded-For
	r.Use(chimiddleware.Logger)                    // log mỗi request (Tuần 2 ta sẽ thay = slog)
	r.Use(chimiddleware.Recoverer)                 // catch panic, không crash server
	r.Use(chimiddleware.Timeout(30 * time.Second)) // timeout chung
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Authorization", "Content-Type"},
		MaxAge:         300,
	}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		httputil.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// TODO TUẦN 1: thêm route
	// r.Route("/v1", func(r chi.Router) {
	//     r.Get("/weather", weatherHandler)
	// })

	r.Route("/v1", func(r chi.Router) {
		r.Get("/weather", weatherProv.Handler())
		r.Get("/crypto", cryptoProv.Handler())
	})

	return r
}
