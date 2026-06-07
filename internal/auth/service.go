package auth

import (
	"context"
	"fmt"

	"github.com/eminel9311/freeapi-hub/internal/domain"
)

// UserRepository là interface — service chỉ cần biết shape, không cần biết Postgres/MySQL/...
type UserRepository interface {
	Create(ctx context.Context, email, passwordHash string) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
}

type Service struct {
	users UserRepository
	jwt   *JWTManager
}

func NewService(users UserRepository, jwt *JWTManager) *Service {
	return &Service{users: users, jwt: jwt}
}

// Register tạo user mới + trả access token.
func (s *Service) Register(ctx context.Context, email, password string) (*domain.User, string, error) {
	// Check email đã tồn tại
	existing, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return nil, "", fmt.Errorf("check existing: %w", err)
	}
	if existing != nil {
		return nil, "", domain.ErrEmailTaken
	}

	// Hash password
	hash, err := HashPassword(password)
	if err != nil {
		return nil, "", err
	}

	// Insert user
	user, err := s.users.Create(ctx, email, hash)
	if err != nil {
		return nil, "", err
	}

	// Generate JWT
	token, err := s.jwt.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

// Login verify credentials + trả access token.
func (s *Service) Login(ctx context.Context, email, password string) (*domain.User, string, error) {
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return nil, "", err
	}

	if user == nil {
		return nil, "", domain.ErrInvalidCredentials
	}

	if err := VerifyPassword(password, user.PasswordHash); err != nil {
		return nil, "", domain.ErrInvalidCredentials
	}

	token, err := s.jwt.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, "", err
	}
	return user, token, nil

}
