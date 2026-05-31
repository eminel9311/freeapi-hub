package exchange

import (
	"context"
	"time"

	"github.com/go-resty/resty/v2"
)

type Response struct {
	From      string    `json:"from"`
	To        string    `json:"to"`
	Rate      float64   `json:"rate"`
	Amount    float64   `json:"amount"`
	Result    float64   `json:"result"`
	Timestamp time.Time `json:"timestamp"`
}

// Provider wrap Frankfurter API.
// Free, không cần API key. Docs: https://www.frankfurter.app
type Provider struct {
	client  *resty.Client
	baseURL string
}

func New(baseURL string) *Provider {
	return &Provider{
		client:  resty.New().SetTimeout(5 * time.Second),
		baseURL: baseURL,
	}
}

func (p *Provider) Name() string { return "exchange" }

// TODO TUẦN 2: implement.
// Endpoint: GET /latest?amount=100&from=USD&to=VND
func (p *Provider) Fetch(ctx context.Context, params map[string]string) (any, error) {
	return nil, nil
}
