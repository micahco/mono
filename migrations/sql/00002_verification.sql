-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS verification_token_ (
    hash_ BYTEA PRIMARY KEY,
    expiry_ TIMESTAMPTZ NOT NULL,
    scope_ TEXT NOT NULL,
    email_ CITEXT NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS verification_token_;
-- +goose StatementEnd
