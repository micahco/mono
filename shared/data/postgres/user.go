package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/micahco/mono/shared/data"
	"github.com/micahco/mono/shared/data/internal/uuid"
)

type UserRepository struct {
	Pool *pgxpool.Pool
}

func (r *UserRepository) New(ctx context.Context, email string, passwordHash []byte) (*data.User, error) {
	u := data.User{
		Email:        email,
		PasswordHash: passwordHash,
	}

	sql := `
		INSERT INTO user_ (email_, password_hash_)
		VALUES($1, $2)
		RETURNING id_, version_, created_at_;`
	args := []any{
		u.Email,
		u.PasswordHash,
	}
	err := r.Pool.QueryRow(ctx, sql, args...).Scan(
		&u.ID,
		&u.Version,
		&u.CreatedAt,
	)
	if err != nil {
		switch {
		case pgErrCode(err) == pgerrcode.UniqueViolation:
			return nil, data.ErrDuplicateEmail
		default:
			return nil, err
		}
	}

	return &u, nil
}

func (r *UserRepository) Get(ctx context.Context, id uuid.UUID) (*data.User, error) {
	var u data.User

	sql := `
		SELECT id_, version_, created_at_, email_, password_hash_
		FROM user_ WHERE id_ = $1;`
	args := []any{
		id,
	}
	err := r.Pool.QueryRow(ctx, sql, args...).Scan(
		&u.ID,
		&u.Version,
		&u.CreatedAt,
		&u.Email,
		&u.PasswordHash,
	)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, data.ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &u, nil
}

func (r *UserRepository) GetForCredentials(ctx context.Context, email, plaintextPassword string, cmp data.ComparePasswordAndHash) (*data.User, error) {
	var u data.User

	sql := `
		SELECT id_, version_, created_at_, email_, password_hash_
		FROM user_ WHERE email_ = $1;`
	args := []any{
		email,
	}
	err := r.Pool.QueryRow(ctx, sql, args...).Scan(
		&u.ID,
		&u.Version,
		&u.CreatedAt,
		&u.Email,
		&u.PasswordHash,
	)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, data.ErrRecordNotFound
		default:
			return nil, err
		}
	}

	match, err := cmp(plaintextPassword, u.PasswordHash)
	if err != nil {
		return nil, err
	}
	if !match {
		return nil, data.ErrInvalidCredentials
	}

	return &u, nil
}

func (r *UserRepository) GetForVerificationToken(ctx context.Context, scope string, tokenHash []byte) (*data.User, error) {
	var u data.User
	var expiry time.Time

	sql := `
		SELECT user_.id_, user_.version_, user_.created_at_, 
			user_.email_, user_.password_hash_, verification_token_.expiry_
		FROM user_
		INNER JOIN verification_token_
		ON user_.id_ = verification_token_.user_id_
		WHERE verification_token_.scope_ = $1
		AND verification_token_.hash_ = $2;`
	args := []any{
		scope,
		tokenHash,
	}
	err := r.Pool.QueryRow(ctx, sql, args...).Scan(
		&u.ID,
		&u.Version,
		&u.CreatedAt,
		&u.Email,
		&u.PasswordHash,
		&expiry,
	)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, data.ErrRecordNotFound
		default:
			return nil, err
		}
	}

	if time.Now().After(expiry) {
		return nil, data.ErrExpiredToken
	}

	return &u, nil
}

func (r *UserRepository) GetForAuthenticationToken(ctx context.Context, tokenHash []byte) (*data.User, error) {
	var u data.User
	var expiry time.Time

	sql := `
		SELECT user_.id_, user_.version_, user_.created_at_, 
		user_.email_, user_.password_hash_, authentication_token_.expiry_
		FROM user_
		INNER JOIN authentication_token_
		ON user_.id_ = authentication_token_.user_id_
		WHERE authentication_token_.hash_ = $1;`
	args := []any{
		tokenHash,
	}
	err := r.Pool.QueryRow(ctx, sql, args...).Scan(
		&u.ID,
		&u.Version,
		&u.CreatedAt,
		&u.Email,
		&u.PasswordHash,
		&expiry,
	)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, data.ErrRecordNotFound
		default:
			return nil, err
		}
	}

	if time.Now().After(expiry) {
		return nil, data.ErrExpiredToken
	}

	return &u, nil
}

func (r *UserRepository) ExistsWithEmail(ctx context.Context, email string) (bool, error) {
	var exists bool

	sql := `
		SELECT EXISTS (
			SELECT 1
			FROM user_
			WHERE email_ = $1
		);`
	args := []any{
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

func (r *UserRepository) Update(ctx context.Context, u *data.User) error {
	sql := `
		UPDATE user_ 
        SET email_ = $1, password_hash_ = $2, version_ = version_ + 1
        WHERE id_ = $3 AND version_ = $4
        RETURNING version_;`
	args := []any{
		u.Email,
		u.PasswordHash,
		u.ID,
		u.Version,
	}
	err := r.Pool.QueryRow(ctx, sql, args...).Scan(
		&u.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return data.ErrEditConflict
		case pgErrCode(err) == pgerrcode.UniqueViolation:
			return data.ErrDuplicateEmail
		default:
			return err
		}
	}

	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	sql := `
		DELETE FROM user_
		WHERE id_ = $1;`

	res, err := r.Pool.Exec(ctx, sql, id)
	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return data.ErrRecordNotFound
	}

	return nil
}
