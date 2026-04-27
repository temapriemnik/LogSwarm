package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"authbackend/generated/db"
	"authbackend/internal/domain"
)

type TokenRepository interface {
	Create(ctx context.Context, token *domain.RefreshToken) error
	GetByToken(ctx context.Context, token string) (*domain.RefreshToken, error)
	Delete(ctx context.Context, token string) error
	DeleteAllForUser(ctx context.Context, userID int64) error
}

type tokenRepository struct {
	db *db.Queries
}

func NewTokenRepository(db *db.Queries) TokenRepository {
	return &tokenRepository{db: db}
}

func (r *tokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	params := db.CreateRefreshTokenParams{
		UserID:    token.UserID,
		TokenHash: token.Token,
		ExpiresAt: pgtype.Timestamptz{Time: token.ExpiresAt, Valid: true},
	}

	dbToken, err := r.db.CreateRefreshToken(ctx, params)
	if err != nil {
		return err
	}

	token.ID = dbToken.ID
	token.UserID = dbToken.UserID
	token.Token = dbToken.TokenHash
	token.ExpiresAt = dbToken.ExpiresAt.Time
	token.CreatedAt = dbToken.CreatedAt.Time

	return nil
}

func (r *tokenRepository) GetByToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	dbToken, err := r.db.GetRefreshTokenByToken(ctx, token)
	if err != nil {
		return nil, err
	}

	return &domain.RefreshToken{
		ID:        dbToken.ID,
		UserID:    dbToken.UserID,
		Token:     dbToken.TokenHash,
		ExpiresAt: dbToken.ExpiresAt.Time,
		CreatedAt: dbToken.CreatedAt.Time,
	}, nil
}

func (r *tokenRepository) Delete(ctx context.Context, token string) error {
	return r.db.DeleteRefreshToken(ctx, token)
}

func (r *tokenRepository) DeleteAllForUser(ctx context.Context, userID int64) error {
	return r.db.DeleteAllUserRefreshTokens(ctx, userID)
}

var _ TokenRepository = (*tokenRepository)(nil)