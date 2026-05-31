package domain

import "time"

// User là entity chính của domain.
// Lưu ý: không có JSON tags ở đây — domain types không nên biết về HTTP.
// JSON serialization là việc của HTTP layer.
type User struct {
	ID           int64
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
