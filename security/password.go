package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"

	"golang.org/x/crypto/argon2"
)

// Argon2id parameters tuned for server defaults. Adjust via environment if needed.
var (
	argonTime    uint32 = 1
	argonMemory  uint32 = 64 * 1024 // 64 MB
	argonThreads uint8  = 2
	saltLen      uint32 = 16
	keyLen       uint32 = 32
)

// HashPassword returns encoded hash in the format: base64(salt)|base64(key)
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New("empty password")
	}
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, keyLen)
	return base64.RawStdEncoding.EncodeToString(salt) + "|" + base64.RawStdEncoding.EncodeToString(key), nil
}

// VerifyPassword compares a password to an encoded hash produced by HashPassword.
func VerifyPassword(password, encoded string) (bool, error) {
	if password == "" || encoded == "" {
		return false, errors.New("invalid input")
	}
	parts := [2]string{"", ""}
	sep := -1
	for i := 0; i < len(encoded); i++ {
		if encoded[i] == '|' {
			sep = i
			break
		}
	}
	if sep < 0 {
		return false, errors.New("invalid hash format")
	}
	parts[0] = encoded[:sep]
	parts[1] = encoded[sep+1:]

	salt, err := base64.RawStdEncoding.DecodeString(parts[0])
	if err != nil {
		return false, err
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		return false, err
	}
	have := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, uint32(len(want)))
	if subtle.ConstantTimeCompare(have, want) == 1 {
		return true, nil
	}
	return false, nil
}
