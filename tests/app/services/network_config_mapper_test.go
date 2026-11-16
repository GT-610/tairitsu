package services

import (
	"testing"

	appservices "github.com/GT-610/tairitsu/internal/app/services"
	"github.com/GT-610/tairitsu/internal/zerotier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeNetworkUpdateRequestForcesPrivateAndTrimsDNS(t *testing.T) {
	notPrivate := false
	broadcast := true
	req := appservices.NormalizeNetworkUpdateRequest(&zerotier.NetworkUpdateRequest{
		Private:         &notPrivate,
		EnableBroadcast: &broadcast,
		DNS: &zerotier.DNSConfig{
			Domain:  " home.arpa ",
			Servers: []string{" 1.1.1.1 ", "", " fd00::53 "},
		},
	})

	require.NotNil(t, req)

	// Managed networks are always private, whatever the request said.
	require.NotNil(t, req.Private)
	assert.True(t, *req.Private)

	// Broadcast is the caller's to set, and is carried through unchanged.
	require.NotNil(t, req.EnableBroadcast)
	assert.True(t, *req.EnableBroadcast)

	require.NotNil(t, req.DNS)
	assert.Equal(t, "home.arpa", req.DNS.Domain)
	assert.Equal(t, []string{"1.1.1.1", "fd00::53"}, req.DNS.Servers)
}

// TestNormalizeNetworkUpdateRequestLeavesUnnamedBroadcastUnset is the other half:
// a request that never mentioned broadcast must not gain a value here.
func TestNormalizeNetworkUpdateRequestLeavesUnnamedBroadcastUnset(t *testing.T) {
	req := appservices.NormalizeNetworkUpdateRequest(&zerotier.NetworkUpdateRequest{
		DNS: &zerotier.DNSConfig{Domain: "home.arpa"},
	})

	require.NotNil(t, req)
	assert.Nil(t, req.EnableBroadcast, "normalisation must not invent a broadcast value")
}
