package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"authbackend/generated/db"
	"authbackend/internal/domain"
	"authbackend/internal/usecase"
)

type UserRepository = usecase.UserRepository

type userRepository struct {
	db *db.Queries
}

func NewUserRepository(db *db.Queries) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	params := db.CreateUserParams{
		Name:         user.Name,
		Email:        user.Email,
		PasswordHash: pgtype.Text{String: user.PasswordHash, Valid: true},
	}

	newUser, err := r.db.CreateUser(ctx, params)
	if err != nil {
		return err
	}

	mapToDomain(user, newUser)
	return nil
}

func (r *userRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	user, err := r.db.GetUserByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	return mapFromDB(user), nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := r.db.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, ErrUserNotFound
	}

	return mapFromDB(user), nil
}

func (r *userRepository) UpdatePassword(ctx context.Context, id int64, passwordHash string) error {
	params := db.UpdateUserPasswordParams{
		ID:           id,
		PasswordHash: pgtype.Text{String: passwordHash, Valid: true},
	}

	return r.db.UpdateUserPassword(ctx, params)
}

func (r *userRepository) Delete(ctx context.Context, id int64) error {
	return r.db.DeleteUser(ctx, id)
}

func (r *userRepository) Activate(ctx context.Context, id int64) error {
	return r.db.ActivateUser(ctx, id)
}

func mapFromDB(dbUser db.User) *domain.User {
	return &domain.User{
		ID:           dbUser.ID,
		Name:         dbUser.Name,
		Email:        dbUser.Email,
		PasswordHash: dbUser.PasswordHash.String,
		CreatedAt:    dbUser.CreatedAt.Time,
		IsActive:     dbUser.IsActive,
	}
}

func mapToDomain(domainUser *domain.User, dbUser db.User) {
	domainUser.ID = dbUser.ID
	domainUser.Name = dbUser.Name
	domainUser.Email = dbUser.Email
	domainUser.PasswordHash = dbUser.PasswordHash.String
	domainUser.CreatedAt = dbUser.CreatedAt.Time
	domainUser.IsActive = dbUser.IsActive
}