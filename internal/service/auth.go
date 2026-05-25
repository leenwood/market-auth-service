package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/leenwood/market-auth-service/internal/domain"
	"github.com/leenwood/market-auth-service/internal/repository"
	"github.com/leenwood/market-auth-service/internal/token"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	users   repository.UserRepository
	tokens  repository.TokenRepository
	jwt     *token.Manager
	refreshTTL time.Duration
}

func NewAuthService(
	users repository.UserRepository,
	tokens repository.TokenRepository,
	jwt *token.Manager,
	refreshTTL time.Duration,
) *AuthService {
	return &AuthService{
		users:      users,
		tokens:     tokens,
		jwt:        jwt,
		refreshTTL: refreshTTL,
	}
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

func (s *AuthService) Register(ctx context.Context, email, password, name string) (*domain.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	now := time.Now().UTC()
	user := &domain.User{
		ID:           uuid.New(),
		Email:        email,
		Name:         name,
		PasswordHash: string(hash),
		Role:         domain.RoleBuyer,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*TokenPair, error) {
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}
	if user.DeletedAt != nil {
		return nil, domain.ErrUserDeleted
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	return s.issuePair(ctx, user)
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	userID, err := s.tokens.Get(ctx, refreshToken)
	if err != nil {
		return nil, domain.ErrTokenNotFound
	}

	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.DeletedAt != nil {
		return nil, domain.ErrUserDeleted
	}

	if err := s.tokens.Delete(ctx, refreshToken); err != nil {
		return nil, fmt.Errorf("revoke old refresh token: %w", err)
	}

	return s.issuePair(ctx, user)
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	err := s.tokens.Delete(ctx, refreshToken)
	if err != nil && !errors.Is(err, domain.ErrTokenNotFound) {
		return err
	}
	return nil
}

func (s *AuthService) GetProfile(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.DeletedAt != nil {
		return nil, domain.ErrUserDeleted
	}
	return user, nil
}

func (s *AuthService) DeleteAccount(ctx context.Context, userID uuid.UUID) error {
	return s.users.SoftDelete(ctx, userID)
}

func (s *AuthService) issuePair(ctx context.Context, user *domain.User) (*TokenPair, error) {
	accessToken, err := s.jwt.IssueAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, fmt.Errorf("issue access token: %w", err)
	}

	refreshToken, err := s.jwt.IssueRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("issue refresh token: %w", err)
	}

	if err := s.tokens.Save(ctx, refreshToken, user.ID, s.refreshTTL); err != nil {
		return nil, fmt.Errorf("save refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    s.jwt.AccessTTLSeconds(),
	}, nil
}
