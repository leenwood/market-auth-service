package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	RoleBuyer  = "buyer"
	RoleSeller = "seller"
	RoleAdmin  = "admin"
)

type User struct {
	ID           uuid.UUID
	Email        string
	Name         string
	PasswordHash string
	Role         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}
