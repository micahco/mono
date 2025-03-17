package data

import (
	"context"
	"time"

	"github.com/micahco/mono/lib/data/internal/uuid"
)

type ComparePasswordAndHash func(plaintextPassword string, passwordHash []byte) (bool, error)

type UserRepository interface {
	New(ctx context.Context, email string, passwordHash []byte) (*User, error)
	Get(ctx context.Context, id uuid.UUID) (*User, error)
	GetWithEmail(ctx context.Context, email string) (*User, error)
	GetWithVerificationToken(ctx context.Context, scope string, tokenHash []byte) (*User, error)
	GetWithAuthenticationToken(ctx context.Context, tokenHash []byte) (*User, error)
	ExistsWithEmail(ctx context.Context, email string) (bool, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type User struct {
	ID           uuid.UUID `json:"-"`
	Version      int32     `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	Email        string    `json:"email"`
	PasswordHash []byte    `json:"-"`
}

var AnonymousUser = &User{}

func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}
