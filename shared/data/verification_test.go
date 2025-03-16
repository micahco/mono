package data_test

import (
	"context"
	"testing"
	"time"

	"github.com/micahco/mono/shared/data"
	"github.com/stretchr/testify/assert"
)

func runVerificationTokenRepositoryTests(t *testing.T, db *data.DB) {
	ctx := context.Background()
	testEmail := "test@email.com"
	nonExistantEmail := "unknown@email.com"

	tokenHash := []byte("test_token")
	expiry := time.Now().Add(time.Hour)

	t.Run("TestNew", func(t *testing.T) {
		err := db.VerificationTokens.New(ctx, tokenHash, expiry, data.ScopeRegistration, testEmail)
		assert.NoError(t, err)
	})

	t.Run("TestGet", func(t *testing.T) {
		vt, err := db.VerificationTokens.Get(ctx, tokenHash)
		assert.NoError(t, err)
		assert.NotNil(t, vt)
		assert.Equal(t, tokenHash, vt.Hash)
		assert.WithinDuration(t, time.Now(), vt.Expiry, data.VerificationTokenTTL)
		assert.Equal(t, data.ScopeRegistration, vt.Scope)
		assert.Equal(t, testEmail, vt.Email)
	})

	t.Run("TestExists", func(t *testing.T) {
		exists, err := db.VerificationTokens.Exists(ctx, data.ScopeRegistration, testEmail)
		assert.NoError(t, err)
		assert.True(t, exists)

		// Incorrect scope
		exists, err = db.VerificationTokens.Exists(ctx, data.ScopeEmailChange, testEmail)
		assert.NoError(t, err)
		assert.False(t, exists)

		// Incorrect email
		exists, err = db.VerificationTokens.Exists(ctx, data.ScopeRegistration, nonExistantEmail)
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("TestPurge", func(t *testing.T) {
		err := db.VerificationTokens.Purge(ctx, testEmail)
		assert.NoError(t, err)

		exists, err := db.VerificationTokens.Exists(ctx, data.ScopeRegistration, testEmail)
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("TestVerify", func(t *testing.T) {
		err := db.VerificationTokens.New(ctx, tokenHash, expiry, data.ScopeRegistration, testEmail)
		assert.NoError(t, err)

		// Valid token
		err = db.VerificationTokens.Verify(ctx, tokenHash, data.ScopeRegistration, testEmail)
		assert.NoError(t, err)

		// Incorrect scope
		err = db.VerificationTokens.Verify(ctx, tokenHash, data.ScopeEmailChange, testEmail)
		assert.ErrorIs(t, err, data.ErrRecordNotFound)

		// Expired token
		expiredHash := []byte("expired_hash")
		expiredExpiry := time.Now().Add(-1 * time.Minute)

		err = db.VerificationTokens.New(ctx, expiredHash, expiredExpiry, data.ScopeRegistration, testEmail)
		assert.NoError(t, err)

		err = db.VerificationTokens.Verify(ctx, expiredHash, data.ScopeRegistration, testEmail)
		assert.ErrorIs(t, err, data.ErrExpiredToken)
	})
}
