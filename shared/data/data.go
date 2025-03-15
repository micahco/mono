package data

import "errors"

// Sentinel errors
var (
	ErrRecordNotFound     = errors.New("data: no matching record found")
	ErrInvalidCredentials = errors.New("data: invalid credentials")
	ErrDuplicateEmail     = errors.New("data: duplicate email")
	ErrExpiredToken       = errors.New("data: expired token")
	ErrEditConflict       = errors.New("data: edit conflict")
)

type DB struct {
	Users                UserRepository
	VerificationTokens   VerificationTokenRepository
	AuthenticationTokens AuthenticationTokenRepository
}
