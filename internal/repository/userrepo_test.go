//go:build integration
// +build integration

package repository

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"authbackend/generated/db"
	"authbackend/internal/domain"
)

func setupTestDB(t *testing.T) *db.Queries {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("secret"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Fatalf("failed to start container: %v", err)
	}
	t.Cleanup(func() {
		pgContainer.Terminate(ctx)
	})

	dsn, err := pgContainer.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("failed to get dsn: %v", err)
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	t.Cleanup(func() {
		pool.Close()
	})

	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id BIGSERIAL PRIMARY KEY,
			name text NOT NULL UNIQUE,
			email text NOT NULL,
			password_hash text,
			created_at timestamptz NOT NULL DEFAULT now(),
			is_active boolean NOT NULL DEFAULT false
		);
		CREATE INDEX IF NOT EXISTS users_email_lower_idx ON users (lower(email));
		CREATE INDEX IF NOT EXISTS users_is_active_idx ON users (is_active);
	`)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db.New(pool)
}

func TestUserRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	user := &domain.User{
		Name:         "testuser",
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
	}

	err := repo.Create(context.Background(), user)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if user.ID == 0 {
		t.Error("expected user ID to be set")
	}
}

func TestUserRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	createUser := &domain.User{
		Name:         "getbyid",
		Email:        "getbyid@example.com",
		PasswordHash: "hash",
	}
	if err := repo.Create(context.Background(), createUser); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	found, err := repo.GetByID(context.Background(), createUser.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if found.Name != createUser.Name {
		t.Errorf("expected name %s, got %s", createUser.Name, found.Name)
	}
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	createUser := &domain.User{
		Name:         "getbyemail",
		Email:        "getbyemail@example.com",
		PasswordHash: "hash",
	}
	if err := repo.Create(context.Background(), createUser); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	found, err := repo.GetByEmail(context.Background(), createUser.Email)
	if err != nil {
		t.Fatalf("GetByEmail failed: %v", err)
	}

	if found.Name != createUser.Name {
		t.Errorf("expected name %s, got %s", createUser.Name, found.Name)
	}
}

func TestUserRepository_UpdatePassword(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	createUser := &domain.User{
		Name:         "updatepass",
		Email:        "updatepass@example.com",
		PasswordHash: "oldhash",
	}
	if err := repo.Create(context.Background(), createUser); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	newHash := "newhash"
	err := repo.UpdatePassword(context.Background(), createUser.ID, newHash)
	if err != nil {
		t.Fatalf("UpdatePassword failed: %v", err)
	}

	updated, err := repo.GetByID(context.Background(), createUser.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if updated.PasswordHash != newHash {
		t.Errorf("expected password %s, got %s", newHash, updated.PasswordHash)
	}
}

func TestUserRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	createUser := &domain.User{
		Name:         "deleteuser",
		Email:        "deleteuser@example.com",
		PasswordHash: "hash",
	}
	if err := repo.Create(context.Background(), createUser); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	err := repo.Delete(context.Background(), createUser.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = repo.GetByID(context.Background(), createUser.ID)
	if err != ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUserRepository_Activate(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	createUser := &domain.User{
		Name:         "activateuser",
		Email:        "activateuser@example.com",
		PasswordHash: "hash",
		IsActive:     false,
	}
	if err := repo.Create(context.Background(), createUser); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	err := repo.Activate(context.Background(), createUser.ID)
	if err != nil {
		t.Fatalf("Activate failed: %v", err)
	}

	activated, err := repo.GetByID(context.Background(), createUser.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if !activated.IsActive {
		t.Error("expected user to be active")
	}
}