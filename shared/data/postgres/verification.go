package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/micahco/mono/shared/data"
	"github.com/micahco/mono/shared/data/internal/uuid"
)

type VerificationTokenRepository struct {
	Pool *pgxpool.Pool
}

func (r *VerificationTokenRepository) New(ctx context.Context, tokenHash []byte, ttl time.Duration, scope, email string) (*data.VerificationToken, error) {
	t := &data.VerificationToken{
		Hash:   tokenHash,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
		Email:  email,
		UserID: uuid.Nil, // flag
	}

	sql := `
		INSERT INTO verification_token_ (hash_, expiry_, scope_, email_, user_id_)
		VALUES($1, $2, $3, $4);`
	args := []any{t.Hash, t.Expiry, t.Scope, t.Email}
	_, err := r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (r *VerificationTokenRepository) NewWithUserID(ctx context.Context, tokenHash []byte, ttl time.Duration, scope, email string, userID uuid.UUID) (*data.VerificationToken, error) {
	t := &data.VerificationToken{
		Hash:   tokenHash,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
		Email:  email,
		UserID: userID,
	}

	sql := `
		INSERT INTO verification_token_ (hash_, expiry_, scope_, email_, user_id_)
		VALUES($1, $2, $3, $4, $5);`
	args := []any{t.Hash, t.Expiry, t.Scope, t.Email, t.UserID}
	_, err := r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (r *VerificationTokenRepository) ExistsWithEmail(ctx context.Context, scope, email string) (bool, error) {
	var exists bool

	sql := `
		SELECT EXISTS (
			SELECT 1
			FROM verification_token_
			WHERE scope_ = $1
			AND email_ = $2
		);`
	args := []any{scope, email}
	err := r.Pool.QueryRow(ctx, sql, args...).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *VerificationTokenRepository) ExistsWithEmailAndUserID(ctx context.Context, scope, email string, userID uuid.UUID) (bool, error) {
	var exists bool

	sql := `
		SELECT EXISTS (
			SELECT 1
			FROM verification_token_
			WHERE scope_ = $1
			AND email_ = $2
			AND user_id_ = $3
		);`
	args := []any{scope, email, userID}
	err := r.Pool.QueryRow(ctx, sql, args...).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *VerificationTokenRepository) PurgeWithEmail(ctx context.Context, email string) error {
	sql := `
		DELETE FROM verification_token_
		WHERE email_ = $1;`
	args := []any{email}
	_, err := r.Pool.Exec(ctx, sql, args...)

	return err
}

func (r *VerificationTokenRepository) PurgeWithUserID(ctx context.Context, userID uuid.UUID) error {
	sql := `
		DELETE FROM verification_token_
		WHERE user_id_ = $1;`
	args := []any{userID}
	_, err := r.Pool.Exec(ctx, sql, args...)

	return err
}

func (r *VerificationTokenRepository) Verify(ctx context.Context, tokenHash []byte, scope, email string) error {
	var expiry time.Time

	sql := `
		SELECT expiry_
		FROM verification_token_
		WHERE hash_ = $1
		AND scope_ = $2
		AND email_ = $3;`
	args := []any{tokenHash, scope, email}
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
