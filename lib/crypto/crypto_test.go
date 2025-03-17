package crypto_test

import (
	"testing"

	"github.com/micahco/mono/lib/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewToken(t *testing.T) {
	plainTextToken, err := crypto.GeneratePlaintextToken()
	require.NoError(t, err)
	assert.NotEmpty(t, plainTextToken)
}

func TestTokenHash(t *testing.T) {
	input := "test_token"
	hash1 := crypto.TokenHash(input)
	hash2 := crypto.TokenHash(input)
	assert.Equal(t, hash1, hash2)

	hash3 := crypto.TokenHash("other_token")
	assert.NotEqual(t, hash1, hash3)
}

func TestPasswordHash(t *testing.T) {
	plaintextPassword := "super_secure_password"
	hash, err := crypto.PasswordHash(plaintextPassword)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)

	hash2, _ := crypto.PasswordHash(plaintextPassword)
	assert.NotEqual(t, hash, hash2)
}

func TestComparePasswordAndHash(t *testing.T) {
	plaintextPassword := "super_secure_password"
	hash, err := crypto.PasswordHash(plaintextPassword)
	require.NoError(t, err)
	match, err := crypto.ComparePasswordAndHash(plaintextPassword, hash)
	assert.NoError(t, err)
	assert.True(t, match)
}
