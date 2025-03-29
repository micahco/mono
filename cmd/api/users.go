package main

import (
	"errors"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/micahco/mono/internal/crypto"
	"github.com/micahco/mono/internal/data"
)

const expiredTokenMessage = "expired token"

// Create new user with email and password if provided token
// matches verification.
func (app *application) usersPost(w http.ResponseWriter, r *http.Request) error {
	var input struct {
		Email          string `json:"email"`
		Password       string `json:"password"`
		PlaintextToken string `json:"token"`
	}

	err := app.readJSON(r, &input)
	if err != nil {
		return err
	}

	err = validation.ValidateStruct(&input,
		validation.Field(&input.Email, validation.Required, is.Email),
		validation.Field(&input.Password, validation.Required, passwordLength),
		validation.Field(&input.PlaintextToken, validation.Required),
	)
	if err != nil {
		return err
	}

	tokenHash := crypto.TokenHash(input.PlaintextToken)

	err = app.db.VerificationTokens.Verify(r.Context(), tokenHash, data.ScopeRegistration, input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			return app.writeJSONError(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		case errors.Is(err, data.ErrExpiredToken):
			return app.writeJSONError(w, expiredTokenMessage, http.StatusUnauthorized)
		default:
			return err
		}
	}

	err = app.db.VerificationTokens.Purge(r.Context(), input.Email)
	if err != nil {
		return err
	}

	passwordHash, err := crypto.PasswordHash(input.Password)
	if err != nil {
		return err
	}

	user, err := app.db.Users.New(r.Context(), input.Email, passwordHash)
	if err != nil {
		return err
	}

	res := response{"user": user}

	return app.writeJSON(w, res, http.StatusCreated)
}

// Password reset handler
func (app *application) usersPasswordPut(w http.ResponseWriter, r *http.Request) error {
	var input struct {
		NewPassword    string `json:"password"`
		PlaintextToken string `json:"token"`
	}

	err := app.readJSON(r, &input)
	if err != nil {
		return err
	}

	err = validation.ValidateStruct(&input,
		validation.Field(&input.NewPassword, validation.Required, passwordLength),
		validation.Field(&input.PlaintextToken, validation.Required),
	)
	if err != nil {
		return err
	}

	tokenHash := crypto.TokenHash(input.PlaintextToken)

	user, err := app.db.Users.GetWithVerificationToken(r.Context(), data.ScopePasswordReset, tokenHash)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			return app.writeJSONError(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		case errors.Is(err, data.ErrExpiredToken):
			return app.writeJSONError(w, expiredTokenMessage, http.StatusUnauthorized)
		default:
			return err
		}
	}

	user.PasswordHash, err = crypto.PasswordHash(input.NewPassword)
	if err != nil {
		return err
	}

	err = app.db.Users.Update(r.Context(), user)
	if err != nil {
		switch {
		default:
			return err
		}
	}

	err = app.db.VerificationTokens.Purge(r.Context(), user.Email)
	if err != nil {
		return err
	}

	res := response{"message": "your password was successfully reset"}

	return app.writeJSON(w, res, http.StatusOK)
}

func (app *application) usersMeGet(w http.ResponseWriter, r *http.Request) error {
	user := app.contextGetUser(r.Context())
	res := response{"user": user}

	return app.writeJSON(w, res, http.StatusOK)
}

// Every field is optional. Updating email requires a verificaiton token.
func (app *application) usersMePut(w http.ResponseWriter, r *http.Request) error {
	var input struct {
		Email          *string `json:"email"`
		Password       *string `json:"password"`
		PlaintextToken *string `json:"token"`
	}

	err := app.readJSON(r, &input)
	if err != nil {
		return err
	}

	err = validation.ValidateStruct(&input,
		validation.Field(&input.Email, is.Email),
		validation.Field(&input.Password, passwordLength),
	)
	if err != nil {
		return err
	}

	user := app.contextGetUser(r.Context())

	// Update user email address
	if input.Email != nil {
		if input.PlaintextToken == nil {
			return app.writeJSONError(w, "missing token", http.StatusUnauthorized)
		}

		tokenHash := crypto.TokenHash(*input.PlaintextToken)
		err = app.db.VerificationTokens.Verify(r.Context(), tokenHash, data.ScopeEmailChange, *input.Email)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				return app.writeJSONError(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			case errors.Is(err, data.ErrExpiredToken):
				return app.writeJSONError(w, expiredTokenMessage, http.StatusUnauthorized)
			default:
				return err
			}
		}

		err = app.db.VerificationTokens.Purge(r.Context(), user.Email)
		if err != nil {
			return err
		}

		user.Email = *input.Email
	}

	// Update user password
	if input.Password != nil {
		user.PasswordHash, err = crypto.PasswordHash(*input.Password)
		if err != nil {
			return err
		}
	}

	err = app.db.Users.Update(r.Context(), user)
	if err != nil {
		return err
	}

	res := response{"user": user}

	return app.writeJSON(w, res, http.StatusCreated)
}
