package domain

import "errors"

// Sentinel errors - dùng errors.Is() để check ở các layer trên.
// Pattern: domain layer định nghĩa lỗi, các layer khác wrap thêm context.
var (
	ErrNotFound          = errors.New("resource not found")
	ErrAlreadyExists     = errors.New("resource already exists")
	ErrInvalidInput      = errors.New("invalid input")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden")
	ErrInvalidCredential = errors.New("invalid credential")
	ErrTokenExpired      = errors.New("token expired")
	ErrRateLimited       = errors.New("rate limited")
	ErrUpstreamFailure   = errors.New("upstream API failed")
	ErrTimeout           = errors.New("operation timed out")
)
