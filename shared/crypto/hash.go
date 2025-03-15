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

func NewToken(ttl time.Duration) (Token, error) {
	var token Token

	// Initialize a zero-valued byte slice with a length of 16 bytes.
	b := make([]byte, 16)

	// Fill b with random bytes from CSPRNG
	_, err := rand.Read(b)
	if err != nil {
		return token, err
	}

	// Note that by default base-32 strings may be padded at the end with the =
	// character. We don't need this padding character for the purpose of our tokens, so
	// we use the WithPadding(base32.NoPadding) method in the line below to omit them.
	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)
	token.Hash = TokenHash(token.Plaintext)
	token.Expiry = time.Now().Add(ttl)

	return token, nil
}

// Generate a SHA-256 hash of the plaintext token string
func TokenHash(plaintextToken string) []byte {
	hash := sha256.Sum256([]byte(plaintextToken))
	return hash[:] // convert array to slice
}

func PasswordHash(password string) ([]byte, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return nil, err
	}

	return []byte(hash), nil
}
