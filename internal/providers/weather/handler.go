package weather

import (
	"net/http"

	"github.com/eminel9311/freeapi-hub/internal/httputil"
)

// Handler trả về HTTP handler cho route /v1/weather.
// Đây là pattern phổ biến: method trên Provider trả về handler,
// để handler "capture" được provider qua closure.
func (p *Provider) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		city := r.URL.Query().Get("city")
		if city == "" {
			httputil.Error(w, http.StatusBadRequest, "missing 'city' query param")
			return
		}

		// r.Context() propagate từ chi router - nếu client disconnect,
		// context cancel → resty request cũng cancel.
		data, err := p.Fetch(r.Context(), map[string]string{"city": city})

		if err != nil {
			// 502 Bad Gateway - lỗi từ upstream API (không phải lỗi của ta)
			httputil.Error(w, http.StatusBadGateway, err.Error())
			return
		}
		httputil.JSON(w, http.StatusOK, data)

	}
}
