package utils

import (
	"net/url"
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
	if strings.Contains(urlStr, "/") {
		u, err := url.Parse(urlStr)
		if err != nil {
			return ""
		}
		urlStr = u.Hostname()
	}

	host := strings.ToLower(urlStr)
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

// UrlMatchesPathPattern checks if rawUrl's path starts with pattern.
// If targetHost is supplied, absolute URLs with a non-matching host are rejected.
func UrlMatchesPathPattern(rawUrl, pattern string, targetHost ...string) bool {
	u, err := url.Parse(rawUrl)
	if err != nil {
		return false
	}
	expectedHost := ""
	if len(targetHost) > 0 {
		expectedHost = targetHost[0]
	}
	patU, _ := url.Parse(pattern)
	if expectedHost == "" && patU != nil {
		expectedHost = patU.Host
	}
	if expectedHost != "" && u.Host != "" && Hostname(rawUrl) != Hostname(expectedHost) {
		return false
	}

	pathPattern := pattern
	if patU != nil && patU.Path != "" {
		pathPattern = patU.Path
	}
	return strings.HasPrefix(u.Path, pathPattern)
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
