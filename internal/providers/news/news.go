package news

import (
	"context"
	"time"

	"github.com/go-resty/resty/v2"
)

type Article struct {
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	Source      string    `json:"source"`
	PublishedAt time.Time `json:"published_at"`
}

type Response struct {
	Query    string    `json:"query"`
	Articles []Article `json:"articles"`
}

// Provider wrap một news API.
// Lựa chọn:
//   - NewsAPI (newsapi.org) - cần free API key, 100 req/day
//   - Hacker News API - không cần key: https://hn.algolia.com/api
//
// Recommend dùng Hacker News để bắt đầu (đơn giản, không key).
type Provider struct {
	client  *resty.Client
	baseURL string
	apiKey  string
}

func New(baseURL, apiKey string) *Provider {
	return &Provider{
		client:  resty.New().SetTimeout(5 * time.Second),
		baseURL: baseURL,
		apiKey:  apiKey,
	}
}

func (p *Provider) Name() string { return "news" }

// TODO TUẦN 2: implement.
func (p *Provider) Fetch(ctx context.Context, params map[string]string) (any, error) {
	return nil, nil
}
