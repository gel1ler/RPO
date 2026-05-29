package auth

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateAPISecret returns a random hex secret for terminal device auth.
func GenerateAPISecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
