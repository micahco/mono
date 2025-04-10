package data

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
)

// Default expiry duration
const AuthenticationTokenTTL = time.Hour * 36

type AuthenticationTokenRepository interface {
	New(ctx context.Context, tokenHash []byte, expiry time.Time, userID uuid.UUID) error
	Get(ctx context.Context, tokenHash []byte) (*AuthenticationToken, error)
	Purge(ctx context.Context, userID uuid.UUID) error
}

type AuthenticationToken struct {
	Hash   []byte    `json:"-"`
	Expiry time.Time `json:"expiry"`
	UserID uuid.UUID `json:"-"`
}
