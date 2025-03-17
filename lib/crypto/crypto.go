package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"time"

	"github.com/alexedwards/argon2id"
)

type Token struct {
	Plaintext string
	Hash      []byte
	Expiry    time.Time
}

// Generate a random plaintext token base-32 string
func GeneratePlaintextToken() (string, error) {
	// Initialize a zero-valued byte slice with a length of 16 bytes.
	b := make([]byte, 16)

	// Fill b with random bytes from CSPRNG
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	// Note that by default base-32 strings may be padded at the end with the =
	// character. We don't need this padding character for the purpose of our tokens, so
	// we use the WithPadding(base32.NoPadding) method in the line below to omit them.
	plaintextToken := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)

	return plaintextToken, nil
}

// Generate a SHA-256 hash of the plaintext token string
func TokenHash(plaintextToken string) []byte {
	hash := sha256.Sum256([]byte(plaintextToken))
	return hash[:] // convert array to slice
}

// Generate cryptographically secure password hash
func PasswordHash(plaintextPassword string) ([]byte, error) {
	hash, err := argon2id.CreateHash(plaintextPassword, argon2id.DefaultParams)
	if err != nil {
		return nil, err
	}

	return []byte(hash), nil
}

// Compare plaintext password with hash. It returns true if they match, otherwise it returns false.
func ComparePasswordAndHash(plaintextPassword string, passwordHash []byte) (bool, error) {
	return argon2id.ComparePasswordAndHash(plaintextPassword, string(passwordHash))
}
