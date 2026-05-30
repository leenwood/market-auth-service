package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/leenwood/market-auth-service/internal/core/domain"
	"github.com/leenwood/market-auth-service/internal/core/port"
	"golang.org/x/crypto/bcrypt"
)

type AuthUseCase struct {
	users      port.UserRepository
	tokens     port.TokenRepository
	tm         port.TokenManager
	refreshTTL time.Duration
}

func NewAuthUseCase(
	users port.UserRepository,
	tokens port.TokenRepository,
	tm port.TokenManager,
	refreshTTL time.Duration,
) *AuthUseCase {
	return &AuthUseCase{
		users:      users,
		tokens:     tokens,
		tm:         tm,
		refreshTTL: refreshTTL,
	}
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

type GuestTokenResponse struct {
	AccessToken string
	GuestUserID string
	ExpiresIn   int64
}

func (uc *AuthUseCase) Register(ctx context.Context, email, password, name string) (*domain.User, error) {
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

	if err := uc.users.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (uc *AuthUseCase) Login(ctx context.Context, email, password string) (*TokenPair, error) {
	user, err := uc.users.FindByEmail(ctx, email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}
	if user.DeletedAt != nil {
		return nil, domain.ErrUserDeleted
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}
	return uc.issuePair(ctx, user)
}

func (uc *AuthUseCase) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	userID, err := uc.tokens.Get(ctx, refreshToken)
	if err != nil {
		return nil, domain.ErrTokenNotFound
	}

	user, err := uc.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.DeletedAt != nil {
		return nil, domain.ErrUserDeleted
	}

	if err := uc.tokens.Delete(ctx, refreshToken); err != nil {
		return nil, fmt.Errorf("revoke old refresh token: %w", err)
	}
	return uc.issuePair(ctx, user)
}

func (uc *AuthUseCase) Logout(ctx context.Context, refreshToken string) error {
	err := uc.tokens.Delete(ctx, refreshToken)
	if err != nil && !errors.Is(err, domain.ErrTokenNotFound) {
		return err
	}
	return nil
}

func (uc *AuthUseCase) GetProfile(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	user, err := uc.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.DeletedAt != nil {
		return nil, domain.ErrUserDeleted
	}
	return user, nil
}

func (uc *AuthUseCase) DeleteAccount(ctx context.Context, userID uuid.UUID) error {
	return uc.users.SoftDelete(ctx, userID)
}

func (uc *AuthUseCase) IssueGuestToken(_ context.Context, existingID *uuid.UUID) (*GuestTokenResponse, error) {
	guestID := uuid.New()
	if existingID != nil {
		guestID = *existingID
	}
	token, err := uc.tm.IssueGuestToken(guestID)
	if err != nil {
		return nil, fmt.Errorf("issue guest token: %w", err)
	}
	return &GuestTokenResponse{
		AccessToken: token,
		GuestUserID: guestID.String(),
		ExpiresIn:   uc.tm.GuestTTLSeconds(),
	}, nil
}

func (uc *AuthUseCase) issuePair(ctx context.Context, user *domain.User) (*TokenPair, error) {
	accessToken, err := uc.tm.IssueAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, fmt.Errorf("issue access token: %w", err)
	}

	refreshToken, err := uc.tm.IssueRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("issue refresh token: %w", err)
	}

	if err := uc.tokens.Save(ctx, refreshToken, user.ID, uc.refreshTTL); err != nil {
		return nil, fmt.Errorf("save refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    uc.tm.AccessTTLSeconds(),
	}, nil
}
