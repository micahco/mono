package data

import (
	"context"
	"time"
)

const (
	VerificationTokenTTL = time.Hour * 36
	ScopeRegistration    = "registration"
	ScopeAccountDeletion = "account-deletion"
	ScopeEmailChange     = "email-change"
	ScopePasswordReset   = "password-reset"
)

type VerificationTokenRepository interface {
	New(ctx context.Context, tokenHash []byte, expiry time.Time, scope, email string) error
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
