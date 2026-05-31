package weather

import (
	"context"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

// Response là format chuẩn hóa mà ta trả về cho user.
// API gốc (Open-Meteo) có nhiều field, ta chỉ lấy cái cần.
type Response struct {
	City        string    `json:"city"`
	Temperature float64   `json:"temperature_c"`
	WindSpeed   float64   `json:"wind_kmh"`
	Time        time.Time `json:"time"`
}

// Provider wrap Open-Meteo API.
// Open-Meteo KHÔNG cần API key - hoàn hảo cho người mới học.
// Docs: https://open-meteo.com/en/docs
type Provider struct {
	client        *resty.Client
	geocodingURL  string
	forecastURL   string
}

// New khởi tạo Provider.
// Pattern: constructor function trả về *struct (không phải interface).
// Caller nhận về và pass vào nơi cần Provider interface — implicit satisfaction.
func New(geocodingURL, forecastURL string) *Provider {
	return &Provider{
		client: resty.New().
			SetTimeout(5 * time.Second).
			SetRetryCount(2).
			SetRetryWaitTime(500 * time.Millisecond),
		geocodingURL: geocodingURL,
		forecastURL:  forecastURL,
	}
}

func (p *Provider) Name() string {
	return "weather"
}

// Fetch lấy thời tiết theo city.
// TODO TUẦN 1 - BUỔI 4-5: implement function này.
//
// Hint:
//  1. Lấy "city" từ params, validate không rỗng.
//  2. Open-Meteo cần lat/lon, nên bạn cần geocoding trước.
//     Có thể hardcode 1-2 thành phố để bắt đầu, sau đó dùng:
//     https://geocoding-api.open-meteo.com/v1/search?name=Hanoi
//  3. Gọi endpoint forecast: /forecast?latitude=21.03&longitude=105.85&current=temperature_2m,wind_speed_10m
//  4. Parse JSON, map về Response struct.
//  5. Return Response, hoặc wrap error với fmt.Errorf("weather fetch: %w", err)
func (p *Provider) Fetch(ctx context.Context, params map[string]string) (any, error) {
	// city := params["city"]
	// TODO: implement

	city := params["city"]
	if city == "" {
		return nil, fmt.Errorf("Missing 'city' param")
	}

	// === Bước 1: Geocoding === Đổi tên thành phố thành lat/long  ===

	// Struct anonymous để parse response. Chỉ pick field cần thiết.
	var geoResp struct {
		Results []struct {
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
			Name      string  `json:"name"`
		} `json:"results"`
	}

	_, err := p.client.R().
		SetContext(ctx).
		SetQueryParam("name", city).
		SetQueryParam("count", "1"). // Chỉ lấy kết quả đầu tiên
		SetResult(&geoResp).
		Get(p.geocodingURL + "/v1/search")

	if err != nil {
		// %w để wrap error - sau này errors.Is() check được
		return nil, fmt.Errorf("geocoding API: %w", err)
	}

	if len(geoResp.Results) == 0 {
		return nil, fmt.Errorf("city not found: %s", city)
	}

	loc := geoResp.Results[0]

	// === Bước 2: Forecast - dùng lat/lon lấy thời tiết ===

	var fcResp struct {
		Current struct {
			Time        string  `json:"time"`
			Temperature float64 `json:"temperature_2m"`
			WindSpeed   float64 `json:"wind_speed_10m"`
		} `json:"current"`
	}

	_, err = p.client.R().
		SetContext(ctx).
		SetQueryParam("latitude", fmt.Sprintf("%f", loc.Latitude)).
		SetQueryParam("longitude", fmt.Sprintf("%f", loc.Longitude)).
		SetQueryParam("current", "temperature_2m,wind_speed_10m").
		SetResult(&fcResp).
		Get(p.forecastURL + "/v1/forecast")

	if err != nil {
		return nil, fmt.Errorf("forecast API: %w", err)
	}

	// Open-Meteo trả time format "2006-01-02T15:04" (RFC layout đặc biệt của Go)
	t, _ := time.Parse("2006-01-02T15:04", fcResp.Current.Time)

	return Response{
		City:        loc.Name,
		Temperature: fcResp.Current.Temperature,
		WindSpeed:   fcResp.Current.WindSpeed,
		Time:        t,
	}, nil

}
