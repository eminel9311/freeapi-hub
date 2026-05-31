package crypto

import (
	"context"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

// Response chuẩn hóa cho giá crypto.
type Response struct {
	Symbol      string    `json:"symbol"`
	Name        string    `json:"name"`
	PriceUSD    float64   `json:"price_usd"`
	Change24h   float64   `json:"change_24h_pct"`
	LastUpdated time.Time `json:"last_updated"`
}

// Provider wrap CoinGecko free API.
// Docs: https://www.coingecko.com/en/api/documentation
// Free tier: ~10-30 calls/min, không cần API key cho public endpoints.
type Provider struct {
	client  *resty.Client
	baseURL string
}

func New(baseURL string) *Provider {
	return &Provider{
		client: resty.New().
			SetTimeout(5 * time.Second).
			SetRetryCount(2),
		baseURL: baseURL,
	}
}

func (p *Provider) Name() string {
	return "crypto"
}

// TODO TUẦN 2: implement.
// Endpoint: GET /simple/price?ids=bitcoin&vs_currencies=usd&include_24hr_change=true
// Hoặc: GET /coins/markets?vs_currency=usd&ids=bitcoin,ethereum
func (p *Provider) Fetch(ctx context.Context, params map[string]string) (any, error) {
	coin := params["coin"]

	if coin == "" {
		return nil, fmt.Errorf("missing 'coin' param")
	}

	var raw map[string]struct {
		USD          float64 `json:"usd"`
		USDChange24h float64 `json:"usd_24h_change"`
		LastUpdated  int64   `json:"last_updated_at"`
	}

	_, err := p.client.R().
		SetContext(ctx).
		SetQueryParam("ids", coin).
		SetQueryParam("vs_currencies", "usd").
		SetQueryParam("include_24hr_change", "true").
		SetQueryParam("include_last_updated_at", "true").
		SetResult(&raw).
		Get(p.baseURL + "/simple/price")

	if err != nil {
		return nil, fmt.Errorf("crypto fetch: %w", err)
	}

	info, ok := raw[coin]

	if !ok {
		return nil, fmt.Errorf("crypto fetch: coin not found")
	}

	return Response{
		Symbol:      coin,
		Name:        coin, // CoinGecko không trả về name ở endpoint này, nên tạm dùng symbol.
		PriceUSD:    info.USD,
		Change24h:   info.USDChange24h,
		LastUpdated: time.Unix(info.LastUpdated, 0),
	}, nil

}
