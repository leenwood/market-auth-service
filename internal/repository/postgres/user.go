package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leenwood/market-auth-service/internal/domain"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO users (id, email, name, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		user.ID, user.Email, user.Name, user.PasswordHash, user.Role, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil && isUniqueViolation(err) {
		return domain.ErrEmailTaken
	}
	return err
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	return r.scanUser(ctx, `
		SELECT id, email, name, password_hash, role, created_at, updated_at, deleted_at
		FROM users WHERE email = $1`, email)
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return r.scanUser(ctx, `
		SELECT id, email, name, password_hash, role, created_at, updated_at, deleted_at
		FROM users WHERE id = $1`, id)
}

func (r *UserRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	tag, err := r.db.Exec(ctx,
		`UPDATE users SET deleted_at = $1, updated_at = $1 WHERE id = $2 AND deleted_at IS NULL`,
		now, id,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *UserRepository) scanUser(ctx context.Context, query string, args ...any) (*domain.User, error) {
	u := &domain.User{}
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&u.ID, &u.Email, &u.Name, &u.PasswordHash, &u.Role,
		&u.CreatedAt, &u.UpdatedAt, &u.DeletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	return u, err
}

func isUniqueViolation(err error) bool {
	return err != nil && containsCode(err.Error(), "23505")
}

func containsCode(msg, code string) bool {
	for i := 0; i+len(code) <= len(msg); i++ {
		if msg[i:i+len(code)] == code {
			return true
		}
	}
	return false
}
