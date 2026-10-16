package services

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
)

// setupTokenBytes is the entropy of the first-run setup token.
const setupTokenBytes = 16

// SetupToken is the secret that gates the first-run setup endpoints.
//
// The setup surface has to stay reachable before any account exists, so access
// cannot be decided by authentication. Instead a token is generated on every
// start while the instance is uninitialized and written to the process output,
// which only someone who can read the deployment's logs can see. It is never
// persisted: the state it protects is the running process, so a restart simply
// issues a new one.
type SetupToken struct {
	value string
}

// NewSetupToken generates a fresh setup token.
func NewSetupToken() (*SetupToken, error) {
	buffer := make([]byte, setupTokenBytes)
	if _, err := rand.Read(buffer); err != nil {
		return nil, fmt.Errorf("failed to generate setup token: %w", err)
	}

	return &SetupToken{value: base64.RawURLEncoding.EncodeToString(buffer)}, nil
}

// Value returns the token to present to the operator.
func (t *SetupToken) Value() string {
	if t == nil {
		return ""
	}
	return t.value
}

// Issued reports whether a token exists, which is only true while setup is
// still pending.
func (t *SetupToken) Issued() bool {
	return t.Value() != ""
}

// Matches reports whether candidate is the issued token. The comparison is
// constant time so that a wrong guess does not leak the token one byte at a
// time.
func (t *SetupToken) Matches(candidate string) bool {
	issued := t.Value()
	if issued == "" || candidate == "" {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(issued), []byte(candidate)) == 1
}
