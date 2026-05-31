package crypto

import (
	"net/http"

	"github.com/eminel9311/freeapi-hub/internal/httputil"
)

// Handler trả về HTTP handler cho route /api/v3.
// Đây là pattern phổ biến: method trên Provider trả về handler,
// để handler "capture" được provider qua closure.
func (p *Provider) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		coin := r.URL.Query().Get("coin")
		if coin == "" {
			httputil.Error(w, http.StatusBadRequest, "missing 'coin' query param")
			return
		}

		// r.Context() propagate từ chi router - nếu client disconnect,
		// context cancel → resty request cũng cancel.
		data, err := p.Fetch(r.Context(), map[string]string{"coin": coin})

		if err != nil {
			// 502 Bad Gateway - lỗi từ upstream API (không phải lỗi của ta)
			httputil.Error(w, http.StatusBadGateway, err.Error())
			return
		}
		httputil.JSON(w, http.StatusOK, data)

	}
}
