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

type AuthenticationTokenRepository struct {
	Pool *pgxpool.Pool
}

func (r *AuthenticationTokenRepository) New(ctx context.Context, tokenHash []byte, expiry time.Time, userID uuid.UUID) error {
	sql := `
		INSERT INTO authentication_token_ (hash_, expiry_, user_id_)
		VALUES($1, $2, $3);`
	args := []any{
		tokenHash,
		expiry,
		userID,
	}
	_, err := r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}

	return nil
}

func (r *AuthenticationTokenRepository) Get(ctx context.Context, tokenHash []byte) (*data.AuthenticationToken, error) {
	var at data.AuthenticationToken

	sql := `
		SELECT hash_, expiry_, user_id_
		FROM authentication_token_ WHERE hash_ = $1;`
	args := []any{
		tokenHash,
	}
	err := r.Pool.QueryRow(ctx, sql, args...).Scan(
		&at.Hash,
		&at.Expiry,
		&at.UserID,
	)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, data.ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &at, nil
}

func (r *AuthenticationTokenRepository) Purge(ctx context.Context, userID uuid.UUID) error {
	sql := `
		DELETE FROM authentication_token_
		WHERE user_id_ = $1;
		`
	args := []any{
		userID,
	}
	_, err := r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}

	return nil
}
