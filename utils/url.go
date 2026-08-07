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
