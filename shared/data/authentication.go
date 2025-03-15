package data

import (
	"context"
	"time"

	"github.com/micahco/mono/shared/data/internal/crypto"
	"github.com/micahco/mono/shared/data/internal/uuid"
)

// Default expiry duration
const AuthenticationTokenTTL = time.Hour * 36

type AuthenticationTokenRepository interface {
	New(ctx context.Context, token crypto.Token, userID uuid.UUID) error
	Get(ctx context.Context, tokenHash []byte) (*AuthenticationToken, error)
	Purge(ctx context.Context, userID uuid.UUID) error
}

type AuthenticationToken struct {
	Hash   []byte    `json:"-"`
	Expiry time.Time `json:"expiry"`
	UserID uuid.UUID `json:"-"`
}
