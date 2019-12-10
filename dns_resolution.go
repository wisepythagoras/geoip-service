package main

import (
    "net"
    "context"
)

// GoogleDNSDialer: Get the Google DNS resolver.
func GoogleDNSDialer(ctx context.Context, network, address string) (net.Conn, error) {
    d := net.Dialer{}
    return d.DialContext(ctx, "udp", "8.8.8.8:53")
}

// CloudflareDNSDialer: Get the Cloudflare DNS resolver.
func CloudflareDNSDialer(ctx context.Context, network, address string) (net.Conn, error) {
    d := net.Dialer{}
    return d.DialContext(ctx, "udp", "1.1.1.1:53")
}

// OpenDNSDialer: Get the OpenDNS DNS resolver.
func OpenDNSDialer(ctx context.Context, network, address string) (net.Conn, error) {
    d := net.Dialer{}
    return d.DialContext(ctx, "udp", "208.67.222.222:53")
}

// DynDNSDialer: Get the DynDNS DNS resolver.
func DynDNSDialer(ctx context.Context, network, address string) (net.Conn, error) {
    d := net.Dialer{}
    return d.DialContext(ctx, "udp", "216.146.35.35:53")
}

// OnDSLDNSDialer: Get the OnDSL DNS resolver.
// From https://public-dns.info/nameserver/gr.html
func OnDSLDNSDialer(ctx context.Context, network, address string) (net.Conn, error) {
    d := net.Dialer{}
    return d.DialContext(ctx, "udp", "87.203.2.129:53")
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

// DNSLookup: Make a DNS lookup.
func DNSLookup(domain string) ([]net.IPAddr, error) {
    r := net.Resolver{
        Dial: CloudflareDNSDialer,
    }
    ctx := context.Background()

    // Lookup with Cloudflare's DNS resolver.
    ips, err := r.LookupIPAddr(ctx, domain)

    if err != nil {
        return nil, err
    }

    r = net.Resolver{
        Dial: GoogleDNSDialer,
    }

    // Lookup with Google's DNS resolver.
    ips1, err := r.LookupIPAddr(ctx, domain)

    if err != nil {
        return ips, err
    }

    // Merge the two results.
    ips = MergeIPArrays(ips, ips1)

    r = net.Resolver{
        Dial: OpenDNSDialer,
    }

    // Lookup with OpenDNS's DNS resolver.
    ips1, err = r.LookupIPAddr(ctx, domain)

    // Merge the results again.
    ips = MergeIPArrays(ips, ips1)

    // Merge the two results.
    ips = MergeIPArrays(ips, ips1)

    r = net.Resolver{
        Dial: DynDNSDialer,
    }

    // Lookup with DynDNS's DNS resolver.
    ips1, err = r.LookupIPAddr(ctx, domain)

    // Merge the results again.
    ips = MergeIPArrays(ips, ips1)

    /*
    r = net.Resolver{
        Dial: OnDSLDNSDialer,
    }

    // Lookup with OnDSL's DNS resolver.
    ips1, err = r.LookupIPAddr(ctx, domain)

    // Merge the results again.
    ips = MergeIPArrays(ips, ips1)
    */

    return ips, nil
}

