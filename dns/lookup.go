package dns

import (
	"context"
	"net"

	"github.com/miekg/dns"
)

var DefaultDNSServers = []string{
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
		dnsServers = DefaultDNSServers
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

type DNSCaller func(string, []string) ([]net.IP, error)

func DNSALookup(domain string, dnsServers []string) ([]net.IP, error) {
	client := new(dns.Client)

	if len(dnsServers) == 0 {
		dnsServers = DefaultDNSServers
	}

	ipAddrMap := make(map[string]net.IP)
	ipAddresses := []net.IP{}

	for _, dnsServer := range dnsServers {
		msg := new(dns.Msg)
		msg.SetQuestion(dns.Fqdn(domain), dns.TypeA)
		msg.RecursionDesired = true

		r, _, err := client.Exchange(msg, dnsServer)

		if err != nil {
			return nil, err
		}

		for _, answer := range r.Answer {
			if a, ok := answer.(*dns.A); ok {
				ipAddrMap[a.A.String()] = a.A
			}
		}
	}

	for _, ip := range ipAddrMap {
		ipAddresses = append(ipAddresses, ip)
	}

	return ipAddresses, nil
}

func DNSAAAALookup(domain string, dnsServers []string) ([]net.IP, error) {
	client := new(dns.Client)

	if len(dnsServers) == 0 {
		dnsServers = DefaultDNSServers
	}

	ipAddrMap := make(map[string]net.IP)
	ipAddresses := []net.IP{}

	for _, dnsServer := range dnsServers {
		msgA4 := new(dns.Msg)
		msgA4.SetQuestion(dns.Fqdn(domain), dns.TypeAAAA)
		msgA4.RecursionDesired = true

		rA4, _, err := client.Exchange(msgA4, dnsServer)

		if err != nil {
			return nil, err
		}

		for _, answer := range rA4.Answer {
			if a, ok := answer.(*dns.AAAA); ok {
				ipAddrMap[a.AAAA.String()] = a.AAAA
			}
		}
	}

	for _, ip := range ipAddrMap {
		ipAddresses = append(ipAddresses, ip)
	}

	return ipAddresses, nil
}
