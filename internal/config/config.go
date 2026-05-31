package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config gom toàn bộ cấu hình của app.
// Sử dụng envconfig: mỗi field map với 1 env var.
// Ví dụ: HTTPPort → HTTP_PORT
type Config struct {
	// HTTP server
	HTTPPort         string        `envconfig:"HTTP_PORT" default:"8080"`
	HTTPReadTimeout  time.Duration `envconfig:"HTTP_READ_TIMEOUT" default:"10s"`
	HTTPWriteTimeout time.Duration `envconfig:"HTTP_WRITE_TIMEOUT" default:"15s"`

	// Database
	DatabaseURL string `envconfig:"DATABASE_URL" required:"true"`

	// Redis (optional, fallback in-memory cache nếu không có)
	RedisURL string `envconfig:"REDIS_URL" default:""`

	// JWT
	JWTSecret          string        `envconfig:"JWT_SECRET" required:"true"`
	JWTAccessTokenTTL  time.Duration `envconfig:"JWT_ACCESS_TOKEN_TTL" default:"15m"`
	JWTRefreshTokenTTL time.Duration `envconfig:"JWT_REFRESH_TOKEN_TTL" default:"720h"` // 30 ngày

	// External APIs (free, không cần key cho hầu hết)
	OpenMeteoBaseURL   string `envconfig:"OPEN_METEO_BASE_URL" default:"https://api.open-meteo.com/v1"`
	CoinGeckoBaseURL   string `envconfig:"COIN_GECKO_BASE_URL" default:"https://api.coingecko.com/api/v3"`
	FrankfurterBaseURL string `envconfig:"FRANKFURTER_BASE_URL" default:"https://api.frankfurter.app"`
	NewsAPIKey         string `envconfig:"NEWS_API_KEY" default:""` // optional

	// App behavior
	LogLevel        string        `envconfig:"LOG_LEVEL" default:"info"`
	CacheDefaultTTL time.Duration `envconfig:"CACHE_DEFAULT_TTL" default:"5m"`
	RateLimit       int           `envconfig:"RATE_LIMIT_PER_MIN" default:"60"`
}

// Load đọc env vars và validate.
func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	return &cfg, nil
}
