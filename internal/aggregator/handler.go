package aggregator

import (
	"encoding/json"
	"net/http"

	"github.com/eminel9311/freeapi-hub/internal/httputil"
)

func (s *Service) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}

		results, err := s.FetchAll(r.Context(), req)
		if err != nil {
			httputil.Error(w, http.StatusInternalServerError, err.Error())
			return
		}

		httputil.JSON(w, http.StatusOK, results)

	}

}
