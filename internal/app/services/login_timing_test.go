package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// Login compares against this hash whenever the supplied username matches no
// account, so that the unknown-user path costs the same as the known-user path
// and the response time does not reveal whether a name is registered.
//
// It therefore has to be a real bcrypt hash at the configured cost, and it must
// not accept any input.
func TestUnknownAccountPasswordHashIsAValidHashThatNeverMatches(t *testing.T) {
	cost, err := bcrypt.Cost(unknownAccountPasswordHash)
	require.NoError(t, err)
	assert.Equal(t, bcrypt.DefaultCost, cost)

	candidates := []string{
		"",
		"secret123",
		"tairitsu-unknown-account-placeholder",
		string(unknownAccountPasswordHash),
	}
	for _, candidate := range candidates {
		assert.Error(t, bcrypt.CompareHashAndPassword(unknownAccountPasswordHash, []byte(candidate)))
	}
}
