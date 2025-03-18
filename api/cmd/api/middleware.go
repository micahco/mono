package main

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/micahco/mono/lib/crypto"
	"github.com/micahco/mono/lib/data"
)

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add the "Vary: Authorization" header to the response. This indicates to any
		// caches that the response may vary based on the value of the Authorization
		// header in the request.
		w.Header().Add("Vary", "Authorization")

		authorizationHeader := r.Header.Get("Authorization")
		if authorizationHeader == "" {
			// Create new context with request context setting user anon
			ctx := app.contextSetUser(r.Context(), data.AnonymousUser)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w)
			return
		}

		plaintextToken := headerParts[1]
		tokenHash := crypto.TokenHash(plaintextToken)

		// Use same default timeout for accessing db in seperate context
		dbCtx, cancel := context.WithTimeout(r.Context(), defaultTimeout)
		defer cancel()

		user, err := app.db.Users.GetWithAuthenticationToken(dbCtx, tokenHash)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound),
				errors.Is(err, data.ErrExpiredToken):
				app.invalidAuthenticationTokenResponse(w)
			default:
				app.serverErrorResponse(w, "middleware: authenticate: GetForAuthenticationToken", err)
			}
			return
		}

		// Add authenticated user to this request's context
		ctx := app.contextSetUser(r.Context(), user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r.Context())

		if user.IsAnonymous() {
			app.errorResponse(w, AuthenticationRequiredMessage, http.StatusUnauthorized)

			return
		}

		next.ServeHTTP(w, r)
	})
}
