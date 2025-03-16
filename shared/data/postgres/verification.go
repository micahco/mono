package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/micahco/mono/shared/data"
)

type VerificationTokenRepository struct {
	Pool *pgxpool.Pool
}

// Create a verification token for an email without an associated User
func (r *VerificationTokenRepository) New(ctx context.Context, tokenHash []byte, expiry time.Time, scope, email string) error {
	vt := &data.VerificationToken{
		Hash:   tokenHash,
		Expiry: expiry,
		Scope:  scope,
		Email:  email,
	}

	sql := `
		INSERT INTO verification_token_ (hash_, expiry_, scope_, email_)
		VALUES($1, $2, $3, $4);`
	args := []any{
		vt.Hash,
		vt.Expiry,
		vt.Scope,
		vt.Email,
	}
	_, err := r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}

	return nil
}

func (r *VerificationTokenRepository) Get(ctx context.Context, tokenHash []byte) (*data.VerificationToken, error) {
	var vt data.VerificationToken

	sql := `
		SELECT hash_, expiry_, scope_, email_
		FROM verification_token_ WHERE hash_ = $1;`
	args := []any{
		tokenHash,
	}
	err := r.Pool.QueryRow(ctx, sql, args...).Scan(
		&vt.Hash,
		&vt.Expiry,
		&vt.Scope,
		&vt.Email,
	)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, data.ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &vt, nil
}

func (r *VerificationTokenRepository) Exists(ctx context.Context, scope, email string) (bool, error) {
	var exists bool

	sql := `
		SELECT EXISTS (
			SELECT 1
			FROM verification_token_
			WHERE scope_ = $1
			AND email_ = $2
		);`
	args := []any{
		scope,
		email,
	}
	err := r.Pool.QueryRow(ctx, sql, args...).Scan(&exists)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return false, nil
		default:
			return exists, err
		}
	}

	return exists, nil
}

func (r *VerificationTokenRepository) Purge(ctx context.Context, email string) error {
	sql := `
		DELETE FROM verification_token_
		WHERE email_ = $1;`
	args := []any{
		email,
	}
	_, err := r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}

	return nil
}

func (r *VerificationTokenRepository) Verify(ctx context.Context, tokenHash []byte, scope, email string) error {
	var expiry time.Time

	sql := `
		SELECT expiry_
		FROM verification_token_
		WHERE hash_ = $1
		AND scope_ = $2
		AND email_ = $3;`
	args := []any{
		tokenHash,
		scope,
		email,
	}
	err := r.Pool.QueryRow(ctx, sql, args...).Scan(&expiry)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return data.ErrRecordNotFound
		default:
			return err
		}
	}

	if time.Now().After(expiry) {
		return data.ErrExpiredToken
	}

	return nil
}
