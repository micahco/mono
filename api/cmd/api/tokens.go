package main

import (
	"context"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/micahco/mono/shared/crypto"
	"github.com/micahco/mono/shared/data"
)

const (
	verificationMsg  = "A verification email has been sent. Please check your inbox."
	passwordResetMsg = "If that email address is in our database, a token to reset your password will be sent to that address."
)

// Create a verification token with registration scope and
// mail it to the provided email address.
func (app *application) tokensVerificaitonRegistrationPost(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
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
	exists, err := app.db.Users.ExistsWithEmail(ctx, input.Email)
	if err != nil {
		return err
	}
	if exists {
		// User with email already exists. Send the
		// consistent respone message.
		return app.writeJSON(w, http.StatusOK, msg, nil)
	}

	// Check if a verification token has already been created recently
	exists, err = app.db.VerificationTokens.Exists(ctx, data.ScopeRegistration, input.Email)
	if err != nil {
		return err
	}
	if exists {
		// Recent verification sent, don't mail another.
		// Send the same message.
		return app.writeJSON(w, http.StatusOK, msg, nil)
	}

	token, err := crypto.NewToken(data.VerificationTokenTTL)
	if err != nil {
		return err
	}

	err = app.db.VerificationTokens.New(ctx, token, data.ScopeRegistration, input.Email)
	if err != nil {
		return err
	}

	// Mail the plaintext token to the user's email address.
	app.background(func() error {
		data := map[string]any{
			"token": token.Plaintext,
		}

		return app.sendMail(input.Email, "registration.tmpl", data)
	})

	return app.writeJSON(w, http.StatusOK, msg, nil)
}

func (app *application) tokensVerificaitonEmailChangePost(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
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
	exists, err := app.db.Users.ExistsWithEmail(ctx, input.Email)
	if err != nil {
		return err
	}
	if exists {
		// User with email already exists. Send the
		// consistent respone message.
		return app.writeJSON(w, http.StatusOK, msg, nil)
	}

	// Check if a verification token has already been created recently
	exists, err = app.db.VerificationTokens.Exists(ctx, data.ScopeEmailChange, input.Email)
	if err != nil {
		return err
	}
	if exists {
		// Recent verification sent, don't mail another, just
		// send the same message
		return app.writeJSON(w, http.StatusOK, msg, nil)
	}

	token, err := crypto.NewToken(data.VerificationTokenTTL)
	if err != nil {
		return err
	}

	// Create verification token for user with new email address
	err = app.db.VerificationTokens.New(ctx, token, data.ScopeEmailChange, input.Email)
	if err != nil {
		return err
	}

	// Mail the plaintext token to the new email address
	app.background(func() error {
		data := map[string]any{
			"token": token.Plaintext,
		}

		return app.sendMail(input.Email, "email-change.tmpl", data)
	})

	return app.writeJSON(w, http.StatusOK, msg, nil)
}

func (app *application) tokensVerificaitonPasswordResetPost(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
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
	exists, err := app.db.Users.ExistsWithEmail(ctx, input.Email)
	if err != nil {
		return err
	}
	if !exists {
		// User with email does not exist. Still send the same
		// message.
		return app.writeJSON(w, http.StatusOK, msg, nil)
	}

	// Check if a verification token has already been created recently
	exists, err = app.db.VerificationTokens.Exists(ctx, data.ScopePasswordReset, input.Email)
	if err != nil {
		return err
	}
	if exists {
		// Recent verification sent, don't mail another
		return app.writeJSON(w, http.StatusOK, msg, nil)
	}

	token, err := crypto.NewToken(data.VerificationTokenTTL)
	if err != nil {
		return err
	}

	// Create verification token for user with email address
	err = app.db.VerificationTokens.New(ctx, token, data.ScopePasswordReset, input.Email)
	if err != nil {
		return err
	}

	// Mail the plaintext token to the provided email address
	app.background(func() error {
		data := map[string]any{
			"token": token.Plaintext,
		}

		return app.sendMail(input.Email, "password-reset.tmpl", data)
	})

	return app.writeJSON(w, http.StatusOK, msg, nil)
}

func (app *application) tokensAuthenticationPost(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
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

	passwordHash, err := crypto.PasswordHash(input.Password)
	if err != nil {
		return err
	}

	user, err := app.db.Users.GetForCredentials(ctx, input.Email, passwordHash)
	if err != nil {
		if err == data.ErrInvalidCredentials {
			return app.writeError(w, http.StatusUnauthorized, InvalidCredentailsMessage)
		}

		return err
	}

	token, err := crypto.NewToken(data.VerificationTokenTTL)
	if err != nil {
		return err
	}

	err = app.db.AuthenticationTokens.New(ctx, token, user.ID)
	if err != nil {
		return err
	}

	return app.writeJSON(w, http.StatusCreated, envelope{"authentication_token": token.Plaintext}, nil)
}
