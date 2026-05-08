package codex

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// PKCE implements Proof Key for Code Exchange (RFC 7636).
type PKCE struct {
	Verifier  string
	Challenge string
	Method    string
}

// GeneratePKCE creates a new PKCE code verifier and challenge pair.
func GeneratePKCE() (*PKCE, error) {
	verifier, err := generateVerifier(43) // 43 bytes → 43 URL-safe base64 chars
	if err != nil {
		return nil, fmt.Errorf("pkce: failed to generate verifier: %w", err)
	}

	challenge := computeChallenge(verifier)

	return &PKCE{
		Verifier:  verifier,
		Challenge: challenge,
		Method:    "S256",
	}, nil
}

func generateVerifier(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func computeChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}
