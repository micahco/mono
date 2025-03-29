package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gofrs/uuid/v5"
)

type handlerWithError func(w http.ResponseWriter, r *http.Request) error

// http.HandlerFunc wrapper with error handling
func (app *application) handle(h handlerWithError) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			var formErrors FormErrors
			switch {
			case errors.As(err, &formErrors):
				// Redirect to referer with form errors as session data
				app.putFormErrors(r, formErrors)
				http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
			default:
				app.serverError(w, "handled unexpected error", err)
			}
		}
	}
}

// Writes to response writer with error message and status code. Mimics http.Error()
func (app *application) errorResponse(w http.ResponseWriter, errorMessage string, statusCode int) {
	err := app.renderError(w, errorMessage, statusCode)
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

// Reads session authenticated user id key and checks if that user exists.
// If all systems check, then set authenticated context to the request.
func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := app.sessionManager.Get(r.Context(), authenticatedUserIDSessionKey).(uuid.UUID)
		if !ok {
			// Remove invalid session value
			app.sessionManager.Remove(r.Context(), authenticatedUserIDSessionKey)

			// Continue handling request
			next.ServeHTTP(w, r)

			return
		}

		exists, err := app.db.Users.Exists(r.Context(), id)
		if err != nil {
			app.serverError(w, "middleware authenticate", err)

			return
		}
		if exists {
			ctx := context.WithValue(r.Context(), isAuthenticatedContextKey, true)
			r = r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !app.isAuthenticated(r) {
			http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
			return
		}

		// Prevent pages that require authentication from being cached
		w.Header().Add("Cache-Control", "no-store")

		next.ServeHTTP(w, r)
	})
}

func (app *application) csrfFailureHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.logger.Error("csrf failure handler",
			slog.String("method", r.Method),
			slog.String("uri", r.URL.RequestURI()),
		)

		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	})
}
