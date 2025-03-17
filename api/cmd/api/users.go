package main

import (
	"context"
	"errors"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/micahco/mono/shared/crypto"
	"github.com/micahco/mono/shared/data"
)

// Create new user with email and password if provided token
// matches verification.
func (app *application) usersPost(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
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
		// TODO: generalize password length values
		validation.Field(&input.Password, validation.Required, passwordLength),
		validation.Field(&input.PlaintextToken, validation.Required),
	)
	if err != nil {
		return err
	}

	tokenHash := crypto.TokenHash(input.PlaintextToken)

	err = app.db.VerificationTokens.Verify(ctx, tokenHash, data.ScopeRegistration, input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			return app.writeError(w, http.StatusUnauthorized, nil)
		case errors.Is(err, data.ErrExpiredToken):
			return app.writeError(w, http.StatusUnauthorized, "Expired token")
		default:
			return err
		}
	}

	err = app.db.VerificationTokens.Purge(ctx, input.Email)
	if err != nil {
		return err
	}

	passwordHash, err := crypto.PasswordHash(input.Password)
	if err != nil {
		return err
	}

	user, err := app.db.Users.New(ctx, input.Email, passwordHash)
	if err != nil {
		return err
	}

	return app.writeJSON(w, http.StatusCreated, envelope{"user": user}, nil)
}

// Password reset handler
func (app *application) usersPasswordPut(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
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

	user, err := app.db.Users.GetWithVerificationToken(ctx, data.ScopePasswordReset, tokenHash)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			return app.writeError(w, http.StatusUnauthorized, nil)
		case errors.Is(err, data.ErrExpiredToken):
			return app.writeError(w, http.StatusUnauthorized, "Expired token")
		default:
			return err
		}
	}

	user.PasswordHash, err = crypto.PasswordHash(input.NewPassword)
	if err != nil {
		return err
	}

	err = app.db.Users.Update(ctx, user)
	if err != nil {
		switch {
		default:
			return err
		}
	}

	err = app.db.VerificationTokens.Purge(ctx, user.Email)
	if err != nil {
		return err
	}

	msg := envelope{"message": "your password was successfully reset"}

	return app.writeJSON(w, http.StatusOK, msg, nil)
}

func (app *application) usersMeGet(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	user := app.contextGetUser(ctx)

	return app.writeJSON(w, http.StatusOK, envelope{"user": user}, nil)
}

// Every field is optional. Updating email requires a verificaiton token.
func (app *application) usersMePut(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
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

	user := app.contextGetUser(ctx)

	// Update user email address
	if input.Email != nil {
		if input.PlaintextToken == nil {
			return app.writeError(w, http.StatusUnauthorized, "Missing token")
		}

		tokenHash := crypto.TokenHash(*input.PlaintextToken)
		err = app.db.VerificationTokens.Verify(ctx, tokenHash, data.ScopeEmailChange, *input.Email)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				return app.writeError(w, http.StatusUnauthorized, nil)
			case errors.Is(err, data.ErrExpiredToken):
				return app.writeError(w, http.StatusUnauthorized, "Expired token")
			default:
				return err
			}
		}

		err = app.db.VerificationTokens.Purge(ctx, user.Email)
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

	err = app.db.Users.Update(ctx, user)
	if err != nil {
		return err
	}

	return app.writeJSON(w, http.StatusCreated, envelope{"user": user}, nil)
}
