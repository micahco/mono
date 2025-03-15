package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/micahco/mono/shared/data"
	"github.com/micahco/mono/shared/data/internal/uuid"
)

type AuthenticationTokenRepository struct {
	Pool *pgxpool.Pool
}

func (r *AuthenticationTokenRepository) New(ctx context.Context, tokenHash []byte, ttl time.Duration, userID uuid.UUID) (*data.AuthenticationToken, error) {
	t := &data.AuthenticationToken{
		Hash:   tokenHash,
		Expiry: time.Now().Add(ttl),
		UserID: uuid.Nil,
	}

	sql := `
		INSERT INTO authentication_token_ (hash_, expiry_, user_id_)
		VALUES($1, $2, $3);`
	args := []any{t.Hash, t.Expiry, t.UserID}
	_, err := r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (r *AuthenticationTokenRepository) Purge(ctx context.Context, userID uuid.UUID) error {
	sql := `
		DELETE FROM authentication_token_
		WHERE user_id_ = $1;
		`
	args := []any{userID}
	_, err := r.Pool.Exec(ctx, sql, args...)

	return err
}
