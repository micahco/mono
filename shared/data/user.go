package data

import (
	"context"
	"time"

	"github.com/micahco/mono/shared/data/internal/uuid"
)

type ComparePasswordAndHash func(plaintextPassword string, passwordHash []byte) (bool, error)

type UserRepository interface {
	New(ctx context.Context, email string, passwordHash []byte) (*User, error)
	Get(ctx context.Context, id uuid.UUID) (*User, error)
	GetForCredentials(ctx context.Context, email, plaintextPassword string, cmp ComparePasswordAndHash) (*User, error)
	GetForVerificationToken(ctx context.Context, scope string, tokenHash []byte) (*User, error)
	GetForAuthenticationToken(ctx context.Context, tokenHash []byte) (*User, error)
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
