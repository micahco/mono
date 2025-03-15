package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"time"

	"github.com/alexedwards/argon2id"
)

func CreatePasswordHash(password string) ([]byte, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return nil, err
	}

	return []byte(hash), nil
}

func CreateTokenHash(str string) []byte {
	hash := sha256.Sum256([]byte(str))

	return hash[:] // converts array to slice
}

const TokenSize = 16

func GenerateTokenPlaintext(ttl time.Duration) (string, error) {
	b := make([]byte, TokenSize)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	// Note that by default base-32 strings may be padded at the end with the =
	// character. We don't need this padding character for the purpose of our tokens, so
	// we use the WithPadding(base32.NoPadding) method in the line below to omit them.
	s := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)

	return s, nil
}
