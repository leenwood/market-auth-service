package seed

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	internal "github.com/leenwood/market-auth-service/internal"
	"golang.org/x/crypto/bcrypt"
)

type seedUser struct {
	email    string
	name     string
	password string
	role     string
}

var users = []seedUser{
	{email: "admin@example.com", name: "Admin User", password: "Admin1234!", role: "admin"},
	{email: "seller@example.com", name: "Seller User", password: "Seller1234!", role: "seller"},
	{email: "buyer@example.com", name: "Buyer User", password: "Buyer1234!", role: "buyer"},
}

func Run(ctx context.Context) error {
	cfg, err := internal.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	db, err := pgxpool.New(ctx, cfg.DatabaseDSN)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		return fmt.Errorf("ping postgres: %w", err)
	}

	for _, u := range users {
		if err := insertUser(ctx, db, u); err != nil {
			return fmt.Errorf("seed user %s: %w", u.email, err)
		}
	}

	slog.Info("seed completed", "users", len(users))
	return nil
}

func insertUser(ctx context.Context, db *pgxpool.Pool, u seedUser) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	now := time.Now().UTC()
	_, err = db.Exec(ctx, `
		INSERT INTO users (id, email, name, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (email) DO NOTHING`,
		uuid.New(), u.email, u.name, string(hash), u.role, now, now,
	)
	if err != nil {
		return err
	}

	slog.Info("seeded user", "email", u.email, "role", u.role)
	return nil
}
