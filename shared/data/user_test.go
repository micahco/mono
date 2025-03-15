package data_test

import (
	"context"
	"testing"

	"github.com/micahco/mono/shared/data"
	"github.com/micahco/mono/shared/data/internal/crypto"
	"github.com/micahco/mono/shared/data/internal/uuid"
	"github.com/stretchr/testify/assert"
)

func runUserRepositoryTests(t *testing.T, db *data.DB) {
	ctx := context.Background()
	testEmail := "test@email.com"
	updatedEmail := "updated@gmail.com"
	newEmail := "new@email.com"
	nonExistantEmail := "unknown@email.com"
	validPassword := []byte("super_secret_password")
	incorrectPassword := []byte("incorrect_password")
	nonExistantID := uuid.Nil

	var testUser *data.User

	t.Run("TestNew", func(t *testing.T) {
		var err error
		testUser, err = db.Users.New(ctx, testEmail, validPassword)
		assert.NoError(t, err)
		assert.NotNil(t, testUser)
		assert.Equal(t, int32(1), testUser.Version)
		assert.Equal(t, testEmail, testUser.Email)

		// Duplicate email
		_, err = db.Users.New(ctx, testEmail, validPassword)
		assert.ErrorIs(t, err, data.ErrDuplicateEmail)
	})

	t.Run("TestGet", func(t *testing.T) {
		readUser, err := db.Users.Get(ctx, testUser.ID)
		assert.NoError(t, err)
		assert.NotNil(t, readUser)
		assert.Equal(t, testUser, readUser)
	})

	t.Run("TestGetForCredentials", func(t *testing.T) {
		readUser, err := db.Users.GetForCredentials(ctx, testEmail, validPassword)
		assert.NoError(t, err)
		assert.NotNil(t, readUser)
		assert.Equal(t, testUser, readUser)

		// Incorrect credentials
		_, err = db.Users.GetForCredentials(ctx, testEmail, incorrectPassword)
		assert.ErrorIs(t, err, data.ErrInvalidCredentials)
	})

	t.Run("TestGetForAuthenticationToken", func(t *testing.T) {
		// Generate a token
		token, err := crypto.NewToken(data.AuthenticationTokenTTL)
		assert.NoError(t, err)

		// Put token in db
		err = db.AuthenticationTokens.New(ctx, token, testUser.ID)
		assert.NoError(t, err)

		// Get user with token
		readUser, err := db.Users.GetForAuthenticationToken(ctx, token.Hash)
		assert.NoError(t, err)
		assert.NotNil(t, readUser)
		assert.Equal(t, testUser, readUser)
	})

	t.Run("TestExistsWithEmail", func(t *testing.T) {
		exists, err := db.Users.ExistsWithEmail(ctx, testEmail)
		assert.NoError(t, err)
		assert.True(t, exists)

		exists, err = db.Users.ExistsWithEmail(ctx, nonExistantEmail)
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("TestUpdate", func(t *testing.T) {
		testUser.Email = updatedEmail
		currentVersion := testUser.Version
		err := db.Users.Update(ctx, testUser)
		assert.NoError(t, err)
		assert.NotNil(t, testUser)
		assert.Equal(t, currentVersion+1, testUser.Version)

		readUser, err := db.Users.Get(ctx, testUser.ID)
		assert.NoError(t, err)
		assert.NotNil(t, readUser)
		assert.Equal(t, testUser, readUser)

		// Invalid version
		testUser.Version -= 1 // old version
		err = db.Users.Update(ctx, testUser)
		assert.ErrorIs(t, err, data.ErrEditConflict)

		// Duplicate email
		newUser, err := db.Users.New(ctx, newEmail, validPassword)
		assert.NoError(t, err)
		assert.NotNil(t, newUser)
		assert.Equal(t, newEmail, newUser.Email)

		newUser.Email = testEmail
		err = db.Users.Update(ctx, testUser)
		assert.ErrorIs(t, err, data.ErrEditConflict)
	})

	t.Run("TestDelete", func(t *testing.T) {
		err := db.Users.Delete(ctx, testUser.ID)
		assert.NoError(t, err)

		// Non-existant user ID
		err = db.Users.Delete(ctx, nonExistantID)
		assert.ErrorIs(t, err, data.ErrRecordNotFound)
	})
}
