package services

import (
	"testing"

	appservices "github.com/GT-610/tairitsu/internal/app/services"
	"github.com/GT-610/tairitsu/internal/zerotier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSplitIPAssignmentPoolsByFamily(t *testing.T) {
	ipv4, ipv6 := appservices.SplitIPAssignmentPoolsByFamily([]zerotier.IpAssignmentPool{
		{IpRangeStart: "10.0.0.1", IpRangeEnd: "10.0.0.10"},
		{IpRangeStart: "fd00::1", IpRangeEnd: "fd00::ff"},
		{IpRangeStart: "invalid", IpRangeEnd: "invalid"},
	})

	require.Len(t, ipv4, 1)
	require.Len(t, ipv6, 1)
	assert.Equal(t, "10.0.0.1", ipv4[0].IpRangeStart)
	assert.Equal(t, "fd00::1", ipv6[0].IpRangeStart)
}

func TestExtractManagedRoutesPreservesAdditionalRoutesOnly(t *testing.T) {
	managed := appservices.ExtractManagedRoutes([]zerotier.Route{
		{Target: "10.0.0.0/24"},
		{Target: "fd00::/64"},
		{Target: "10.1.0.0/24", Via: "10.0.0.1"},
		{Target: "fd00:1::/64", Via: "fd00::1"},
	})

	assert.Equal(t, []zerotier.Route{
		{Target: "10.1.0.0/24", Via: "10.0.0.1"},
		{Target: "fd00:1::/64", Via: "fd00::1"},
	}, managed)
}

func TestBuildManagedRoutesMergesPrimaryAndManagedRoutes(t *testing.T) {
	primaryIPv4 := &zerotier.Route{Target: "10.0.0.0/24"}
	primaryIPv6 := &zerotier.Route{Target: "fd00::/64"}
	managedRoutes := []zerotier.Route{{Target: "10.1.0.0/24", Via: "10.0.0.1"}}

	routes := appservices.BuildManagedRoutes(primaryIPv4, primaryIPv6, managedRoutes)

	assert.Equal(t, []zerotier.Route{
		{Target: "10.0.0.0/24"},
		{Target: "fd00::/64"},
		{Target: "10.1.0.0/24", Via: "10.0.0.1"},
	}, routes)
}

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
