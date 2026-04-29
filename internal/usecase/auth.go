package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"authbackend/internal/config"
	"authbackend/internal/domain"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")
)

type AuthService interface {
	Register(ctx context.Context, name, email, password string) (*domain.User, error)
	Login(ctx context.Context, email, password string) (*domain.TokenPair, *domain.User, error)
	Refresh(ctx context.Context, refreshToken string) (*domain.TokenPair, error)
	ValidateToken(ctx context.Context, accessToken string) (*domain.User, error)
}

type authService struct {
	userRepo   UserRepository
	tokenRepo  TokenRepository
	jwtConfig config.JWTConfig
}

type JWTClaims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

func NewAuthService(userRepo UserRepository, tokenRepo TokenRepository, jwtCfg config.JWTConfig) AuthService {
	return &authService{
		userRepo:   userRepo,
		tokenRepo:  tokenRepo,
		jwtConfig:  jwtCfg,
	}
}

func (s *authService) Register(ctx context.Context, name, email, password string) (*domain.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Name:         name,
		Email:        email,
		PasswordHash: string(hash),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	user.PasswordHash = ""
	return user, nil
}

func (s *authService) Login(ctx context.Context, email, password string) (*domain.TokenPair, *domain.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, nil, ErrInvalidPassword
	}

	tokens, err := s.generateTokens(ctx, user.ID)
	if err != nil {
		return nil, nil, err
	}

	user.PasswordHash = ""
	return tokens, user, nil
}

func (s *authService) Refresh(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	token, err := s.tokenRepo.GetByToken(ctx, refreshToken)
	if err != nil {
		return nil, ErrInvalidToken
	}

	if time.Now().After(token.ExpiresAt) {
		s.tokenRepo.Delete(ctx, refreshToken)
		return nil, ErrTokenExpired
	}

	if err := s.tokenRepo.Delete(ctx, refreshToken); err != nil {
		return nil, err
	}

	tokens, err := s.generateTokens(ctx, token.UserID)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func (s *authService) ValidateToken(ctx context.Context, accessToken string) (*domain.User, error) {
	claims := &JWTClaims{}

	token, err := jwt.ParseWithClaims(accessToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtConfig.Secret), nil
	})

	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	if time.Now().After(claims.ExpiresAt.Time) {
		return nil, ErrTokenExpired
	}

	return s.userRepo.GetByID(ctx, claims.UserID)
}

func (s *authService) generateTokens(ctx context.Context, userID int64) (*domain.TokenPair, error) {
	accessToken, err := s.generateAccessToken(userID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateRefreshToken()
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(s.jwtConfig.RefreshExpiry)

	if err := s.tokenRepo.Create(ctx, &domain.RefreshToken{
		UserID:    userID,
		Token:    refreshToken,
		ExpiresAt: expiresAt,
	}); err != nil {
		return nil, err
	}

	return &domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *authService) generateAccessToken(userID int64) (string, error) {
	expiresAt := time.Now().Add(s.jwtConfig.AccessExpiry)

	claims := JWTClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtConfig.Secret))
}

func (s *authService) generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

var _ AuthService = (*authService)(nil)