package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/micahco/mono/lib/crypto"
	"github.com/micahco/mono/lib/data"
)

type handlerWithError func(w http.ResponseWriter, r *http.Request) error

func (app *application) handle(h handlerWithError) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			var validationError validation.Errors
			switch {
			case errors.As(err, &validationError):
				app.errorResponse(w, validationError.Error(), http.StatusUnprocessableEntity)
			default:
				app.serverError(w, "handled unexpected error", err)
			}
		}
	}
}

// Writes to response writer with error message and status code. Mimics http.Error()
func (app *application) errorResponse(w http.ResponseWriter, errorMessage string, statusCode int) {
	err := app.writeJSONError(w, errorMessage, statusCode)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// Logs error and responds with generic internal server error message.
func (app *application) serverError(w http.ResponseWriter, logMsg string, err error) {
	app.logger.Error(
		logMsg,
		slog.Any("err", err),
		slog.String("type", fmt.Sprintf("%T", err)),
	)

	app.errorResponse(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *application) recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")

				app.serverError(w, "recovered from panic", fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *application) authenticate(next http.Handler) http.Handler {
	invalidAuthenticationTokenMessage := "invalid or expired authentication token"
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
			w.Header().Set("WWW-Authenticate", "Bearer")
			app.errorResponse(w, invalidAuthenticationTokenMessage, http.StatusUnauthorized)
			return
		}

		plaintextToken := headerParts[1]
		tokenHash := crypto.TokenHash(plaintextToken)

		user, err := app.db.Users.GetWithAuthenticationToken(r.Context(), tokenHash)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound),
				errors.Is(err, data.ErrExpiredToken):
				w.Header().Set("WWW-Authenticate", "Bearer")
				app.errorResponse(w, invalidAuthenticationTokenMessage, http.StatusUnauthorized)
			default:
				app.serverError(w, "unable to get authentication token", err)
			}
			return
		}

		// Add authenticated user to this request's context
		ctx := app.contextSetUser(r.Context(), user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) requireAuthentication(next http.Handler) http.Handler {
	authenticationRequiredMessage := "invalid or expired authentication token"
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r.Context())

		if user.IsAnonymous() {
			app.errorResponse(w, authenticationRequiredMessage, http.StatusUnauthorized)

			return
		}

		next.ServeHTTP(w, r)
	})
}
