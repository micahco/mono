package data

import (
	"context"
	"time"

	"github.com/micahco/mono/shared/data/internal/uuid"
)

type VerificationTokenRepository interface {
	New(ctx context.Context, tokenHash []byte, ttl time.Duration, scope, email string) (*VerificationToken, error)
	NewWithUserID(ctx context.Context, tokenHash []byte, ttl time.Duration, scope, email string, userID uuid.UUID) (*VerificationToken, error)
	ExistsWithEmail(ctx context.Context, scope, email string) (bool, error)
	ExistsWithEmailAndUserID(ctx context.Context, scope, email string, userID uuid.UUID) (bool, error)
	PurgeWithEmail(ctx context.Context, email string) error
	PurgeWithUserID(ctx context.Context, userID uuid.UUID) error
	Verify(ctx context.Context, tokenHash []byte, scope, email string) error
}

type VerificationToken struct {
	Hash   []byte    `json:"-"`
	Expiry time.Time `json:"expiry"`
	Scope  string    `json:"-"`
	Email  string    `json:"-"`
	UserID uuid.UUID `json:"-"`
}
