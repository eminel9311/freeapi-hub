package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// Cost mặc định cho bcrypt. 10 = ~100ms hash time, đủ chậm để chống brute-force.
const bcryptCost = 12

// HashPassword băm password với bcrypt.
// Bcrypt tự sinh salt và embed vào output → không cần lưu salt riêng.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

// VerifyPassword check password có khớp với hash không.
// Constant-time compare → an toàn chống timing attack.
func VerifyPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
