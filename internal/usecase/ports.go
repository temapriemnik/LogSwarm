package usecase

import (
	"context"

	"authbackend/internal/domain"
)

// UserRepository defines the interface for user data access.
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id int64) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	UpdatePassword(ctx context.Context, id int64, passwordHash string) error
	Delete(ctx context.Context, id int64) error
	Activate(ctx context.Context, id int64) error
}

// TokenRepository defines the interface for token data access.
type TokenRepository interface {
	Create(ctx context.Context, token *domain.RefreshToken) error
	GetByToken(ctx context.Context, token string) (*domain.RefreshToken, error)
	Delete(ctx context.Context, token string) error
	DeleteAllForUser(ctx context.Context, userID int64) error
}