//go:build integration
// +build integration

package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/ory/dockertest/v4"

	"authbackend/generated/db"
	"authbackend/internal/domain"
)

var testDB *db.Queries
var testDSN string

func setupTest(t *testing.T) *db.Queries {
	pool := dockertest.NewPoolT(t, "")
	dbContainer := pool.RunT(t, "postgres",
		dockertest.WithTag("15"),
		dockertest.WithEnv([]string{
			"POSTGRES_PASSWORD=secret",
			"POSTGRES_USER=postgres",
			"POSTGRES_DB=postgres",
		}),
	)

	hostPort := dbContainer.GetHostPort("5432/tcp")
	testDSN = fmt.Sprintf("postgres://postgres:secret@%s/postgres?sslmode=disable", hostPort)

	err := pool.Retry(t.Context(), 30*time.Second, func() error {
		ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
		defer cancel()

		conn, err := pgx.Connect(ctx, testDSN)
		if err != nil {
			return err
		}
		defer conn.Close(ctx)

		_, err = conn.Exec(ctx, `
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
		return err
	})
	if err != nil {
		t.Fatalf("could not setup database: %v", err)
	}

	conn, err := pgx.Connect(t.Context(), testDSN)
	if err != nil {
		t.Fatalf("could not connect: %v", err)
	}
	_ = conn

	testDB = db.New(conn)

	return testDB
}

func TestUserRepository_Create(t *testing.T) {
	db := setupTest(t)
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
	db := setupTest(t)
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
	db := setupTest(t)
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
	db := setupTest(t)
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
	db := setupTest(t)
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
	db := setupTest(t)
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