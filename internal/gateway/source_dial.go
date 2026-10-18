package gateway

import (
	"context"
	"crypto/rand"
	"fmt"
	"net"
	"time"
)

// sourceDialContext binds outbound TCP to a local IPv6 address and refuses
// IPv4 destinations. A dual-stack dial with an IPv6 LocalAddr fails the IPv4
// candidate with "invalid argument" and can lose the race against the AAAA
// record, so the lookup here is IPv6 only.
func sourceDialContext(ip net.IP, timeout time.Duration) func(ctx context.Context, network, addr string) (net.Conn, error) {
	dialer := &net.Dialer{
		Timeout:   timeout,
		KeepAlive: 30 * time.Second,
		LocalAddr: &net.TCPAddr{IP: ip},
	}
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}
		targets, err := ipv6Targets(ctx, host)
		if err != nil {
			return nil, err
		}
		var last error
		for _, target := range targets {
			conn, err := dialer.DialContext(ctx, "tcp6", net.JoinHostPort(target.String(), port))
			if err == nil {
				return conn, nil
			}
			last = err
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
		}
		if last == nil {
			last = fmt.Errorf("no ipv6 address for %s", host)
		}
		return nil, last
	}
}

// prefixDialContext binds every new TCP connection to a different random
// address inside prefix. Callers disable connection reuse so one upstream
// request cannot stay glued to a single source.
func prefixDialContext(prefix *net.IPNet, timeout time.Duration) func(ctx context.Context, network, addr string) (net.Conn, error) {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		dialer := &net.Dialer{
			Timeout:   timeout,
			KeepAlive: 30 * time.Second,
			LocalAddr: &net.TCPAddr{IP: randomInPrefix(prefix)},
		}
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}
		targets, err := ipv6Targets(ctx, host)
		if err != nil {
			return nil, err
		}
		var last error
		for _, target := range targets {
			conn, err := dialer.DialContext(ctx, "tcp6", net.JoinHostPort(target.String(), port))
			if err == nil {
				return conn, nil
			}
			last = err
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
		}
		if last == nil {
			last = fmt.Errorf("no ipv6 address for %s", host)
		}
		return nil, last
	}
}

// randomInPrefix returns a uniformly random address covered by prefix.
// Network bits are taken from the prefix; the host bits come from crypto/rand.
func randomInPrefix(prefix *net.IPNet) net.IP {
	ip := prefix.IP.Mask(prefix.Mask).To16()
	out := make(net.IP, net.IPv6len)
	copy(out, ip)
	var rnd [net.IPv6len]byte
	if _, err := rand.Read(rnd[:]); err != nil {
		return out
	}
	ones, bits := prefix.Mask.Size()
	for bit := ones; bit < bits; bit++ {
		if rnd[bit/8]&(1<<uint(7-bit%8)) != 0 {
			out[bit/8] |= 1 << uint(7-bit%8)
		}
	}
	return out
}

func ipv6Targets(ctx context.Context, host string) ([]net.IP, error) {
	if ip := net.ParseIP(host); ip != nil {
		if ip.To4() != nil {
			return nil, fmt.Errorf("source ipv6 cannot dial ipv4 host %s", host)
		}
		return []net.IP{ip}, nil
	}
	ips, err := net.DefaultResolver.LookupIP(ctx, "ip6", host)
	if err != nil {
		return nil, err
	}
	out := make([]net.IP, 0, len(ips))
	for _, ip := range ips {
		if ip.To4() == nil {
			out = append(out, ip)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no ipv6 address for %s", host)
	}
	return out, nil
}
