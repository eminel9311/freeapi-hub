package auth

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/eminel9311/freeapi-hub/internal/domain"
	"github.com/eminel9311/freeapi-hub/internal/httputil"
)

// Request bodies
type registerReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Response body
type authResp struct {
	User  userView `json:"user"`
	Token string   `json:"access_token"`
}

type userView struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
}

func (s *Service) RegisterHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req registerReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		user, token, err := s.Register(r.Context(), req.Email, req.Password)
		if err != nil {
			if errors.Is(err, domain.ErrEmailTaken) {
				httputil.Error(w, http.StatusConflict, "email already taken")
				return
			}
			httputil.Error(w, http.StatusInternalServerError, err.Error())
			return
		}

		httputil.JSON(w, http.StatusCreated, authResp{
			User:  userView{ID: user.ID, Email: user.Email},
			Token: token,
		})

	}
}

func (s *Service) LoginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req loginReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httputil.Error(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		user, token, err := s.Login(r.Context(), req.Email, req.Password)
		if err != nil {
			if errors.Is(err, domain.ErrInvalidCredentials) {
				httputil.Error(w, http.StatusUnauthorized, "invalid credentials")
				return
			}
			httputil.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		httputil.JSON(w, http.StatusOK, authResp{
			User:  userView{ID: user.ID, Email: user.Email},
			Token: token,
		})

	}

}
