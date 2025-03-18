package main

import (
	"errors"
	"net/http"
	"time"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/micahco/mono/lib/crypto"
	"github.com/micahco/mono/lib/data"
)

const (
	verificationMsg  = "A verification email has been sent. Please check your inbox."
	passwordResetMsg = "If that email address is in our database, a token to reset your password will be sent to that address."
)

type Token struct {
	Plaintext string
	Hash      []byte
	Expiry    time.Time
}

func newToken(ttl time.Duration) (Token, error) {
	var token Token
	var err error

	token.Plaintext, err = crypto.GeneratePlaintextToken()
	if err != nil {
		return token, err
	}

	token.Hash = crypto.TokenHash(token.Plaintext)
	token.Expiry = time.Now().Add(ttl)

	return token, nil
}

// Create a verification token with registration scope and
// mail it to the provided email address.
func (app *application) tokensVerificaitonRegistrationPost(w http.ResponseWriter, r *http.Request) error {
	var input struct {
		Email string `json:"email"`
	}

	err := app.readJSON(r, &input)
	if err != nil {
		return err
	}

	err = validation.ValidateStruct(&input,
		validation.Field(&input.Email, validation.Required, is.Email),
	)
	if err != nil {
		return err
	}

	// This will be the consistent message. Even if a user
	// already exists with this email, send this message.
	msg := envelope{"message": verificationMsg}

	// Check if user with email already exists
	exists, err := app.db.Users.ExistsWithEmail(r.Context(), input.Email)
	if err != nil {
		return err
	}
	if exists {
		// User with email already exists. Send the
		// consistent respone message.
		return app.writeJSON(w, http.StatusOK, msg, nil)
	}

	// Check if a verification token has already been created recently
	exists, err = app.db.VerificationTokens.Exists(r.Context(), data.ScopeRegistration, input.Email)
	if err != nil {
		return err
	}
	if exists {
		// Recent verification sent, don't mail another.
		// Send the same message.
		return app.writeJSON(w, http.StatusOK, msg, nil)
	}

	// Create new token for user
	token, err := newToken(data.VerificationTokenTTL)
	if err != nil {
		return err
	}

	err = app.db.VerificationTokens.New(r.Context(), token.Hash, token.Expiry, data.ScopeRegistration, input.Email)
	if err != nil {
		return err
	}

	// Mail the plaintext token to the user's email address.
	app.background(func() error {
		data := map[string]any{
			"token": token.Plaintext,
		}

		return app.mailer.Send(input.Email, "registration.tmpl", data)
	})

	return app.writeJSON(w, http.StatusOK, msg, nil)
}

func (app *application) tokensVerificaitonEmailChangePost(w http.ResponseWriter, r *http.Request) error {
	var input struct {
		Email string `json:"email"`
	}

	err := app.readJSON(r, &input)
	if err != nil {
		return err
	}

	err = validation.ValidateStruct(&input,
		validation.Field(&input.Email, validation.Required, is.Email),
	)
	if err != nil {
		return err
	}

	// This will be the consistent message. Even if a user
	// already exists with this email, send this message.
	msg := envelope{"message": verificationMsg}

	// Check if user with email already exists
	exists, err := app.db.Users.ExistsWithEmail(r.Context(), input.Email)
	if err != nil {
		return err
	}
	if exists {
		// User with email already exists. Send the
		// consistent respone message.
		return app.writeJSON(w, http.StatusOK, msg, nil)
	}

	// Check if a verification token has already been created recently
	exists, err = app.db.VerificationTokens.Exists(r.Context(), data.ScopeEmailChange, input.Email)
	if err != nil {
		return err
	}
	if exists {
		// Recent verification sent, don't mail another, just
		// send the same message
		return app.writeJSON(w, http.StatusOK, msg, nil)
	}

	// Create verification token for user with new email address
	token, err := newToken(data.VerificationTokenTTL)
	if err != nil {
		return err
	}

	err = app.db.VerificationTokens.New(r.Context(), token.Hash, token.Expiry, data.ScopeEmailChange, input.Email)
	if err != nil {
		return err
	}

	// Mail the plaintext token to the new email address
	app.background(func() error {
		data := map[string]any{
			"token": token.Plaintext,
		}

		return app.mailer.Send(input.Email, "email-change.tmpl", data)
	})

	return app.writeJSON(w, http.StatusOK, msg, nil)
}

func (app *application) tokensVerificaitonPasswordResetPost(w http.ResponseWriter, r *http.Request) error {
	var input struct {
		Email string `json:"email"`
	}

	err := app.readJSON(r, &input)
	if err != nil {
		return err
	}

	err = validation.ValidateStruct(&input,
		validation.Field(&input.Email, validation.Required, is.Email),
	)
	if err != nil {
		return err
	}

	// This will be the consistent message. Even if a user
	// already exists with this email, send this message.
	msg := envelope{"message": verificationMsg}

	// Check if user with email exists
	exists, err := app.db.Users.ExistsWithEmail(r.Context(), input.Email)
	if err != nil {
		return err
	}
	if !exists {
		// User with email does not exist. Still send the same
		// message.
		return app.writeJSON(w, http.StatusOK, msg, nil)
	}

	// Check if a verification token has already been created recently
	exists, err = app.db.VerificationTokens.Exists(r.Context(), data.ScopePasswordReset, input.Email)
	if err != nil {
		return err
	}
	if exists {
		// Recent verification sent, don't mail another
		return app.writeJSON(w, http.StatusOK, msg, nil)
	}

	token, err := newToken(data.VerificationTokenTTL)
	if err != nil {
		return err
	}

	// Create verification token for user with email address
	err = app.db.VerificationTokens.New(r.Context(), token.Hash, token.Expiry, data.ScopePasswordReset, input.Email)
	if err != nil {
		return err
	}

	// Mail the plaintext token to the provided email address
	app.background(func() error {
		data := map[string]any{
			"token": token.Plaintext,
		}

		return app.mailer.Send(input.Email, "password-reset.tmpl", data)
	})

	return app.writeJSON(w, http.StatusOK, msg, nil)
}

func (app *application) tokensAuthenticationPost(w http.ResponseWriter, r *http.Request) error {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(r, &input)
	if err != nil {
		return err
	}

	err = validation.ValidateStruct(&input,
		validation.Field(&input.Email, validation.Required, is.Email),
		validation.Field(&input.Password, validation.Required, passwordLength),
	)
	if err != nil {
		return err
	}

	user, err := app.db.Users.GetWithEmail(r.Context(), input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			// User with email does not exist
			return app.writeError(w, http.StatusUnauthorized, InvalidCredentailsMessage)
		default:
			return err
		}
	}

	match, err := crypto.ComparePasswordAndHash(input.Password, user.PasswordHash)
	if err != nil {
		return err
	}
	if !match {
		// Incorrect password
		return app.writeError(w, http.StatusUnauthorized, InvalidCredentailsMessage)
	}

	// Create authentication token for user with new email address
	token, err := newToken(data.AuthenticationTokenTTL)
	if err != nil {
		return err
	}

	err = app.db.AuthenticationTokens.New(r.Context(), token.Hash, token.Expiry, user.ID)
	if err != nil {
		return err
	}

	return app.writeJSON(w, http.StatusCreated, envelope{"authentication_token": token.Plaintext}, nil)
}
