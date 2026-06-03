package services

import (
	"strings"

	"github.com/GT-610/tairitsu/internal/zerotier"
)

func NormalizeNetworkUpdateRequest(req *zerotier.NetworkUpdateRequest) *zerotier.NetworkUpdateRequest {
	if req == nil {
		return nil
	}

	normalized := *req
	normalized.Private = true
	normalized.Routes = cloneRoutes(req.Routes)
	normalized.IpAssignmentPools = cloneAssignmentPools(req.IpAssignmentPools)

	if req.DNS != nil {
		dns := zerotier.DNSConfig{
			Domain:  strings.TrimSpace(req.DNS.Domain),
			Servers: make([]string, 0, len(req.DNS.Servers)),
		}
		for _, server := range req.DNS.Servers {
			trimmed := strings.TrimSpace(server)
			if trimmed != "" {
				dns.Servers = append(dns.Servers, trimmed)
			}
		}
		normalized.DNS = &dns
	}

	return &normalized
}

func cloneRoutes(routes []zerotier.Route) []zerotier.Route {
	if len(routes) == 0 {
		return nil
	}

	cloned := make([]zerotier.Route, len(routes))
	copy(cloned, routes)
	return cloned
}

func cloneAssignmentPools(pools []zerotier.IpAssignmentPool) []zerotier.IpAssignmentPool {
	if len(pools) == 0 {
		return nil
	}

	cloned := make([]zerotier.IpAssignmentPool, len(pools))
	copy(cloned, pools)
	return cloned
}
