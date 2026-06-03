package services

import (
	"testing"

	appservices "github.com/GT-610/tairitsu/internal/app/services"
	"github.com/GT-610/tairitsu/internal/zerotier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeNetworkUpdateRequestForcesPrivateAndTrimsDNS(t *testing.T) {
	req := appservices.NormalizeNetworkUpdateRequest(&zerotier.NetworkUpdateRequest{
		Private: false,
		DNS: &zerotier.DNSConfig{
			Domain:  " home.arpa ",
			Servers: []string{" 1.1.1.1 ", "", " fd00::53 "},
		},
	})

	require.NotNil(t, req)
	assert.True(t, req.Private)
	require.NotNil(t, req.DNS)
	assert.Equal(t, "home.arpa", req.DNS.Domain)
	assert.Equal(t, []string{"1.1.1.1", "fd00::53"}, req.DNS.Servers)
}
