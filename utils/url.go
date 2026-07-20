package utils

import (
	"context"
	"net"
	"net/url"
	"slices"
	"strings"

	"golang.org/x/net/publicsuffix"
)

func IsAbsolute(urlStr string) bool {
	u, err := url.Parse(strings.TrimSpace(urlStr))
	return err == nil && u.IsAbs()
}

func ToAbsoluteUrl(base *url.URL, urlStr string) string {
	urlStr = strings.TrimSpace(urlStr)
	if urlStr == "" || base == nil {
		return ""
	}
	u, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}
	if u.IsAbs() {
		return urlStr
	}
	return base.ResolveReference(u).String()
}

func RemoveTrailingSlash(urlStr string) string {
	return strings.TrimSuffix(urlStr, "/")
}

func BaseUrl(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}
	u.Path = ""
	u.RawQuery = ""
	u.Fragment = ""
	return u.String()
}

func Hostname(urlStr string) string {
	if !strings.Contains(urlStr, "/") {
		return urlStr
	}
	u, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}

	host := strings.ToLower(u.Hostname())
	host = strings.TrimPrefix(host, "www.")
	return host
}

func DomainZone(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}
	parts := strings.Split(strings.ToLower(u.Hostname()), ".")
	if len(parts) < 2 {
		return ""
	}
	return parts[len(parts)-1]
}

func UrlMatchesPathPattern(rawUrl, pattern string) bool {
	u, err := url.Parse(rawUrl)
	if err != nil {
		return false
	}
	return strings.HasPrefix(u.Path, pattern)
}

func HostAlias(urlStr string) string {
	alias := Hostname(urlStr)

	// remove public suffix (e.g. ".com", ".co.uk")
	suffix, _ := publicsuffix.PublicSuffix(alias)
	alias = strings.TrimSuffix(alias, "."+suffix)

	// remove common API subdomain prefix
	alias = strings.TrimPrefix(alias, "api.")

	// replace dots with underscores for map key safety
	alias = strings.ReplaceAll(alias, ".", "_")
	return alias
}

// contextKey is unexported so values set with it can't collide with keys
// defined by other packages using the same underlying string.
type contextKey string

// AllowLoopbackKey when set to true on a context, disables the private/loopback
// check in IsPrivateOrLoopbackHost. Intended for tests that spin up a
// httptest server, which is only reachable via a loopback address.
const AllowLoopbackKey contextKey = "allow_loopback"

// IsPrivateOrLoopbackHost returns true if the host is a private, loopback, link-local, unspecified, or local-only IP/domain.
func IsPrivateOrLoopbackHost(ctx context.Context, host string) bool {
	if ctx != nil {
		if allow, ok := ctx.Value(AllowLoopbackKey).(bool); ok && allow {
			return false
		}
	}
	host = strings.TrimSpace(strings.ToLower(host))
	if host == "" {
		return true
	}

	// 1. If it's a raw IP address, parse and check it directly.
	if ip := net.ParseIP(host); ip != nil {
		return isIPPrivateOrLoopback(ip)
	}

	// 2. Check for common local hostnames (e.g. localhost, local)
	if host == "localhost" || strings.HasSuffix(host, ".local") || strings.HasSuffix(host, ".internal") {
		return true
	}

	// 3. Resolve DNS to get all IP addresses and check them.
	if ctx == nil {
		ctx = context.Background()
	}
	ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
	if err != nil {
		// If DNS resolution fails, we cannot connect to it, but to prevent SSRF
		// bypasses or weird errors, we treat it as unsafe.
		return true
	}

	return slices.ContainsFunc(ips, isIPPrivateOrLoopback)
}

func isIPPrivateOrLoopback(ip net.IP) bool {
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified()
}
