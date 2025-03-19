module github.com/micahco/mono/web

go 1.24.1

require (
	github.com/alexedwards/scs/pgxstore v0.0.0-20250212122300-421ef1d8611c
	github.com/alexedwards/scs/v2 v2.8.0
	github.com/go-chi/chi/v5 v5.2.1
	github.com/go-playground/form/v4 v4.2.1
	github.com/go-playground/validator/v10 v10.25.0
	github.com/gofrs/uuid/v5 v5.3.1
	github.com/justinas/nosurf v1.1.1
	github.com/lmittmann/tint v1.0.7
	github.com/micahco/mono/lib/crypto v0.0.0
	github.com/micahco/mono/lib/data v0.0.0
	github.com/micahco/mono/lib/mailer v0.0.0
	github.com/micahco/mono/lib/middleware v0.0.0
)

replace github.com/micahco/mono/lib/crypto v0.0.0 => ../lib/crypto

replace github.com/micahco/mono/lib/data v0.0.0 => ../lib/data

replace github.com/micahco/mono/lib/mailer v0.0.0 => ../lib/mailer

replace github.com/micahco/mono/lib/middleware v0.0.0 => ../lib/middleware

require (
	github.com/a-h/templ v0.3.833 // indirect
	github.com/alexedwards/argon2id v1.0.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.8 // indirect
	github.com/go-chi/httprate v0.14.1 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/jackc/pgerrcode v0.0.0-20240316143900-6e2875d9b438 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx-gofrs-uuid v0.0.0-20230224015001-1d428863c2e2 // indirect
	github.com/jackc/pgx/v5 v5.7.2 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	golang.org/x/crypto v0.36.0 // indirect
	golang.org/x/net v0.34.0 // indirect
	golang.org/x/sync v0.12.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df // indirect
)
