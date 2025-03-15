package data_test

import (
	"context"
	"testing"
	"time"

	"github.com/micahco/mono/shared/crypto"
	"github.com/micahco/mono/shared/data"
	"github.com/stretchr/testify/assert"
)

func runAuthenticationTokenRepositoryTests(t *testing.T, db *data.DB) {
	ctx := context.Background()
	testEmail := "test@email.com"
	validPassword := []byte("super_secret_password")

	// Create a test user
	testUser, err := db.Users.New(ctx, testEmail, validPassword)
	assert.NoError(t, err)

	// Generate a token
	token, err := crypto.NewToken(data.AuthenticationTokenTTL)
	assert.NoError(t, err)

	t.Run("TestNew", func(t *testing.T) {
		err = db.AuthenticationTokens.New(ctx, token, testUser.ID)
		assert.NoError(t, err)
	})

	t.Run("TestGet", func(t *testing.T) {
		at, err := db.AuthenticationTokens.Get(ctx, token.Hash)
		assert.NoError(t, err)
		assert.NotNil(t, at)
		assert.Equal(t, token.Hash, at.Hash)
		assert.WithinDuration(t, at.Expiry, time.Now(), data.AuthenticationTokenTTL)
		assert.Equal(t, testUser.ID, at.UserID)
	})

	t.Run("TestPurge", func(t *testing.T) {
		err = db.AuthenticationTokens.Purge(ctx, testUser.ID)
		assert.NoError(t, err)

		_, err := db.AuthenticationTokens.Get(ctx, token.Hash)
		assert.ErrorIs(t, err, data.ErrRecordNotFound)
	})
}
