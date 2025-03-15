package data

import (
	"context"
	"time"

	"github.com/micahco/mono/shared/data/internal/uuid"
)

type AuthenticationTokenRepository interface {
	New(ctx context.Context, tokenHash []byte, ttl time.Duration, userID uuid.UUID) (*AuthenticationToken, error)
	Purge(ctx context.Context, userID uuid.UUID) error
}

type AuthenticationToken struct {
	Hash   []byte    `json:"-"`
	Expiry time.Time `json:"expiry"`
	UserID uuid.UUID `json:"-"`
}
