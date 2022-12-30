package main

import (
	"context"
	"net"
)

var defaultDNSServers = []string{
	"1.1.1.1:53",
	"8.8.8.8:53",
	"208.67.222.222:53",
}

// MergeIPArrays: Merge two IP arrays.
func MergeIPArrays(a []net.IPAddr, b []net.IPAddr) []net.IPAddr {
	if len(a) == 0 {
		return b
	}

	for i := 0; i < len(b); i++ {
		found := false

		for j := 0; j < len(a); j++ {
			if a[j].String() == b[i].String() {
				found = true
				break
			}
		}

		if found {
			continue
		}

		a = append(a, b[i])
	}

	return a
}

// CreateDialer creates a dialer callback function for use with DNSLookup.
func CreateDialer(dnsServer string) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		d := net.Dialer{}
		return d.DialContext(ctx, "udp", dnsServer)
	}
}

// DNSLookup queries the specified DNS servers (or the default ones).
func DNSLookup(domain string, dnsServers []string) ([]net.IPAddr, error) {
	ipAddresses := []net.IPAddr{}

	if len(dnsServers) == 0 {
		dnsServers = defaultDNSServers
	}

	for _, dnsServer := range dnsServers {
		r := net.Resolver{
			Dial: CreateDialer(dnsServer),
		}
		ctx := context.Background()
		ips, err := r.LookupIPAddr(ctx, domain)

		if err != nil {
			return nil, err
		}

		ipAddresses = MergeIPArrays(ipAddresses, ips)
	}

	return ipAddresses, nil
}
