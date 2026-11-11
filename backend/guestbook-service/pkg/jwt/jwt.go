package jwtpkg

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func GenerateRandomString() (string, error) {
	b := make([]byte, 32)

	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("rand.Read: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}
