package data_test

import (
	"context"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/micahco/mono/internal/data"
	"github.com/stretchr/testify/assert"
)

func runUserRepositoryTests(t *testing.T, db *data.DB) {
	ctx := context.Background()
	testEmail := "test@email.com"
	updatedEmail := "updated@gmail.com"
	newEmail := "new@email.com"
	nonExistantEmail := "unknown@email.com"
	testPassword := "super_secret_password"
	nonExistantID := uuid.Nil

	var testUser *data.User

	t.Run("TestNew", func(t *testing.T) {
		var err error
		testUser, err = db.Users.New(ctx, testEmail, []byte(testPassword))
		assert.NoError(t, err)
		assert.NotNil(t, testUser)
		assert.Equal(t, int32(1), testUser.Version)
		assert.Equal(t, testEmail, testUser.Email)

		// Duplicate email
		_, err = db.Users.New(ctx, testEmail, []byte(testPassword))
		assert.ErrorIs(t, err, data.ErrDuplicateEmail)
	})

	t.Run("TestGet", func(t *testing.T) {
		readUser, err := db.Users.Get(ctx, testUser.ID)
		assert.NoError(t, err)
		assert.NotNil(t, readUser)
		assert.Equal(t, testUser, readUser)

		_, err = db.Users.Get(ctx, nonExistantID)
		assert.ErrorIs(t, err, data.ErrRecordNotFound)
	})

	t.Run("TestGetWithEmail", func(t *testing.T) {
		readUser, err := db.Users.GetWithEmail(ctx, testEmail)
		assert.NoError(t, err)
		assert.NotNil(t, readUser)
		assert.Equal(t, testUser, readUser)

		_, err = db.Users.GetWithEmail(ctx, nonExistantEmail)
		assert.ErrorIs(t, err, data.ErrRecordNotFound)
	})

	t.Run("TestGetWithVerificationToken", func(t *testing.T) {
		tokenHash := []byte("test_token")
		expiry := time.Now().Add(time.Hour)
		scope := "testing"

		// Put token in db
		err := db.VerificationTokens.New(ctx, tokenHash, expiry, scope, testEmail)
		assert.NoError(t, err)

		// Get user with token
		readUser, err := db.Users.GetWithVerificationToken(ctx, scope, tokenHash)
		assert.NoError(t, err)
		assert.Equal(t, testUser, readUser)
	})

	t.Run("TestGetWithAuthenticationToken", func(t *testing.T) {
		tokenHash := []byte("test_token")
		expiry := time.Now().Add(time.Hour)

		// Put token in db
		err := db.AuthenticationTokens.New(ctx, tokenHash, expiry, testUser.ID)
		assert.NoError(t, err)

		// Get user with token
		readUser, err := db.Users.GetWithAuthenticationToken(ctx, tokenHash)
		assert.NoError(t, err)
		assert.Equal(t, testUser, readUser)
	})

	t.Run("TestExists", func(t *testing.T) {
		exists, err := db.Users.Exists(ctx, testUser.ID)
		assert.NoError(t, err)
		assert.True(t, exists)

		exists, err = db.Users.Exists(ctx, nonExistantID)
		assert.NoError(t, err)
		assert.False(t, exists)
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
		newUser, err := db.Users.New(ctx, newEmail, []byte(testPassword))
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
