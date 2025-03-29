package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gofrs/uuid/v5"
	"github.com/justinas/nosurf"
	"github.com/micahco/mono/lib/crypto"
	"github.com/micahco/mono/lib/data"
	"github.com/micahco/mono/lib/mailer/emails"
	"github.com/micahco/mono/web/pages"
)

type contextKey string

const (
	authenticatedUserIDSessionKey = "authenticatedUserID"
	verificationEmailSessionKey   = "verificationEmail"
	verificationTokenSessionKey   = "verificationToken"
	resetEmailSessionKey          = "resetEmail"
	resetTokenSessionKey          = "resetToken"
	isAuthenticatedContextKey     = contextKey("isAuthenticated")
)

func (app *application) login(r *http.Request, userID uuid.UUID) error {
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		return err
	}

	app.sessionManager.Put(r.Context(), authenticatedUserIDSessionKey, userID)

	return nil
}

func (app *application) logout(r *http.Request) error {
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		return err
	}

	app.sessionManager.Remove(r.Context(), authenticatedUserIDSessionKey)

	return nil
}

// Check the auth context set by the authenticate middleware
func (app *application) isAuthenticated(r *http.Request) bool {
	isAuthenticated, ok := r.Context().Value(isAuthenticatedContextKey).(bool)
	if !ok {
		return false
	}

	return isAuthenticated
}

func (app *application) getSessionUserID(r *http.Request) (uuid.UUID, error) {
	id, ok := app.sessionManager.Get(r.Context(), authenticatedUserIDSessionKey).(uuid.UUID)
	if !ok {
		return uuid.UUID{}, fmt.Errorf("unable to parse session id as int")
	}

	return id, nil
}
func (app *application) handleAuthLoginPost(w http.ResponseWriter, r *http.Request) error {
	if app.isAuthenticated(r) {
		return app.renderError(w, "already authenticated", http.StatusBadRequest)
	}

	var form struct {
		Email    string `form:"email" validate:"required,email"`
		Password string `form:"password" validate:"required"`
	}

	err := app.parseForm(r, &form)
	if err != nil {
		return err
	}

	user, err := app.db.Users.GetWithEmail(r.Context(), form.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			// User with email does not exist
			return app.renderError(w, "invalid credentials", http.StatusUnauthorized)
		default:
			return err
		}
	}

	match, err := crypto.ComparePasswordAndHash(form.Password, user.PasswordHash)
	if err != nil {
		return err
	}
	if !match {
		// Incorrect password
		return app.renderError(w, "invalid credentials", http.StatusUnauthorized)
	}

	err = app.login(r, user.ID)
	if err != nil {
		return err
	}

	// Redirect to homepage after authenticating the user.
	http.Redirect(w, r, "/", http.StatusSeeOther)

	return nil
}

func (app *application) handleAuthLogoutPost(w http.ResponseWriter, r *http.Request) error {
	err := app.logout(r)
	if err != nil {
		return err
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)

	return nil
}

func (app *application) handleAuthSignupPost(w http.ResponseWriter, r *http.Request) error {
	if app.isAuthenticated(r) {
		return app.renderError(w, "already authenticated", http.StatusBadRequest)
	}

	var form struct {
		Email string `form:"email" validate:"required,email"`
	}

	err := app.parseForm(r, &form)
	if err != nil {
		return err
	}

	// Check if user with email already exists
	exists, err := app.db.Users.ExistsWithEmail(r.Context(), form.Email)
	if err != nil {
		return err
	}
	if exists {
		// User with email already exists.
		// TODO: respond with message
		app.refresh(w, r)
		return nil
	}

	// Check if a verification token has already been created recently
	exists, err = app.db.VerificationTokens.Exists(r.Context(), data.ScopeRegistration, form.Email)
	if err != nil {
		return err
	}
	if exists {
		// A verfication token has already been sent.
		// TODO: respond with message
		app.refresh(w, r)
		return nil
	}

	// Create new token for user
	token, err := crypto.NewToken(data.VerificationTokenTTL)
	if err != nil {
		return err
	}

	err = app.db.VerificationTokens.New(r.Context(), token.Hash, token.Expiry, data.ScopeRegistration, form.Email)
	if err != nil {
		return err
	}

	// Create link with token
	ref, err := url.Parse("/auth/register")
	if err != nil {
		return err
	}
	q := ref.Query()
	q.Set("email", form.Email)
	q.Set("token", token.Plaintext)
	ref.RawQuery = q.Encode()
	link := app.baseURL.ResolveReference(ref)

	// Send mail in background routine
	app.background(func() error {
		component := emails.Registration(link.String())
		return app.mailer.Send(form.Email, "Registration", component)
	})

	// TODO: respond with message
	app.refresh(w, r)

	return nil
}

func (app *application) handleAuthRegisterGet(w http.ResponseWriter, r *http.Request) error {
	if app.isAuthenticated(r) {
		return app.renderError(w, "already authenticated", http.StatusBadRequest)
	}

	plaintextToken := r.URL.Query().Get("token")
	if plaintextToken == "" {
		return app.renderError(w, "missing verification token", http.StatusBadRequest)
	}

	email := r.URL.Query().Get("email")
	if email == "" {
		return app.renderError(w, "missing verification token", http.StatusBadRequest)
	}

	app.sessionManager.Put(r.Context(), verificationTokenSessionKey, plaintextToken)

	component := pages.Register(nosurf.Token(r), app.popFormErrors(r), email)

	return app.render(w, r, http.StatusOK, "Register", component)
}

func (app *application) handleAuthRegisterPost(w http.ResponseWriter, r *http.Request) error {
	if app.isAuthenticated(r) {
		return app.renderError(w, "already authenticated", http.StatusBadRequest)
	}

	var form struct {
		Email    string `form:"email" validate:"required,email,max=254"`
		Password string `form:"password" validate:"required,min=8,max=72"`
	}

	form.Email = app.sessionManager.GetString(r.Context(), verificationEmailSessionKey)
	err := app.parseForm(r, &form)
	if err != nil {
		return err
	}

	plaintextToken := app.sessionManager.GetString(r.Context(), verificationTokenSessionKey)
	if plaintextToken == "" {
		return app.renderError(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	}

	tokenHash := crypto.TokenHash(plaintextToken)

	err = app.db.VerificationTokens.Verify(r.Context(), tokenHash, data.ScopeRegistration, form.Email)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			return app.renderError(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		}
		if errors.Is(err, data.ErrExpiredToken) {
			// TODO: flash message
			http.Redirect(w, r, "/", http.StatusSeeOther)

			return nil
		}

		return err
	}

	// Upon registration, purge db of all verifications with email.
	err = app.db.VerificationTokens.Purge(r.Context(), form.Email)
	if err != nil {
		return err
	}

	passwordHash, err := crypto.PasswordHash(form.Password)
	if err != nil {
		return err
	}

	user, err := app.db.Users.New(r.Context(), form.Email, passwordHash)
	if err != nil {
		if errors.Is(err, data.ErrDuplicateEmail) {
			return app.renderError(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		}

		return err
	}

	// Login user
	app.sessionManager.Clear(r.Context())
	err = app.login(r, user.ID)
	if err != nil {
		return err
	}

	// TODO: respond with success message created account
	http.Redirect(w, r, "/", http.StatusSeeOther)

	return nil
}

func (app *application) handleAuthResetGet(w http.ResponseWriter, r *http.Request) error {
	email := ""

	// Get users email if already authenticated.
	if app.isAuthenticated(r) {
		suid, err := app.getSessionUserID(r)
		if err != nil {
			return err
		}

		user, err := app.db.Users.Get(r.Context(), suid)
		if err != nil {
			return err
		}

		email = user.Email
	}

	component := pages.ResetPassword(nosurf.Token(r), app.popFormErrors(r), email)

	return app.render(w, r, http.StatusOK, "Reset Password", component)
}

func (app *application) handleAuthResetPost(w http.ResponseWriter, r *http.Request) error {
	var form struct {
		Email string `form:"email" validate:"required,email"`
	}

	err := app.parseForm(r, &form)
	if err != nil {
		return err
	}

	exists, err := app.db.Users.ExistsWithEmail(r.Context(), form.Email)
	if err != nil {
		return err
	}
	if !exists {
		// respond with consistent message email sent
		app.refresh(w, r)

		return nil
	}

	exists, err = app.db.VerificationTokens.Exists(r.Context(), data.ScopePasswordReset, form.Email)
	if err != nil {
		return err
	}
	if exists {
		// respond with consistent message email sent
		app.refresh(w, r)

		return nil
	}

	token, err := crypto.NewToken(data.VerificationTokenTTL)
	if err != nil {
		return err
	}

	err = app.db.VerificationTokens.New(r.Context(), token.Hash, token.Expiry, data.ScopePasswordReset, form.Email)
	if err != nil {
		return err
	}

	// Create link to reset password with token and email and token
	// as query parameters.
	ref, err := url.Parse("/auth/reset/update")
	if err != nil {
		return err
	}
	q := ref.Query()
	q.Set("email", form.Email)
	q.Set("token", token.Plaintext)
	ref.RawQuery = q.Encode()
	link := app.baseURL.ResolveReference(ref)

	// Send mail in background routine
	app.background(func() error {
		component := emails.PasswordReset(link.String())
		return app.mailer.Send(form.Email, "Password Reset", component)
	})

	// respond with consistent message email sent
	app.refresh(w, r)

	return nil
}

func (app *application) handleAuthResetUpdateGet(w http.ResponseWriter, r *http.Request) error {
	plaintextToken := r.URL.Query().Get("token")
	if plaintextToken == "" {
		return app.renderError(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	}

	email := r.URL.Query().Get("email")
	if email == "" {
		return app.renderError(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	}

	component := pages.UpdatePassword(nosurf.Token(r), app.popFormErrors(r), plaintextToken, email)

	return app.render(w, r, http.StatusOK, "Update Password", component)
}

func (app *application) handleAuthResetUpdatePost(w http.ResponseWriter, r *http.Request) error {
	var form struct {
		PlaintextToken string `form:"token" validate:"required"`
		Email          string `form:"email" validate:"required,email,max=254"`
		Password       string `form:"password" validate:"required,min=8,max=72"`
	}

	form.Email = app.sessionManager.GetString(r.Context(), resetEmailSessionKey)
	err := app.parseForm(r, &form)
	if err != nil {
		return err
	}

	tokenHash := crypto.TokenHash(form.PlaintextToken)

	err = app.db.VerificationTokens.Verify(r.Context(), tokenHash, data.ScopePasswordReset, form.Email)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			return app.renderError(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		}
		if errors.Is(err, data.ErrExpiredToken) {
			// TODO: flash expired token. please try again
			http.Redirect(w, r, "/auth/reset", http.StatusSeeOther)

			return nil
		}

		return err
	}

	user, err := app.db.Users.GetWithEmail(r.Context(), form.Email)
	if err != nil {
		return err
	}

	user.PasswordHash, err = crypto.PasswordHash(form.Password)
	if err != nil {
		return err
	}

	err = app.db.Users.Update(r.Context(), user)
	if err != nil {
		return err
	}

	err = app.db.VerificationTokens.Purge(r.Context(), form.Email)
	if err != nil {
		return err
	}

	app.sessionManager.Clear(r.Context())

	// TODO: flash password updated success please login

	http.Redirect(w, r, "/", http.StatusSeeOther)

	return nil
}
