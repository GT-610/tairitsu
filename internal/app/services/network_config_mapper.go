package services

import (
	"net"
	"strings"

	"github.com/GT-610/tairitsu/internal/zerotier"
)

func SplitIPAssignmentPoolsByFamily(pools []zerotier.IpAssignmentPool) (ipv4 []zerotier.IpAssignmentPool, ipv6 []zerotier.IpAssignmentPool) {
	for _, pool := range pools {
		ip := net.ParseIP(strings.TrimSpace(pool.IpRangeStart))
		switch {
		case ip == nil:
			continue
		case ip.To4() != nil:
			ipv4 = append(ipv4, pool)
		default:
			ipv6 = append(ipv6, pool)
		}
	}

	return ipv4, ipv6
}

func ExtractPrimaryRoutes(routes []zerotier.Route) (ipv4 *zerotier.Route, ipv6 *zerotier.Route) {
	for _, route := range routes {
		if strings.TrimSpace(route.Via) != "" {
			continue
		}

		_, cidr, err := net.ParseCIDR(strings.TrimSpace(route.Target))
		if err != nil || cidr == nil {
			continue
		}

		routeCopy := route
		if cidr.IP.To4() != nil && ipv4 == nil {
			ipv4 = &routeCopy
			continue
		}
		if cidr.IP.To4() == nil && ipv6 == nil {
			ipv6 = &routeCopy
		}
	}

	return ipv4, ipv6
}

func ExtractManagedRoutes(routes []zerotier.Route) []zerotier.Route {
	primaryIPv4, primaryIPv6 := ExtractPrimaryRoutes(routes)

	managedRoutes := make([]zerotier.Route, 0, len(routes))
	for _, route := range routes {
		switch {
		case primaryIPv4 != nil && route.Target == primaryIPv4.Target && strings.TrimSpace(route.Via) == "":
			continue
		case primaryIPv6 != nil && route.Target == primaryIPv6.Target && strings.TrimSpace(route.Via) == "":
			continue
		default:
			managedRoutes = append(managedRoutes, route)
		}
	}

	return managedRoutes
}

func BuildManagedRoutes(primaryIPv4, primaryIPv6 *zerotier.Route, managedRoutes []zerotier.Route) []zerotier.Route {
	routes := make([]zerotier.Route, 0, len(managedRoutes)+2)
	if primaryIPv4 != nil {
		routes = append(routes, *primaryIPv4)
	}
	if primaryIPv6 != nil {
		routes = append(routes, *primaryIPv6)
	}
	routes = append(routes, cloneRoutes(managedRoutes)...)
	return routes
}

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
