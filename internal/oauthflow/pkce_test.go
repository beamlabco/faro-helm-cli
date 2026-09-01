package oauthflow

import (
	"crypto/sha256"
	"encoding/base64"
	"strings"
	"testing"
)

func TestGeneratePKCE_ChallengeMatchesVerifier(t *testing.T) {
	pkce, err := GeneratePKCE()
	if err != nil {
		t.Fatalf("GeneratePKCE() returned error: %v", err)
	}

	sum := sha256.Sum256([]byte(pkce.Verifier))
	want := base64.RawURLEncoding.EncodeToString(sum[:])

	if pkce.Challenge != want {
		t.Errorf("Challenge = %q, want S256(Verifier) = %q", pkce.Challenge, want)
	}
}

func TestGeneratePKCE_VerifierLengthWithinRFC7636Bounds(t *testing.T) {
	// RFC 7636 §4.1: code_verifier is 43-128 characters.
	pkce, err := GeneratePKCE()
	if err != nil {
		t.Fatalf("GeneratePKCE() returned error: %v", err)
	}

	if len(pkce.Verifier) < 43 || len(pkce.Verifier) > 128 {
		t.Errorf("len(Verifier) = %d, want between 43 and 128", len(pkce.Verifier))
	}
}

func TestGeneratePKCE_NoPaddingCharacters(t *testing.T) {
	// base64url with padding ('=') isn't a valid unreserved character per
	// RFC 7636 §4.1 — must use raw (unpadded) encoding.
	pkce, err := GeneratePKCE()
	if err != nil {
		t.Fatalf("GeneratePKCE() returned error: %v", err)
	}
	if strings.Contains(pkce.Verifier, "=") || strings.Contains(pkce.Challenge, "=") {
		t.Error("verifier/challenge must not contain base64 padding ('=')")
	}
}

func TestGeneratePKCE_UniquePerCall(t *testing.T) {
	a, err := GeneratePKCE()
	if err != nil {
		t.Fatalf("GeneratePKCE() returned error: %v", err)
	}
	b, err := GeneratePKCE()
	if err != nil {
		t.Fatalf("GeneratePKCE() returned error: %v", err)
	}
	if a.Verifier == b.Verifier {
		t.Error("two consecutive calls returned the same verifier")
	}
}

func TestGenerateState_UniquePerCall(t *testing.T) {
	a, err := GenerateState()
	if err != nil {
		t.Fatalf("GenerateState() returned error: %v", err)
	}
	b, err := GenerateState()
	if err != nil {
		t.Fatalf("GenerateState() returned error: %v", err)
	}
	if a == b {
		t.Error("two consecutive calls returned the same state")
	}
	if len(a) < 16 {
		t.Errorf("len(state) = %d, want a reasonably unguessable length (>=16)", len(a))
	}
}
