package data

import (
	"context"
	"time"

	"github.com/micahco/mono/shared/data/internal/crypto"
)

const (
	VerificationTokenTTL = time.Hour * 36
	ScopeRegistration    = "registration"
	ScopeAccountDeletion = "account-deletion"
	ScopeEmailChange     = "email-change"
	ScopePasswordReset   = "password-reset"
)

type VerificationTokenRepository interface {
	New(ctx context.Context, token crypto.Token, scope, email string) error
	Get(ctx context.Context, tokenHash []byte) (*VerificationToken, error)
	Exists(ctx context.Context, scope, email string) (bool, error)
	Purge(ctx context.Context, email string) error
	Verify(ctx context.Context, tokenHash []byte, scope, email string) error
}

type VerificationToken struct {
	Hash   []byte    `json:"-"`
	Expiry time.Time `json:"expiry"`
	Scope  string    `json:"-"`
	Email  string    `json:"-"`
}
