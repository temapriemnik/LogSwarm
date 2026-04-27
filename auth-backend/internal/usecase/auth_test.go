package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"authbackend/internal/config"
	"authbackend/internal/domain"
)

type mockUserRepoAuth struct {
	mock.Mock
}

func (m *mockUserRepoAuth) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepoAuth) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepoAuth) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepoAuth) UpdatePassword(ctx context.Context, id int64, passwordHash string) error {
	args := m.Called(ctx, id, passwordHash)
	return args.Error(0)
}

func (m *mockUserRepoAuth) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockUserRepoAuth) Activate(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type mockTokenRepo struct {
	mock.Mock
}

func (m *mockTokenRepo) Create(ctx context.Context, token *domain.RefreshToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *mockTokenRepo) GetByToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.RefreshToken), args.Error(1)
}

func (m *mockTokenRepo) Delete(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *mockTokenRepo) DeleteAllForUser(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func TestAuthService_Register(t *testing.T) {
	mockUserRepo := new(mockUserRepoAuth)
	mockTokenRepo := new(mockTokenRepo)
	jwtCfg := config.JWTConfig{
		Secret:         "test-secret",
		AccessExpiry:   15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
	}
	service := NewAuthService(mockUserRepo, mockTokenRepo, jwtCfg)

	mockUserRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	user, err := service.Register(context.Background(), "testuser", "test@example.com", "password123")

	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, "testuser", user.Name)
	require.Empty(t, user.PasswordHash)

	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_Login(t *testing.T) {
	mockUserRepo := new(mockUserRepoAuth)
	mockTokenRepo := new(mockTokenRepo)
	jwtCfg := config.JWTConfig{
		Secret:         "test-secret-key-for-testing-purposes-only",
		AccessExpiry:   15 * time.Minute,
		RefreshExpiry: 7 * 24 * time.Hour,
	}
	service := NewAuthService(mockUserRepo, mockTokenRepo, jwtCfg)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user := &domain.User{
		ID:           1,
		Name:         "testuser",
		Email:        "test@example.com",
		PasswordHash: string(hashedPassword),
	}

	mockUserRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(user, nil)
	mockTokenRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

	tokens, returnedUser, err := service.Login(context.Background(), "test@example.com", "password123")

	require.NoError(t, err)
	require.NotNil(t, tokens)
	require.NotNil(t, tokens.AccessToken)
	require.NotNil(t, tokens.RefreshToken)
	require.NotNil(t, returnedUser)
	require.Empty(t, returnedUser.PasswordHash)

	mockUserRepo.AssertExpectations(t)
	mockTokenRepo.AssertExpectations(t)
}

func generateTestToken(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return hex.EncodeToString(b)
}