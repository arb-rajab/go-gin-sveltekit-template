// Package ssrf guards outbound HTTP requests whose destination a user (or an
// attacker who has stolen a user's session) controls, such as webhook URLs.
// Left unvalidated, such a request can be pointed at anything reachable from
// the backend's own network position: an internal admin panel, another
// service in a shared VPC, or a cloud provider's instance-metadata endpoint
// (169.254.169.254 on AWS/GCP/Azure), which typically needs no credential and
// can hand back real cloud credentials.
//
// The guard runs at two points, deliberately:
//
//   - ValidateURL, at configuration time: a same-request rejection with a
//     useful 422, so a private-looking destination is refused immediately
//     rather than silently failing later.
//   - GuardedDialContext (via NewGuardedClient), at the moment of the real
//     TCP dial on every attempt: the only check that matters for DNS
//     rebinding, since a hostname that resolved to a public address at
//     configuration time can be repointed at a private one before the
//     request is made.
package ssrf

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
)

// ErrDestinationBlocked is returned (wrapped) whenever a destination is
// refused because it names, or resolves to, a private, loopback, link-local,
// or otherwise disallowed address.
var ErrDestinationBlocked = errors.New("outbound destination resolves to a private, loopback, link-local, or otherwise disallowed address")

// NewGuardedClient returns an http.Client whose every connection goes through
// GuardedDialContext. Redirects are not followed: a redirect target would be
// a second, unvalidated destination.
func NewGuardedClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			DialContext: GuardedDialContext(&net.Dialer{Timeout: 10 * time.Second}),
		},
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
}

// ValidateURL checks destination well-formedness and, best-effort,
// whether it already names a disallowed address. It is not the real
// enforcement point (GuardedDialContext is) — a hostname that resolves fine
// right now can still be blocked at dispatch time, and a hostname that
// fails to resolve right now (a typo, a host that doesn't exist yet, or a
// network with no outbound DNS at all) is not
// itself treated as a reason to refuse configuring the channel, since
// GuardedDialContext will enforce the real check when it actually matters.
func ValidateURL(ctx context.Context, destination string) error {
	u, err := url.Parse(destination)
	if err != nil {
		return fmt.Errorf("%w: not a valid URL", ErrDestinationBlocked)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("%w: scheme must be http or https", ErrDestinationBlocked)
	}
	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("%w: missing host", ErrDestinationBlocked)
	}

	if ip := net.ParseIP(host); ip != nil {
		if blockedIP(ip) {
			return fmt.Errorf("%w: %s", ErrDestinationBlocked, ip)
		}
		return nil
	}

	lookupCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	addrs, err := net.DefaultResolver.LookupIPAddr(lookupCtx, host)
	if err != nil {
		// Not itself a rejection reason — see the doc comment above.
		return nil
	}
	for _, addr := range addrs {
		if blockedIP(addr.IP) {
			return fmt.Errorf("%w: %s resolves to %s", ErrDestinationBlocked, host, addr.IP)
		}
	}
	return nil
}

// blockedIP is the actual range list: loopback (127.0.0.0/8, ::1),
// link-local (169.254.0.0/16, fe80::/10 — the range that includes every
// cloud provider's instance-metadata endpoint, 169.254.169.254 named
// explicitly since it's the specific, high-value target this exists to
// stop), RFC 1918/RFC 4193 private ranges, the unspecified address, and
// multicast. Nothing here is scoped to a particular cloud provider — the
// range is the control, not a hardcoded metadata IP.
func blockedIP(ip net.IP) bool {
	return ip.IsLoopback() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsPrivate() ||
		ip.IsUnspecified() ||
		ip.IsMulticast()
}

// GuardedDialContext wraps a *net.Dialer so it re-resolves and re-validates
// the target host on every connection attempt (including retries), then dials
// the validated IP address directly rather than handing the hostname back to
// the standard library dialer for a second, independent resolution. That
// second point matters as much as the re-validation: without it, an
// attacker's DNS server could answer this function's lookup with a safe
// address and the dialer's own lookup with a private one, and the request
// would still reach the private target.
func GuardedDialContext(dialer *net.Dialer) func(ctx context.Context, network, addr string) (net.Conn, error) {
	if dialer == nil {
		dialer = &net.Dialer{Timeout: 10 * time.Second}
	}
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, fmt.Errorf("outbound dial target %q: %w", addr, err)
		}

		var candidates []net.IP
		if ip := net.ParseIP(host); ip != nil {
			candidates = []net.IP{ip}
		} else {
			addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
			if err != nil {
				return nil, fmt.Errorf("resolve outbound host %q: %w", host, err)
			}
			for _, a := range addrs {
				candidates = append(candidates, a.IP)
			}
		}
		if len(candidates) == 0 {
			return nil, fmt.Errorf("resolve outbound host %q: no addresses found", host)
		}
		for _, ip := range candidates {
			if blockedIP(ip) {
				return nil, fmt.Errorf("%w: %s", ErrDestinationBlocked, ip)
			}
		}
		return dialer.DialContext(ctx, network, net.JoinHostPort(candidates[0].String(), port))
	}
}
