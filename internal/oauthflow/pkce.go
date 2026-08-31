// Package oauthflow implements the client-side mechanics of the Authorization
// Code + PKCE grant (RFC 6749, RFC 7636) that faro-helm-cli uses to sign in:
// PKCE generation, the loopback callback server, authorize-URL construction,
// and browser launching. It has no dependency on Bubble Tea or on
// faro-helm-cli's own config/API packages — internal/auth wires this
// together with persistence and the token-exchange HTTP call.
package oauthflow

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

// ClientID is faro-helm-cli's registered OAuth 2.0 client_id in
// core.oauth_clients (faro-db). Public client — no client_secret.
const ClientID = "faro-helm-cli"

// PKCE holds a code_verifier and its S256 code_challenge (RFC 7636).
type PKCE struct {
	Verifier  string
	Challenge string
}

// GeneratePKCE creates a new code_verifier/code_challenge pair. The verifier
// is 32 random bytes, base64url-encoded (unpadded) — 43 characters, within
// RFC 7636's required 43-128 range. The challenge is S256(verifier),
// base64url-encoded (unpadded) — 'plain' is never used.
func GeneratePKCE() (PKCE, error) {
	verifier, err := randomURLSafeString(32)
	if err != nil {
		return PKCE{}, err
	}
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])
	return PKCE{Verifier: verifier, Challenge: challenge}, nil
}

// GenerateState creates a random CSRF token for the authorize request's
// state parameter.
func GenerateState() (string, error) {
	return randomURLSafeString(24)
}

func randomURLSafeString(numBytes int) (string, error) {
	b := make([]byte, numBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
