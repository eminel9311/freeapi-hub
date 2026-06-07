package aggregator

import (
	"context"
	"sync"

	"github.com/eminel9311/freeapi-hub/internal/providers"
	"golang.org/x/sync/errgroup"
)

// Dashboard gom kết quả từ nhiều providers chạy song song.
//
// TUẦN 3 - phần thú vị nhất của Go: concurrency.
//
// Đây là pattern "fan-out / fan-in":
//  1. Nhận 1 request từ user.
//  2. Gọi N API ngoài SONG SONG (mỗi cái 1 goroutine).
//  3. Đợi tất cả xong (hoặc timeout).
//  4. Gom kết quả, trả về cho user.
//
// Lý do dùng errgroup thay vì WaitGroup tự code:
//   - WaitGroup: bạn phải tự handle error trong từng goroutine.
//   - errgroup: nếu 1 goroutine fail, context tự cancel các goroutine còn lại.
//     Bạn cũng có WithContext() để propagate cancel signal.
type Service struct {
	providers map[string]providers.Provider
}

func New(provs ...providers.Provider) *Service {
	m := make(map[string]providers.Provider, len(provs))
	for _, p := range provs {
		m[p.Name()] = p
	}
	return &Service{providers: m}
}

// Request mô tả người dùng muốn gọi providers nào với params gì.
type Request struct {
	Providers map[string]map[string]string // providerName -> params
}

// Result là 1 phần trong response trả về.
type Result struct {
	Provider string `json:"provider"`
	Data     any    `json:"data,omitempty"`
	Error    string `json:"error,omitempty"`
}

// FetchAll gọi tất cả providers song song.
//
// TODO TUẦN 3 - BUỔI 5: implement function này.
//
// Hint:
//
//	import "golang.org/x/sync/errgroup"
//
//	g, gctx := errgroup.WithContext(ctx)
//	g.SetLimit(10) // giới hạn số goroutine chạy đồng thời
//
//	results := make([]Result, 0, len(req.Providers))
//	var mu sync.Mutex
//
//	for name, params := range req.Providers {
//	    name, params := name, params // ⚠️ Go <1.22 cần capture biến trong vòng lặp
//	    g.Go(func() error {
//	        provider, ok := s.providers[name]
//	        if !ok { return nil } // skip unknown
//	        data, err := provider.Fetch(gctx, params)
//	        mu.Lock()
//	        defer mu.Unlock()
//	        if err != nil {
//	            results = append(results, Result{Provider: name, Error: err.Error()})
//	            return nil // không return err → không cancel cả group
//	        }
//	        results = append(results, Result{Provider: name, Data: data})
//	        return nil
//	    })
//	}
//	if err := g.Wait(); err != nil {
//	    return nil, err
//	}
//	return results, nil

func (s *Service) FetchAll(ctx context.Context, req Request) ([]Result, error) {
	// TODO: implement
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(10) // giới hạn số goroutine chạy đồng thời

	results := make([]Result, 0, len(req.Providers))
	var mu sync.Mutex

	for name, params := range req.Providers {
		name, params := name, params // ⚠️ Go <1.22 cần capture biến trong vòng lặp
		g.Go(func() error {
			provider, ok := s.providers[name]
			if !ok {
				return nil
			}
			data, err := provider.Fetch(gctx, params)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				results = append(results, Result{Provider: name, Error: err.Error()})
				return nil
			}
			results = append(results, Result{Provider: name, Data: data})
			return nil
		})

	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return results, nil
}
